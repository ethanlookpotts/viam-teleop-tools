package main

import (
	"fmt"

	"context"
	"encoding/json"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	datapb "go.viam.com/api/app/data/v1"
	"go.viam.com/rdk/logging"
	"go.viam.com/utils/rpc"
)

type DataSource struct {
	AppUrl         string `json:"app_url"`
	OrganizationID string `json:"organization_id"`
	PartID         string `json:"part_id"`
	APIKeyID       string `json:"api_key_id"`
	APIKeyValue    string `json:"api_key_value"`
}

type DataDestination struct {
	MongoDBURL     string `json:"mongodb_url"`
	OrganizationID string `json:"organization_id"`
	LocationID     string `json:"location_id"`
	MachineID      string `json:"machine_id"`
	PartID         string `json:"part_id"`
}

type Config struct {
	Source        DataSource      `json:"source"`
	Destination   DataDestination `json:"destination"`
	SyncBackNDays float64         `json:"sync_back_n_days"`
}

const (
	QueryableTabularDatabaseName   = "sensorData"
	QueryableTabularCollectionName = "readings"
)

func getConfig(args []string) (*Config, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected path to config file, got %d", len(args))
	}

	configFile, err := os.Open(args[1])
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	return &config, nil
}

func main() {
	logger := logging.NewDebugLogger("sync-data")
	ctx := context.Background()
	config, err := getConfig(os.Args)
	if err != nil {
		logger.Fatalw("error getting config", "error", err)
	}
	logger.Infow("syncing data", "config", config)

	conn, err := dialApp(ctx, config.Source)
	if err != nil {
		logger.Fatalw("error dialing app", "error", err)
	}
	dataClient := datapb.NewDataServiceClient(conn)
	matchStage, err := getMatchStage(config.Source, config.SyncBackNDays)
	if err != nil {
		logger.Fatalw("", "error", err)
	}
	tabularDataByMQLResponse, err := dataClient.TabularDataByMQL(context.Background(), &datapb.TabularDataByMQLRequest{
		OrganizationId: config.Source.OrganizationID,
		MqlBinary:      [][]byte{matchStage},
	})
	if err != nil {
		logger.Fatalw("error querying source data", "error", err)
	}

	if len(tabularDataByMQLResponse.Data) == 0 {
		logger.Fatal("Zero documents returned from TabularDataByMQL")
	}

	logger.Infof("found %d documents", len(tabularDataByMQLResponse.RawData))

	unmarshalledData := []map[string]any{}
	for _, data := range tabularDataByMQLResponse.RawData {
		tabularDataMap, err := unmarshallRawData[map[string]any](data)
		if err != nil {
			logger.Fatalw("error unmarshaling tabular data", "error", err)
		}
		// TODO: This needs to be typed. This is very brittle.
		tabularDataMap["organization_id"] = config.Destination.OrganizationID
		tabularDataMap["robot_id"] = config.Destination.MachineID
		tabularDataMap["location_id"] = config.Destination.LocationID
		tabularDataMap["part_id"] = config.Destination.PartID

		unmarshalledData = append(unmarshalledData, tabularDataMap)
		tabularDataJson, err := json.MarshalIndent(tabularDataMap, "", "    ")
		if err != nil {
			logger.Fatalw("error marshaling tabular data", "error", err)
		}
		logger.Info(string(tabularDataJson))
	}
	err = reuploadData(ctx, logger, config.Destination, unmarshalledData)
	if err != nil {
		logger.Fatalw("error reuploading data", "error", err)
	}
	logger.Info("Sync complete")
	// TODO: Add verify step that creates a data client to the local app and ensures that the new data
	// matches the old data (-org-id, machine-id, etc)
}

func dialApp(ctx context.Context, source DataSource) (rpc.ClientConn, error) {
	appURL, err := url.Parse(source.AppUrl)
	if err != nil {
		return nil, err
	}

	conn, err := rpc.DialDirectGRPC(
		ctx,
		appURL.Host,
		nil,
		rpc.WithEntityCredentials(source.APIKeyID,
			rpc.Credentials{
				Type:    rpc.CredentialsTypeAPIKey,
				Payload: source.APIKeyValue,
			},
		),
	)

	return conn, err
}

func reuploadData(ctx context.Context, logger logging.Logger, destination DataDestination, data []map[string]any) error {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(destination.MongoDBURL))
	if err != nil {
		return errors.Wrap(err, "failed to connect to local mongo")
	}
	defer mongoClient.Disconnect(ctx)
	coll := mongoClient.Database(QueryableTabularDatabaseName).Collection(QueryableTabularCollectionName)
	// My brutal index to speed up re-checking
	// With this index, this func takes ~10 seconds locally. Without this, this func can take >10 minutes
	indexName := "viam-teleop-tools-time-received-index"

	logger.Infof("Creating %s.", indexName)
	_, err = coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "time_received", Value: 1}},
		Options: &options.IndexOptions{
			Name: &indexName,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create index")
	}
	logger.Infof("index created")

	numInserted := 0

	// pretty inefficient but totally fine for localdb.
	for i, datum := range data {
		res := coll.FindOne(ctx, bson.M{"time_received": datum["time_received"]})
		err := res.Err()
		isNoDocs := errors.Is(err, mongo.ErrNoDocuments)
		if err != nil && !isNoDocs {
			return errors.Wrap(err, "failed to find matching record")
		}
		// insert
		if isNoDocs {
			_, err = coll.InsertOne(ctx, datum)
			if err != nil {
				return errors.Wrap(err, "failed to insert data")
			}
			numInserted += 1
		}
		if i%100 == 0 || i == len(data)-1 {
			logger.Infof("Upload: %d/%d. Newly inserted points: %d", i+1, len(data), numInserted)
		}

	}

	return nil
}

func unmarshallRawData[T any](data []byte) (T, error) {
	var tabularDataMap T
	err := bson.Unmarshal(data, &tabularDataMap)
	if err != nil {
		return tabularDataMap, err
	}
	return tabularDataMap, nil
}

func getMatchStage(source DataSource, syncBackNDays float64) ([]byte, error) {
	matchStage := bson.D{
		bson.E{
			Key: "$match",
			Value: bson.M{
				"part_id": source.PartID,
				"time_received": bson.M{
					"$gte": primitive.NewDateTimeFromTime(
						time.Now().Add(time.Duration(-1 * syncBackNDays * float64(time.Hour) * 24))),
				},
			},
		},
	}
	matchStageBytes, err := bson.Marshal(matchStage)
	if err != nil {
		return nil, fmt.Errorf("error getting match stage: %w", err)
	}

	return matchStageBytes, nil
}
