package cluster

import (
	"context"
	"dashboard/api/internal/scopes/cluster/entities"
	"fmt"
)

func (c *Cluster) DatabasesDetailed(ctx context.Context, filter entities.DatabasesFilter) ([]entities.DatabaseDetails, error) {

	fmt.Printf("%+v\n", filter)

	databases, err := c.storage.DatabasesDetails(ctx, filter)
	if err != nil {
		c.logger.ErrorContext(ctx, "details", "error", err)
		return nil, err
	}

	return databases, nil
}
