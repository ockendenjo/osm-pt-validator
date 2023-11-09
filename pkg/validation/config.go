package validation

type Config struct {
	NaptanPlatformTags           bool    `json:"naptanPlatformTags"`
	MinimumNodeMembers           int     `json:"minimumNodeMembers"`
	IgnoreTraversalDirectionWays []int64 `json:"ignoreTraversalDirectionWays"`
	ignoreTraversalMap           map[int64]bool
	MinimumRouteVariants         int `json:"minimumRouteVariants"`
}

func DefaultConfig() Config {
	return Config{NaptanPlatformTags: true}
}

func (c *Config) IsWayDirectionIgnored(wayId int64) bool {
	if c.ignoreTraversalMap == nil {
		c.buildMap()
	}
	value, found := c.ignoreTraversalMap[wayId]
	if found {
		return value
	}
	return false
}

func (c *Config) buildMap() {
	m := map[int64]bool{}
	for _, way := range c.IgnoreTraversalDirectionWays {
		m[way] = true
	}
	c.ignoreTraversalMap = m
}
