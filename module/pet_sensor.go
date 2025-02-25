package main

// pet_sensor is a rdk:component:sensor component that returns a picture of Zack's cat (Ashley)
// in the GetReadings return value. If future authors want to add (small <1MB)  pictures
// of their pets, go ahead!

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	_ "embed"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

//go:embed pet_sensor_images/ashley.jpg
var ashleyImg []byte

var nameToImage = map[string][]byte{
	"ashley": ashleyImg,
}

type Pet struct {
	resource.Named
	resource.TriviallyCloseable
	resource.TriviallyReconfigurable
	cfg *PetConfig
}

type PetConfig struct {
	PetName string `json:"pet_name"`
}

func (c *PetConfig) Validate(path string) ([]string, error) {
	if c.PetName == "" {
		c.PetName = "ashley"
	}
	if _, ok := nameToImage[c.PetName]; !ok {
		return nil, fmt.Errorf("pet name %s not found. Valid names are: %v", c.PetName, nameToImage)
	}
	return nil, nil
}

func newPet(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	cfg, err := resource.NativeConfig[*PetConfig](conf)
	if err != nil {
		return nil, err
	}
	if cfg.PetName == "" {
		cfg.PetName = "ashley"
	}
	return &Pet{
		Named: conf.ResourceName().AsNamed(),
		cfg:   cfg,
	}, nil
}

func (p *Pet) getImg() string {
	img := nameToImage[p.cfg.PetName]
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(img))
}

func (p *Pet) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	return map[string]any{
		"current_time": time.Now().Format(time.RFC3339),
		"pet_name":     p.cfg.PetName,
		"pet_image":    p.getImg(),
	}, nil
}

func (p *Pet) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return map[string]any{"pet_name": p.cfg.PetName, "pet_image": p.getImg()}, nil
}
