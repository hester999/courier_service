package db

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type TRMRunner struct {
	manager *manager.Manager
}

func NewTRMRunner(m *manager.Manager) *TRMRunner {
	return &TRMRunner{manager: m}
}

func (r *TRMRunner) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.manager.Do(ctx, fn)
}
