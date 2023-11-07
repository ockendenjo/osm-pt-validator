# OSM public transport validator

Validator for public transport bus routes in OpenStreetMap.

Provided as a Go script (runnable from terminal) and an AWS application for daily verification.

## Features

* Validates tags on the relation
* Validates that platforms/stops are ordered before ways
* Validates that ways are correctly ordered in a continuous path
* Validates that oneway ways are traversed in the correct direction
* Validates that nodes have expected tags
* Validates order of stops, and they are part of the route

## Limitations

* Only for bus routes

## Script

```shell
# go run scripts/validate/main.go [-npt] -r <relationId>
go run scripts/validate/main.go -r 103630
```

## AWS application

Requires AWS CDK to be installed

```shell
make deploy
```

Looks for `.json` files in `s3://<bucketName>/routes/**.json`

See [routefile.schema.json](schema/routefile.schema.json) for the JSON-schema or [routes](routes) for example files.
