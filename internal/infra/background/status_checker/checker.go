package status_checker

import (
	"context"
	"errors"
	"sync"
	"time"

	"avito/internal/apperr"
	"avito/pkg"
)

type Checker struct {
	checker checker

	log     pkg.Logger
	delay   int
	errChan chan error // Теперь приватное
}

func NewChecker(checker checker, log pkg.Logger, delay int) *Checker {
	var (
		once        sync.Once
		checkerInst *Checker
	)
	once.Do(func() {
		checkerInst = &Checker{
			checker: checker,
			log:     log,
			delay:   delay,
			errChan: make(chan error, 5),
		}
	})
	return checkerInst
}
func (c *Checker) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(c.delay) * time.Second)
	defer ticker.Stop()
	defer close(c.errChan)

	evictChan := c.checker.GetEvictChan()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("Checker exit")
			return

		case <-ticker.C:
			c.log.Info("Checking status after seconds %v", c.delay)
			err := c.checker.CheckAndSetStatus(ctx)
			if err != nil {
				var appErr *apperr.AppErr
				if errors.As(err, &appErr) {
					if !appErr.Is(apperr.ErrNotFound) {
						c.errChan <- err
					}
				}
			}

		case evicted, ok := <-evictChan:
			if !ok {
				c.log.Info("Evict channel closed")
				return
			}
			err := c.checker.ProcessEvicted(ctx, evicted)
			if err != nil {
				c.log.Error("Failed to process evicted couriers: %v", err)
				c.errChan <- err
			}
		}
	}
}

func (c *Checker) GetErrChan() <-chan error { // Может быть использовано далее, например для записи в файл с логами. Дальше вроде будут воркшопы, по логированию. Буду читать канал ошибок и писать в лог.
	return c.errChan
}
