package magnet

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	balancedIndicator   = "✓"
	unbalancedIndicator = "✗"

	red   = color.New(color.FgRed).SprintfFunc()
	green = color.New(color.FgGreen).SprintfFunc()
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
	PrintJobs(s)
	if !IsBalanced(s) {
		rec := RuleRecommendations(s)
		rec.PrintReport()
		// err = i.Converge(ctx, s, rec)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}

// IsBalanced determines whether the state of a deployment is balanced.
// A deployment is balanced jobs are spread across as many hosts as possible.
func IsBalanced(s *State) bool {
	jobHosts := make(map[string]hostList)
	for _, vm := range s.VMs {
		jobHosts[vm.Job] = append(jobHosts[vm.Job], vm.HostUUID)
	}

	hostCount := len(s.Hosts)
	for _, hosts := range jobHosts {
		if hosts.exceedsMax(hostCount) {
			return false
		}
	}
	return true
}

// PrintJobs while indicating if each job is balanced
func PrintJobs(s *State) {
	jobHosts := make(map[string]hostList)
	for _, vm := range s.VMs {
		jobHosts[vm.Job] = append(jobHosts[vm.Job], vm.HostUUID)
	}

	hostCount := len(s.Hosts)
	for jobName, hosts := range jobHosts {
		isBalanced := !hosts.exceedsMax(hostCount)
		var status string
		if isBalanced {
			status = greenSprintf("%s", balancedIndicator)
		} else {
			status = redSprintf("%s", unbalancedIndicator)
		}
		fmt.Fprintf(output, "%s  %s\n", status, jobName)
	}
}

// PrintReport writes a user-friendly description of the reccomendations to w.
func (r *RuleRecommendation) PrintReport() {
	fmt.Fprintln(output, "Recommendations:")
	tw := tabwriter.NewWriter(output, 8, 4, 1, ' ', 0)
	defer tw.Flush()

	if len(r.Stale) > 0 {
		fmt.Fprintln(output, redSprintf("--REMOVE--"))
		for i := range r.Stale {
			writeRule(tw, &r.Stale[i])
		}
	}
	if len(r.Missing) > 0 {
		fmt.Fprintln(output, greenSprintf("--ADD--"))
		for i := range r.Missing {
			writeRule(tw, &r.Missing[i])
		}
	}
}

func writeRule(w io.Writer, r *Rule) {
	buf := &bytes.Buffer{}
	for i, vm := range r.VMs {
		if i > 0 && i < len(r.VMs) {
			fmt.Fprintf(buf, ", ")
		}
		fmt.Fprint(buf, vm.Name)
	}
	fmt.Fprintf(w, "%s\t%s\n", r.Name, buf.String())
}

// RuleRecommendations looks at the state of the system and makes reccomendations
// about how to achieve anti-affinity.
func RuleRecommendations(s *State) *RuleRecommendation {
	vmsForJob := make(map[string][]*VM)
	for _, vm := range s.VMs {
		vmsForJob[vm.Job] = append(vmsForJob[vm.Job], vm)
	}
	expectedRules := make(map[string]Rule)
	for job, vms := range vmsForJob {
		if len(vms) <= 1 {
			continue
		}
		expectedRules[job] = Rule{
			Name:      job,
			Enabled:   true,
			Mandatory: true,
			VMs:       vms,
		}
	}
	result := &RuleRecommendation{}

	existingRules := make(map[string]struct{})

	// identify each of our currently defined rules as valid or stale
	for _, currentRule := range s.Rules {
		exp, exists := expectedRules[currentRule.Name]
		if exists {
			existingRules[currentRule.Name] = struct{}{}
			if rulesEqual(currentRule, &exp) {
				// we already have a rule that is equivalent to the expected rule -> VALID
				result.Valid = append(result.Valid, *currentRule)
			} else {
				// we have a rule but it doesn't match -> STALE and MISSING
				result.Stale = append(result.Stale, *currentRule)
				result.Missing = append(result.Missing, exp)
			}
		} else {
			// we don't need a rule for this job (report it as stale)
			result.Stale = append(result.Stale, *currentRule)
		}
	}

	// identify any missing rules (rules that are expected but not valid)
	for _, expectedRule := range expectedRules {
		if _, exists := existingRules[expectedRule.Name]; exists {
			continue
		}
		result.Missing = append(result.Missing, expectedRule)
	}

	return result
}

// rulesEqual determines if two rules are logically equivalent.
// This means that the rules have the same name and consist of the
// same VMs.  The ID of the rules or the ordering of their VMs do
// not impact equivalence.
func rulesEqual(r0, r1 *Rule) bool {
	if r0.Name == r1.Name && len(r0.VMs) == len(r1.VMs) {
		r1VMs := make(map[*VM]bool)
		for _, vm := range r1.VMs {
			r1VMs[vm] = true
		}
		for _, vm := range r0.VMs {
			_, ok := r1VMs[vm]
			if !ok {
				return false
			}
		}
		return true
	}
	return false
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
