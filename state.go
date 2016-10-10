package magnet

import (
	"context"
	"io"
	"log"
	"math"
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

type hostList []string

func (h hostList) exceedsMax(hostCount int) bool {
	totalJobs := len(h)
	maxJobsPerHost := int(math.Ceil(float64(totalJobs) / float64(hostCount)))

	counts := make(map[string]int)
	for _, host := range h {
		counts[host] = counts[host] + 1
	}

	for _, v := range counts {
		if v > maxJobsPerHost {
			return true
		}
	}
	return false
}

// IsBalanced determines whether the state of a deployment is balanced.
// A deployment is balanced jobs are spread across as many hosts as possible.
func IsBalanced(s *State) bool {
	hostCount := len(s.Hosts)

	jobHosts := make(map[string]hostList)
	for _, vm := range s.VMs {
		jobHosts[vm.Job] = append(jobHosts[vm.Job], vm.Host)
	}

	for jobName, hosts := range jobHosts {
		if hosts.exceedsMax(hostCount) {
			log.Printf("job %s is not balanced", jobName)
			return false
		}
	}
	return true
}

func Balance(s *State) *State {
	return nil
}

func (s *State) PrintJobs(w io.Writer) {

}

func (s *State) PrintDelta(s1 *State) {

}
