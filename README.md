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

```text
Usage:
  -f string
        Routes file (validation config read from file too)
  -npt
        Verify NaPTAN platform tags
  -r int
        Relation ID
```

## AWS application

Requires AWS CDK to be installed

```shell
make deploy
```

Looks for `.json` files in `s3://<bucketName>/routes/**.json`

See [routefile.schema.json](schema/routefile.schema.json) for the JSON-schema or [routes](routes) for example files.

## Tasks

[xcfile.dev](https://xcfile.dev) tasks

### build-cmd

requires: clean

```shell
go run ./scripts/build-cmd --zip
```

### clean

```shell
rm -rf build/* || true
```

### format

directory: stack

```shell
terraform fmt --recursive
```

### init

directory: stack
environment: AWS_PROFILE=osmptv

```shell
terraform init -backend-config=tfvars/backend.tfvars
```

### plan

requires: upload-cmd
directory: stack
environment: AWS_PROFILE=osmptv

```shell
terraform plan -var-file=tfvars/pro.tfvars
```

### apply

requires: upload-cmd
directory: stack
environment: AWS_PROFILE=osmptv

```shell
terraform apply -var-file=tfvars/pro.tfvars -auto-approve
```

### upload-cmd

requires: build-cmd
environment: AWS_PROFILE=osmptv

```shell
#!/bin/bash
source stack/tfvars/.env
BINARY_BUCKET=$BINARY_BUCKET go run ./scripts/upload-binaries
```
