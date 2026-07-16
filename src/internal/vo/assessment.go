package vo

// HealthResponse is returned by GET /healthz.
type HealthResponse struct {
	OK          bool   `json:"ok"`
	Service     string `json:"service"`
	RuleVersion string `json:"rule_version"`
}

// IngestAcceptedResponse is returned on successful event ingest.
type IngestAcceptedResponse struct {
	Status      string `json:"status"`
	EvidenceID  string `json:"evidence_id"`
	EvidenceRef string `json:"evidence_ref"`
	TenantID    string `json:"tenant_id"`
}

// ErrorResponse is a standard API error body.
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

// Assessment is the explainable rebound assessment response.
type Assessment struct {
	AssessmentID        string             `json:"assessment_id"`
	TenantID            string             `json:"tenant_id"`
	OptimizationID      string             `json:"optimization_id,omitempty"`
	Status              string             `json:"status"`
	Summary             string             `json:"summary"`
	Grade               string             `json:"grade"`
	Uncertainty         float64            `json:"uncertainty"`
	EvidenceRefs        []string           `json:"evidence_refs"`
	HumanReviewRequired bool               `json:"human_review_required"`
	RuleVersion         string             `json:"rule_version"`
	Metrics             map[string]float64 `json:"metrics"`
	Assumptions         []string           `json:"assumptions"`
	MissingData         []string           `json:"missing_data"`
	StaleEvidence       bool               `json:"stale_evidence"`
	Disclaimer          string             `json:"disclaimer"`
	Annotations         []Annotation       `json:"annotations,omitempty"`
	CreatedAt           string             `json:"created_at"`
}

// Annotation is an append-only human review note.
type Annotation struct {
	Reviewer  string `json:"reviewer"`
	Decision  string `json:"decision"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`
}

// AnalyzeAcceptedResponse wraps a newly created assessment.
type AnalyzeAcceptedResponse struct {
	Status     string     `json:"status"`
	Assessment Assessment `json:"assessment"`
}
