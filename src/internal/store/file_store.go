package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/vo"
)

var ErrNotFound = errors.New("not found")

type EvidenceRecord struct {
	EvidenceID    string                 `json:"evidence_id"`
	TenantID      string                 `json:"tenant_id"`
	Project       string                 `json:"project"`
	Source        string                 `json:"source"`
	ObservedAt    string                 `json:"observed_at"`
	IngestedAt    string                 `json:"ingested_at"`
	Metric        string                 `json:"metric"`
	Value         float64                `json:"value"`
	Unit          string                 `json:"unit,omitempty"`
	EvidenceRef   string                 `json:"evidence_ref"`
	SchemaVersion string                 `json:"schema_version"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	ContentHash   string                 `json:"content_hash"`
}

type FileStore struct {
	evidenceDir   string
	assessmentDir string
	mu            sync.Mutex
}

func NewFileStore(evidenceDir, assessmentDir string) (*FileStore, error) {
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(assessmentDir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{evidenceDir: evidenceDir, assessmentDir: assessmentDir}, nil
}

func (s *FileStore) SaveEvidence(rec EvidenceRecord) (EvidenceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload, err := json.Marshal(rec)
	if err != nil {
		return EvidenceRecord{}, err
	}
	sum := sha256.Sum256(payload)
	rec.ContentHash = "sha256:" + hex.EncodeToString(sum[:])
	if rec.EvidenceRef == "" {
		rec.EvidenceRef = rec.ContentHash
	}
	if rec.EvidenceID == "" {
		rec.EvidenceID = fmt.Sprintf("ev-%d", time.Now().UTC().UnixNano())
	}

	dir := filepath.Join(s.evidenceDir, sanitize(rec.TenantID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return EvidenceRecord{}, err
	}
	out, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return EvidenceRecord{}, err
	}
	path := filepath.Join(dir, rec.EvidenceID+".json")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return EvidenceRecord{}, err
	}
	return rec, nil
}

func (s *FileStore) ListEvidence(tenantID string) ([]EvidenceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.evidenceDir, sanitize(tenantID))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []EvidenceRecord{}, nil
		}
		return nil, err
	}
	out := make([]EvidenceRecord, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var rec EvidenceRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			return nil, err
		}
		if rec.TenantID != tenantID {
			continue
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EvidenceID < out[j].EvidenceID })
	return out, nil
}

func (s *FileStore) SaveAssessment(a vo.Assessment) (vo.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if a.AssessmentID == "" {
		a.AssessmentID = fmt.Sprintf("as-%d", time.Now().UTC().UnixNano())
	}
	if a.CreatedAt == "" {
		a.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	dir := filepath.Join(s.assessmentDir, sanitize(a.TenantID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return vo.Assessment{}, err
	}
	out, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return vo.Assessment{}, err
	}
	path := filepath.Join(dir, a.AssessmentID+".json")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return vo.Assessment{}, err
	}
	return a, nil
}

func (s *FileStore) ListAssessments(tenantID string) ([]vo.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tenantID != "" {
		return s.listAssessmentsLocked(tenantID)
	}

	entries, err := os.ReadDir(s.assessmentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []vo.Assessment{}, nil
		}
		return nil, err
	}
	all := make([]vo.Assessment, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		items, err := s.listAssessmentsLocked(e.Name())
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].AssessmentID < all[j].AssessmentID })
	return all, nil
}

func (s *FileStore) listAssessmentsLocked(tenantID string) ([]vo.Assessment, error) {
	dir := filepath.Join(s.assessmentDir, sanitize(tenantID))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []vo.Assessment{}, nil
		}
		return nil, err
	}
	out := make([]vo.Assessment, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var a vo.Assessment
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		if a.TenantID != tenantID {
			continue
		}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AssessmentID < out[j].AssessmentID })
	return out, nil
}

func (s *FileStore) GetAssessment(tenantID, assessmentID string) (vo.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.assessmentDir, sanitize(tenantID), assessmentID+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vo.Assessment{}, ErrNotFound
		}
		return vo.Assessment{}, err
	}
	var a vo.Assessment
	if err := json.Unmarshal(b, &a); err != nil {
		return vo.Assessment{}, err
	}
	if a.TenantID != tenantID {
		return vo.Assessment{}, ErrNotFound
	}
	return a, nil
}

func sanitize(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "/", "_")
	v = strings.ReplaceAll(v, "..", "_")
	if v == "" {
		return "_unknown"
	}
	return v
}
