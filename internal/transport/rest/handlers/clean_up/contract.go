package clean_up

import (
	"context"
)

type Docker interface {
	CleanupPipelinesContainers(ctx context.Context) error
}
