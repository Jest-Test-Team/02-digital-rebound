package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/config"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/dto"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/rules"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/store"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/vo"
)

const disclaimer = "Synthetic-data research assessment only. Not a causal claim; human review required before consequential action."

type Service struct {
	cfg    config.Config
	store  *store.FileStore
	schema *jsonschema.Schema
}

func New(cfg config.Config, fs *store.FileStore) (*Service, error) {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	sch, err := compiler.Compile(cfg.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return &Service{cfg: cfg, store: fs, schema: sch}, nil
}

func (s *Service) Health() vo.HealthResponse {
	return vo.HealthResponse{
		OK:          true,
		Service:     "digital-rebound",
		RuleVersion: s.cfg.RuleVersion,
	}
}

func (s *Service) IngestEvent(req dto.IngestEventRequest) (vo.IngestAcceptedResponse, []string, error) {
	if details := s.validateEvent(req); len(details) > 0 {
		return vo.IngestAcceptedResponse{}, details, fmt.Errorf("validation failed")
	}

	ingestedAt := req.IngestedAt
	if ingestedAt == "" {
		ingestedAt = time.Now().UTC().Format(time.RFC3339)
	}

	rec := store.EvidenceRecord{
		TenantID:      req.TenantID,
		Project:       req.Project,
		Source:        req.Source,
		ObservedAt:    req.ObservedAt,
		IngestedAt:    ingestedAt,
		Metric:        req.Metric,
		Value:         *req.Value,
		Unit:          req.Unit,
		EvidenceRef:   req.EvidenceRef,
		SchemaVersion: req.SchemaVersion,
		Attributes:    req.Attributes,
	}
	saved, err := s.store.SaveEvidence(rec)
	if err != nil {
		return vo.IngestAcceptedResponse{}, nil, err
	}

	return vo.IngestAcceptedResponse{
		Status:      "accepted",
		EvidenceID:  saved.EvidenceID,
		EvidenceRef: saved.EvidenceRef,
		TenantID:    saved.TenantID,
	}, nil, nil
}

func (s *Service) validateEvent(req dto.IngestEventRequest) []string {
	details := make([]string, 0)

	raw, err := json.Marshal(req)
	if err != nil {
		return []string{"invalid json payload"}
	}
	var doc interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return []string{"invalid json payload"}
	}
	if err := s.schema.Validate(doc); err != nil {
		details = append(details, err.Error())
	}
	if req.Project != "digital-rebound" {
		details = append(details, "project must be digital-rebound")
	}
	if req.SchemaVersion != "0.1.0" {
		details = append(details, "schema_version must be 0.1.0")
	}
	if _, err := time.Parse(time.RFC3339, req.ObservedAt); err != nil {
		details = append(details, "observed_at must be RFC3339 date-time")
	}
	if req.IngestedAt != "" {
		if _, err := time.Parse(time.RFC3339, req.IngestedAt); err != nil {
			details = append(details, "ingested_at must be RFC3339 date-time")
		}
	}
	return details
}

func (s *Service) Analyze(req dto.AnalyzeSeriesRequest) (vo.AnalyzeAcceptedResponse, error) {
	result := rules.Evaluate(rules.SeriesInput{
		UnitCostBefore:     req.UnitCostBefore,
		UnitCostAfter:      req.UnitCostAfter,
		TotalConsumeBefore: req.TotalConsumeBefore,
		TotalConsumeAfter:  req.TotalConsumeAfter,
	})

	stale := false
	uncertainty := result.Uncertainty
	if req.ObservedAt != "" {
		if observed, err := time.Parse(time.RFC3339, req.ObservedAt); err == nil {
			if time.Since(observed) > s.cfg.StaleAfter {
				stale = true
				uncertainty = minFloat(1, uncertainty+0.25)
				result.Assumptions = append(result.Assumptions, "Evidence observed_at exceeds stale threshold relative to analysis time.")
			}
		}
	}

	refs := req.EvidenceRefs
	if refs == nil {
		refs = []string{}
	}

	assessment := vo.Assessment{
		TenantID:            req.TenantID,
		OptimizationID:      req.OptimizationID,
		Status:              "review-required",
		Summary:             result.Summary,
		Grade:               result.Grade,
		Uncertainty:         uncertainty,
		EvidenceRefs:        refs,
		HumanReviewRequired: true,
		RuleVersion:         s.cfg.RuleVersion,
		Metrics:             result.Metrics,
		Assumptions:         result.Assumptions,
		MissingData:         result.MissingData,
		StaleEvidence:       stale,
		Disclaimer:          disclaimer,
	}

	saved, err := s.store.SaveAssessment(assessment)
	if err != nil {
		return vo.AnalyzeAcceptedResponse{}, err
	}
	return vo.AnalyzeAcceptedResponse{Status: "created", Assessment: saved}, nil
}

func (s *Service) ListAssessments(tenantID string) ([]vo.Assessment, error) {
	return s.store.ListAssessments(tenantID)
}

func (s *Service) Annotate(tenantID, assessmentID string, req dto.AnnotateAssessmentRequest) (vo.Assessment, error) {
	a, err := s.store.GetAssessment(tenantID, assessmentID)
	if err != nil {
		return vo.Assessment{}, err
	}
	a.Annotations = append(a.Annotations, vo.Annotation{
		Reviewer:  req.Reviewer,
		Decision:  req.Decision,
		Note:      req.Note,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	switch req.Decision {
	case "accepted", "acknowledged", "corrected":
		a.Status = "accepted"
	case "rejected":
		a.Status = "rejected"
	case "needs-more-data":
		a.Status = "review-required"
	}
	a.HumanReviewRequired = true
	return s.store.SaveAssessment(a)
}

func (s *Service) EnsureSchemaReadable() error {
	_, err := os.Stat(s.cfg.SchemaPath)
	return err
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func NormalizeTenantQuery(raw string) string {
	return strings.TrimSpace(raw)
}
