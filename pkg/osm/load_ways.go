package osm

import (
	"context"
)

func (c *OSMClient) LoadWays(ctx context.Context, wayIds []int64) map[int64]*Way {
	wayMap := map[int64]*Way{}

	for _, wayId := range wayIds {
		wayResult := loadWay(ctx, c, wayId)
		wayMap[wayResult.WayID] = wayResult.Way
	}

	return wayMap
}

func loadWay(ctx context.Context, client *OSMClient, wayId int64) wayResult {
	way, err := client.GetWay(ctx, wayId)
	if err != nil {
		return wayResult{
			WayID: wayId,
			Way:   nil,
		}
	}
	return wayResult{
		WayID: wayId,
		Way:   &way,
	}
}

type wayResult struct {
	WayID int64
	Way   *Way
}
