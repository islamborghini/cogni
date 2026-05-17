package bench

import "sort"

// Cost weights expressed as multipliers of the base input-token rate, so the
// weighted total is in "input-token equivalents". Values match Anthropic's
// public pricing for Claude Sonnet 4: $3/Mtok input, $15/Mtok output, cache
// writes at 1.25× input, cache reads at 0.1× input.
const (
	costWeightOutput        = 5.0
	costWeightCacheCreation = 1.25
	costWeightCacheRead     = 0.10
)

// CondStats summarises a (task, condition) cell across its n runs.
type CondStats struct {
	Condition         Condition
	Runs              int
	PassCount         int
	MeanInputTokens   float64
	MedianInput       int
	MeanOutput        float64
	MedianOutput      int
	MeanTotal         float64
	MedianTotal       int
	MeanCacheCreation float64
	MeanCacheRead     float64
	// MeanWeighted is the cost-weighted total expressed in input-token
	// equivalents (see weight constants above). Comparing this across
	// conditions is closer to comparing actual API spend than MeanTotal.
	MeanWeighted float64
}

// TaskSummary holds the side-by-side baseline vs cogni stats for one task.
// TokenReductionPct is the historical headline: positive means cogni used
// fewer raw (input+output, no cache) tokens than baseline. CostReductionPct
// uses the price-weighted total and is the more honest measure when cogni
// has materially more cached context than baseline.
type TaskSummary struct {
	TaskID            string
	Family            Family
	Baseline          CondStats
	Cogni             CondStats
	TokenReductionPct float64
	CostReductionPct  float64
}

type cell struct {
	input, output, total                  []int
	cacheCreation, cacheRead              []int
	weighted                              []float64
	passed                                int
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
		c.cacheCreation = append(c.cacheCreation, s.Run.CacheCreationTokens)
		c.cacheRead = append(c.cacheRead, s.Run.CacheReadTokens)
		c.weighted = append(c.weighted, weightedCost(s.Run))
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
		ts.CostReductionPct = reduction(ts.Baseline.MeanWeighted, ts.Cogni.MeanWeighted)
		out = append(out, ts)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TaskID < out[j].TaskID })
	return out
}

func statsFromCell(cond Condition, c cell) CondStats {
	return CondStats{
		Condition:         cond,
		Runs:              len(c.total),
		PassCount:         c.passed,
		MeanInputTokens:   mean(c.input),
		MedianInput:       median(c.input),
		MeanOutput:        mean(c.output),
		MedianOutput:      median(c.output),
		MeanTotal:         mean(c.total),
		MedianTotal:       median(c.total),
		MeanCacheCreation: mean(c.cacheCreation),
		MeanCacheRead:     mean(c.cacheRead),
		MeanWeighted:      meanFloat(c.weighted),
	}
}

func weightedCost(r RunResult) float64 {
	return float64(r.InputTokens) +
		costWeightOutput*float64(r.OutputTokens) +
		costWeightCacheCreation*float64(r.CacheCreationTokens) +
		costWeightCacheRead*float64(r.CacheReadTokens)
}

func meanFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
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
