package rules

import "testing"

func TestEvaluatePartialRebound(t *testing.T) {
	// unit cost down 20%, total consumption down only 5% → partial rebound
	res := Evaluate(SeriesInput{
		UnitCostBefore:     []float64{10, 10, 10},
		UnitCostAfter:      []float64{8, 8, 8},
		TotalConsumeBefore: []float64{100, 100, 100},
		TotalConsumeAfter:  []float64{95, 95, 95},
	})
	if res.Grade != GradePartial {
		t.Fatalf("expected partial, got %s", res.Grade)
	}
	if res.Metrics["rebound_ratio"] <= 0 || res.Metrics["rebound_ratio"] >= 1 {
		t.Fatalf("unexpected rebound_ratio=%v", res.Metrics["rebound_ratio"])
	}
}

func TestEvaluateBackfire(t *testing.T) {
	// unit cost down 20%, total consumption up 10% → backfire
	res := Evaluate(SeriesInput{
		UnitCostBefore:     []float64{10, 10, 10},
		UnitCostAfter:      []float64{8, 8, 8},
		TotalConsumeBefore: []float64{100, 100, 100},
		TotalConsumeAfter:  []float64{110, 110, 110},
	})
	if res.Grade != GradeBackfire {
		t.Fatalf("expected backfire, got %s metrics=%v", res.Grade, res.Metrics)
	}
}

func TestEvaluateNone(t *testing.T) {
	// unit cost down 20%, total consumption down 20% → no rebound
	res := Evaluate(SeriesInput{
		UnitCostBefore:     []float64{10, 10, 10},
		UnitCostAfter:      []float64{8, 8, 8},
		TotalConsumeBefore: []float64{100, 100, 100},
		TotalConsumeAfter:  []float64{80, 80, 80},
	})
	if res.Grade != GradeNone {
		t.Fatalf("expected none, got %s metrics=%v", res.Grade, res.Metrics)
	}
}

func TestEvaluateMissingData(t *testing.T) {
	res := Evaluate(SeriesInput{})
	if res.Grade != GradeInconclusive {
		t.Fatalf("expected inconclusive, got %s", res.Grade)
	}
	if len(res.MissingData) == 0 {
		t.Fatal("expected missing_data entries")
	}
}
