package validation

type Config struct {
	NaptanPlatformTags bool `json:"naptanPlatformTags"`
	MinimumNodeMembers int  `json:"minimumNodeMembers"`
}

func DefaultConfig() Config {
	return Config{NaptanPlatformTags: true}
}
