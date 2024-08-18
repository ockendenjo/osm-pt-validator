package validation

import (
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

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

type ValidationError struct {
	URL     string `json:"url,omitempty"`
	Message string `json:"message"`
}

func (v ValidationError) String() string {
	return fmt.Sprintf("%s - %s", v.Message, v.URL)
}
