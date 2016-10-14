package mock

import (
	"context"

	"github.com/pivotalservices/magnet"
)

// IaaS is a mock IaaS whose State and Converge functions
// can be replaced.
type IaaS struct {
	StateFn    func(ctx context.Context) (*magnet.State, error)
	ConvergeFn func(ctx context.Context, state *magnet.State, rec *magnet.RuleRecommendation) error
}

// State runs the IaaS's supplied StateFn.
// If no state function was provided, it returns a non-nil state
// and a nil error.
func (m *IaaS) State(ctx context.Context) (*magnet.State, error) {
	if m.StateFn != nil {
		return m.StateFn(ctx)
	}
	return &magnet.State{}, nil
}

// Converge runs the IaaS's supplied ConvergeFn.
// If no converge function was provided it returns a nil error.
func (m *IaaS) Converge(ctx context.Context, state *magnet.State, rec *magnet.RuleRecommendation) error {
	if m.ConvergeFn != nil {
		return m.ConvergeFn(ctx, state, rec)
	}
	return nil
}
