# OSM public transport validator

Validator for public transport bus routes in OpenStreetMap.

## Usage

```shell
# go run scripts/validate/main.go <relationId>
go run scripts/validate/main.go 103630
```

## Features

* Validates tags on the relation
* Validates that platforms/stops are ordered before ways
* Validates that ways are correctly ordered in a continuous path
* Validates that oneway ways are traversed in the correct direction
* Validates that nodes have expected tags 

## Limitations

* Does not verify if stop nodes are also part of a way
