package osm

import (
	"context"
)

func (c *OSMClient) LoadWays(ctx context.Context, wayIds []int64) map[int64]*Way {
	ch := make(chan wayResult, len(wayIds))
	wayMap := map[int64]*Way{}

	remaining := 0
	for idx, wayId := range wayIds {
		go loadWay(ctx, c, wayId, ch)
		remaining++
		if idx >= c.parallelReqs {
			//Wait before starting next request
			wayResult := <-ch
			remaining--
			wayMap[wayResult.WayID] = wayResult.Way
		}
	}
	for i := 0; i < remaining; i++ {
		wayResult := <-ch
		wayMap[wayResult.WayID] = wayResult.Way
	}
	return wayMap
}

func loadWay(ctx context.Context, client *OSMClient, wayId int64, c chan wayResult) {
	way, err := client.GetWay(ctx, wayId)
	if err != nil {
		c <- wayResult{
			WayID: wayId,
			Way:   nil,
		}
		return
	}
	c <- wayResult{
		WayID: wayId,
		Way:   &way,
	}
}

type wayResult struct {
	WayID int64
	Way   *Way
}
