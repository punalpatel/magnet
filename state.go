package magnet

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/fatih/color"
)

var (
	balancedIndicator   = color.GreenString("✓")
	unbalancedIndicator = color.RedString("✗")
)

// Check gets the state of the deployment on the specified IaaS,
// checks whether is it balanced, and attempts to rebalence
// if necessary.
func Check(ctx context.Context, i IaaS) error {
	s, err := i.State(ctx)
	_ = s
	if err != nil {
		// Need to log this
		return err
	}
	PrintJobs(s, os.Stdout)
	if !IsBalanced(s) {
		newState := Balance(s)
		err = i.Converge(ctx, newState)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsBalanced determines whether the state of a deployment is balanced.
// A deployment is balanced jobs are spread across as many hosts as possible.
func IsBalanced(s *State) bool {
	jobHosts := make(map[string]hostList)
	for _, vm := range s.VMs {
		jobHosts[vm.Job] = append(jobHosts[vm.Job], vm.Host)
	}

	hostCount := len(s.Hosts)
	for _, hosts := range jobHosts {
		if hosts.exceedsMax(hostCount) {
			return false
		}
	}
	return true
}

func PrintJobs(s *State, w io.Writer) {
	jobHosts := make(map[string]hostList)
	for _, vm := range s.VMs {
		jobHosts[vm.Job] = append(jobHosts[vm.Job], vm.Host)
	}

	hostCount := len(s.Hosts)
	for jobName, hosts := range jobHosts {
		isBalanced := !hosts.exceedsMax(hostCount)
		var status string
		if isBalanced {
			status = balancedIndicator
		} else {
			status = unbalancedIndicator
		}
		fmt.Printf("%s  %s\n", status, jobName)
	}
}

// Balance accepts an unbalanced state as input and produces a new
// balanced state as output.
func Balance(s *State) *State {
	return s
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
