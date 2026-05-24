package collectors

import "context"

type Collector interface {
	Run(ctx context.Context) error
	Close() error
}
