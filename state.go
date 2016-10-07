package magnet

import (
	"context"
	"io"
)

func Check(ctx context.Context, i IaaS) error {
	s, err := i.State(ctx)
	_ = s
	if err != nil {
		// Need to log this
		return err
	}
	//
	// if !IsBalanced(s) {
	// 	s.PrintJobs(os.Stdout)
	// 	expected := Balance(s)
	// 	s.PrintDelta(expected)
	// 	// Has anything actually changed?
	// 	i.Converge(ctx, expected)
	// 	s, err = i.State(ctx)
	// 	if err != nil {
	// 		// Need to log this
	// 		return err
	// 	}
	// 	if !IsBalanced(s) {
	// 		s.PrintJobs(os.Stdout)
	// 	}
	// }

	return nil
}

func IsBalanced(s *State) bool {
	return false
}

func Balance(s *State) *State {
	return nil
}

func (s *State) PrintJobs(w io.Writer) {

}

func (s *State) PrintDelta(s1 *State) {

}
