package get_orders

import (
	"context"
	"time"

	"avito/pkg"
)

type PollingOrder struct {
	poller poller

	errChan       chan error
	periodPooling int
	log           pkg.Logger
}

func NewPollingOrder(p poller, periodPooling int, log pkg.Logger) *PollingOrder {
	return &PollingOrder{
		periodPooling: periodPooling,
		poller:        p,
		errChan:       make(chan error, 10),
		log:           log,
	}
}

func (p *PollingOrder) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.periodPooling) * time.Second)
	defer ticker.Stop()
	defer close(p.errChan)
	defer p.poller.CloseGateWay()

	for {
		select {
		case <-ctx.Done():
			p.log.Info("poller stopped")
			return
		case <-ticker.C:
			p.log.Info("polling orders")
			err := p.poller.AssignCouriersOnOrdersFromGateWay(ctx)
			if err != nil {
				p.errChan <- err
			}
		}
	}
}

func (p *PollingOrder) GetErrChan() <-chan error {
	return p.errChan
}
