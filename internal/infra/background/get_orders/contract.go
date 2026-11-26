package get_orders

import (
	"context"
)

type poller interface {
	AssignCouriersOnOrdersFromGateWay(ctx context.Context) error
	CloseGateWay()
}
