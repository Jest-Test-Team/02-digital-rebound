package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr           string
	EvidenceDir    string
	AssessmentDir  string
	SchemaPath     string
	StaleAfter     time.Duration
	RuleVersion    string
	SingleTenantID string
}

func Load() Config {
	staleHours := envInt("STALE_AFTER_HOURS", 720) // 30 days
	return Config{
		Addr:           env("ADDR", ":8081"),
		EvidenceDir:    env("EVIDENCE_DIR", "data/evidence"),
		AssessmentDir:  env("ASSESSMENT_DIR", "data/assessments"),
		SchemaPath:     env("SCHEMA_PATH", "schemas/event.schema.json"),
		StaleAfter:     time.Duration(staleHours) * time.Hour,
		RuleVersion:    env("RULE_VERSION", "rebound-rules@0.1.0"),
		SingleTenantID: env("SINGLE_TENANT_ID", ""),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
