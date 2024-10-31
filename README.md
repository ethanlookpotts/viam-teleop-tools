# viam-teleop-tools

Setup & build the module:

```
just setup
just build
```

## Provided Components

- `viz-team:teleop:everything`: A sensor that returns every interesting object that we want to support.
- `viz-team:teleop:globetrotter`: A movement sensor that travels around the world.

## Sync data from the cloud to your local db

The sync-data go package creates a simple CLI that reads from a file (canonically, in ./sync-configs/your-new-filename-here.json (will be ignored by git)). It then:

- connects to the remote viam instance specified in the src
- runs TabularDataByMQL on the provided robot for the specified timeframe
- Pulls all the new datapoints into your local app instance as if they were emitted by the `destination` robot
- (it does not duplicate points if you run this multiple times.)

Ex:

```
just sync ./sync-configs/fleet-rover-02.json
```
