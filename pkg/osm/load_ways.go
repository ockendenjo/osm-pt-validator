package osm

import (
	"context"
)

func (c *OSMClient) LoadWays(ctx context.Context, wayIds []int64) map[int64]*Way {
	wayMap := map[int64]*Way{}

	for _, wayId := range wayIds {
		wayObj := loadWay(ctx, c, wayId)
		wayMap[wayId] = wayObj
	}

	return wayMap
}

func loadWay(ctx context.Context, client *OSMClient, wayId int64) *Way {
	way, err := client.GetWay(ctx, wayId)
	if err != nil {
		return nil
	}
	return &way
}
