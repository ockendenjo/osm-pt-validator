package validation

type Config struct {
	NaptanPlatformTags bool `json:"naptanPlatformTags"`
}

func DefaultConfig() Config {
	return Config{NaptanPlatformTags: true}
}
