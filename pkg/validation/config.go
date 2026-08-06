package validation

type Config struct {
	NaptanPlatformTags   bool          `json:"naptanPlatformTags"`
	MinimumNodeMembers   int           `json:"minimumNodeMembers"`
	MinimumRouteVariants int           `json:"minimumRouteVariants"`
	Ignore               *IgnoreConfig `json:"ignore"`
}

type IgnoreConfig struct {
	Ways  *IgnoreWayConfig   `json:"ways"`
	Nodes *IgnoreNodesConfig `json:"nodes"`
}

type IgnoreWayConfig struct {
	TraversalDirection []int64 `json:"traversalDirection"`
	traversalMap       map[int64]bool
}

type IgnoreNodesConfig struct {
	Any    []int64 `json:"any"`
	anyMap map[int64]bool
}

func DefaultConfig() Config {
	return Config{NaptanPlatformTags: true}
}

func (c *Config) IsWayDirectionIgnored(wayId int64) bool {
	if c.Ignore == nil || c.Ignore.Ways == nil {
		return false
	}
	if c.Ignore.Ways.traversalMap == nil {
		c.buildTraversalMap()
	}
	value, found := c.Ignore.Ways.traversalMap[wayId]
	if found {
		return value
	}
	return false
}

func (c *Config) buildTraversalMap() {
	m := map[int64]bool{}
	for _, way := range c.Ignore.Ways.TraversalDirection {
		m[way] = true
	}
	c.Ignore.Ways.traversalMap = m
}

func (c *Config) IsNodeErrorIgnored(nodeId int64) bool {
	if c.Ignore == nil || c.Ignore.Nodes == nil {
		return false
	}
	if c.Ignore.Nodes.anyMap == nil {
		c.buildNodeMap()
	}
	value, found := c.Ignore.Nodes.anyMap[nodeId]
	if found {
		return value
	}
	return false
}

func (c *Config) buildNodeMap() {
	m := map[int64]bool{}
	for _, node := range c.Ignore.Nodes.Any {
		m[node] = true
	}
	c.Ignore.Nodes.anyMap = m
}
