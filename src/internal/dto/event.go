package dto

// IngestEventRequest is the API request body for evidence ingestion.
type IngestEventRequest struct {
	TenantID      string                 `json:"tenant_id" binding:"required,min=1"`
	Project       string                 `json:"project" binding:"required"`
	Source        string                 `json:"source" binding:"required,min=1"`
	ObservedAt    string                 `json:"observed_at" binding:"required"`
	IngestedAt    string                 `json:"ingested_at"`
	Metric        string                 `json:"metric" binding:"required"`
	Value         *float64               `json:"value" binding:"required"`
	Unit          string                 `json:"unit"`
	EvidenceRef   string                 `json:"evidence_ref"`
	SchemaVersion string                 `json:"schema_version" binding:"required"`
	Attributes    map[string]interface{} `json:"attributes"`
}

// AnalyzeSeriesRequest triggers a rebound assessment from before/after windows.
type AnalyzeSeriesRequest struct {
	TenantID           string    `json:"tenant_id" binding:"required,min=1"`
	OptimizationID     string    `json:"optimization_id" binding:"required,min=1"`
	UnitCostBefore     []float64 `json:"unit_cost_before" binding:"required,min=1"`
	UnitCostAfter      []float64 `json:"unit_cost_after" binding:"required,min=1"`
	TotalConsumeBefore []float64 `json:"total_consumption_before" binding:"required,min=1"`
	TotalConsumeAfter  []float64 `json:"total_consumption_after" binding:"required,min=1"`
	EvidenceRefs       []string  `json:"evidence_refs"`
	ObservedAt         string    `json:"observed_at"`
}

// AnnotateAssessmentRequest appends a human review annotation.
type AnnotateAssessmentRequest struct {
	Reviewer string `json:"reviewer" binding:"required,min=1"`
	Decision string `json:"decision" binding:"required,oneof=acknowledged needs-more-data rejected corrected accepted"`
	Note     string `json:"note"`
}
