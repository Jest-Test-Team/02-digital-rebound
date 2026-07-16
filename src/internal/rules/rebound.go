package rules

import "math"

const Version = "rebound-rules@0.1.0"

const (
	GradeNone         = "none"
	GradePartial      = "partial"
	GradeBackfire     = "backfire"
	GradeInconclusive = "inconclusive"
)

// SeriesInput holds before/after windows for rebound metrics.
type SeriesInput struct {
	UnitCostBefore     []float64
	UnitCostAfter      []float64
	TotalConsumeBefore []float64
	TotalConsumeAfter  []float64
}

// Result is the explainable metric output for one assessment.
type Result struct {
	Metrics     map[string]float64
	Grade       string
	Uncertainty float64
	Assumptions []string
	MissingData []string
	Summary     string
}

func mean(xs []float64) (float64, bool) {
	if len(xs) == 0 {
		return 0, false
	}
	var sum float64
	for _, v := range xs {
		sum += v
	}
	return sum / float64(len(xs)), true
}

// Evaluate computes pinned rebound metrics without ML.
func Evaluate(in SeriesInput) Result {
	missing := make([]string, 0)
	assumptions := []string{
		"Windows are synthetic and seasonality is not adjusted.",
		"Formulas are descriptive heuristics, not causal claims.",
		"Consequential actions require human review.",
	}

	unitBefore, okUB := mean(in.UnitCostBefore)
	unitAfter, okUA := mean(in.UnitCostAfter)
	totalBefore, okTB := mean(in.TotalConsumeBefore)
	totalAfter, okTA := mean(in.TotalConsumeAfter)

	if !okUB {
		missing = append(missing, "unit_cost_before")
	}
	if !okUA {
		missing = append(missing, "unit_cost_after")
	}
	if !okTB {
		missing = append(missing, "total_consumption_before")
	}
	if !okTA {
		missing = append(missing, "total_consumption_after")
	}

	if len(missing) > 0 {
		return Result{
			Metrics:     map[string]float64{},
			Grade:       GradeInconclusive,
			Uncertainty: 1.0,
			Assumptions: assumptions,
			MissingData: missing,
			Summary:     "Insufficient window data for rebound assessment.",
		}
	}

	unitDelta := (unitAfter - unitBefore) / unitBefore
	totalDelta := (totalAfter - totalBefore) / totalBefore

	expectedSavings := -unitDelta * totalBefore
	var reboundRatio float64
	if math.Abs(expectedSavings) < 1e-12 {
		missing = append(missing, "expected_savings_near_zero")
		return Result{
			Metrics: map[string]float64{
				"unit_cost_delta":         unitDelta,
				"total_consumption_delta": totalDelta,
			},
			Grade:       GradeInconclusive,
			Uncertainty: 0.9,
			Assumptions: assumptions,
			MissingData: missing,
			Summary:     "Expected savings near zero; rebound ratio inconclusive.",
		}
	}

	actualSavings := totalBefore - totalAfter
	reboundRatio = 1 - (actualSavings / expectedSavings)

	var elasticity float64
	if math.Abs(unitDelta) < 1e-12 {
		missing = append(missing, "unit_cost_delta_near_zero")
		elasticity = 0
	} else {
		elasticity = totalDelta / unitDelta
	}

	backfireProb := backfireProbability(reboundRatio, len(missing))
	grade := gradeFor(reboundRatio)

	uncertainty := 0.15
	if len(in.UnitCostBefore) < 3 || len(in.UnitCostAfter) < 3 {
		uncertainty += 0.2
		assumptions = append(assumptions, "Short windows increase uncertainty.")
	}

	summary := "Partial rebound suspected; efficiency gains may be offset by demand growth."
	switch grade {
	case GradeNone:
		summary = "No rebound detected in synthetic windows; total consumption fell with efficiency."
	case GradeBackfire:
		summary = "Backfire candidate: total consumption rose enough to offset or exceed expected savings."
	case GradeInconclusive:
		summary = "Rebound assessment inconclusive given available evidence."
	}

	return Result{
		Metrics: map[string]float64{
			"unit_cost_delta":         round4(unitDelta),
			"total_consumption_delta": round4(totalDelta),
			"rebound_ratio":           round4(reboundRatio),
			"demand_elasticity":       round4(elasticity),
			"backfire_probability":    round4(backfireProb),
		},
		Grade:       grade,
		Uncertainty: round4(math.Min(uncertainty, 1)),
		Assumptions: assumptions,
		MissingData: missing,
		Summary:     summary,
	}
}

func gradeFor(reboundRatio float64) string {
	switch {
	case reboundRatio <= 0:
		return GradeNone
	case reboundRatio >= 1:
		return GradeBackfire
	default:
		return GradePartial
	}
}

func backfireProbability(reboundRatio float64, missingCount int) float64 {
	var p float64
	switch {
	case reboundRatio >= 1:
		p = 1.0
	case reboundRatio >= 0.5:
		p = 0.7
	case reboundRatio >= 0.2:
		p = 0.4
	default:
		p = 0.1
	}
	p += float64(missingCount) * 0.05
	if p > 1 {
		p = 1
	}
	return p
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
