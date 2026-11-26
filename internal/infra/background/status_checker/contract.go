package status_checker

import "context"

type checker interface {
	CheckAndSetStatus(ctx context.Context) error
	GetEvictChan() <-chan map[int64]struct{}
	ProcessEvicted(ctx context.Context, evicted map[int64]struct{}) error
}
