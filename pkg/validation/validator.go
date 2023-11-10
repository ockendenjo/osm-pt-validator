package validation

import "github.com/ockendenjo/osm-pt-validator/pkg/osm"

func DefaultValidator(client *osm.OSMClient) *Validator {
	return &Validator{config: DefaultConfig(), osmClient: client}
}

func NewValidator(config Config, client *osm.OSMClient) *Validator {
	return &Validator{config: config, osmClient: client}
}

type Validator struct {
	config    Config
	osmClient *osm.OSMClient
}

func (v *Validator) GetConfig() Config {
	return v.config
}
