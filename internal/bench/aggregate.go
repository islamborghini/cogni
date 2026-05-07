package bench

import "sort"

// CondStats summarises a (task, condition) cell across its n runs.
type CondStats struct {
	Condition       Condition
	Runs            int
	PassCount       int
	MeanInputTokens float64
	MedianInput     int
	MeanOutput      float64
	MedianOutput    int
	MeanTotal       float64
	MedianTotal     int
}

// TaskSummary holds the side-by-side baseline vs cogni stats for one task.
// TokenReductionPct is the headline number: positive means cogni used fewer
// total tokens than baseline. Zero when baseline mean is zero (no runs).
type TaskSummary struct {
	TaskID            string
	Family            Family
	Baseline          CondStats
	Cogni             CondStats
	TokenReductionPct float64
}

type cell struct {
	input, output, total []int
	passed               int
}

type pair struct {
	baseline cell
	cogni    cell
}

// Aggregate groups scores by task and condition and computes per-task
// summaries. Tasks with no scores at all are omitted.
func Aggregate(set *TaskSet, scores []Score) []TaskSummary {
	byTask := map[string]*pair{}
	for _, s := range scores {
		p, ok := byTask[s.Run.TaskID]
		if !ok {
			p = &pair{}
			byTask[s.Run.TaskID] = p
		}
		var c *cell
		switch s.Run.Condition {
		case ConditionBaseline:
			c = &p.baseline
		case ConditionCogni:
			c = &p.cogni
		default:
			continue
		}
		c.input = append(c.input, s.Run.InputTokens)
		c.output = append(c.output, s.Run.OutputTokens)
		c.total = append(c.total, s.Run.InputTokens+s.Run.OutputTokens)
		if s.Pass {
			c.passed++
		}
	}

	out := make([]TaskSummary, 0, len(byTask))
	for _, t := range set.Tasks {
		p, ok := byTask[t.ID]
		if !ok {
			continue
		}
		ts := TaskSummary{
			TaskID:   t.ID,
			Family:   t.Family,
			Baseline: statsFromCell(ConditionBaseline, p.baseline),
			Cogni:    statsFromCell(ConditionCogni, p.cogni),
		}
		ts.TokenReductionPct = reduction(ts.Baseline.MeanTotal, ts.Cogni.MeanTotal)
		out = append(out, ts)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TaskID < out[j].TaskID })
	return out
}

func statsFromCell(cond Condition, c cell) CondStats {
	return CondStats{
		Condition:       cond,
		Runs:            len(c.total),
		PassCount:       c.passed,
		MeanInputTokens: mean(c.input),
		MedianInput:     median(c.input),
		MeanOutput:      mean(c.output),
		MedianOutput:    median(c.output),
		MeanTotal:       mean(c.total),
		MedianTotal:     median(c.total),
	}
}

func mean(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0
	for _, x := range xs {
		sum += x
	}
	return float64(sum) / float64(len(xs))
}

func median(xs []int) int {
	if len(xs) == 0 {
		return 0
	}
	s := append([]int(nil), xs...)
	sort.Ints(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

func reduction(baseline, cogni float64) float64 {
	if baseline == 0 {
		return 0
	}
	return (baseline - cogni) / baseline * 100
}
