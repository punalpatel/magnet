package mock

import (
	"context"

	"github.com/pivotalservices/magnet"
)

type IaaS struct {
	StateFn    func(ctx context.Context) (*magnet.State, error)
	ConvergeFn func(ctx context.Context, state *magnet.State) error
}

func (m *IaaS) State(ctx context.Context) (*magnet.State, error) {
	if m.StateFn != nil {
		return m.StateFn(ctx)
	}
	return nil, nil
}

func (m *IaaS) Converge(ctx context.Context, state *magnet.State) error {
	if m.ConvergeFn != nil {
		return m.ConvergeFn(ctx, state)
	}
	return nil
}
