package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/config"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/handler"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/service"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/store"
)

func main() {
	cfg := config.Load()
	cfg.SchemaPath = resolveSchemaPath(cfg.SchemaPath)

	fs, err := store.NewFileStore(cfg.EvidenceDir, cfg.AssessmentDir)
	if err != nil {
		log.Fatalf("store init: %v", err)
	}
	svc, err := service.New(cfg, fs)
	if err != nil {
		log.Fatalf("service init: %v", err)
	}
	if err := svc.EnsureSchemaReadable(); err != nil {
		log.Fatalf("schema path %s: %v", cfg.SchemaPath, err)
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	handler.New(svc).Register(r)

	log.Printf("digital-rebound listening on %s (rules=%s)", cfg.Addr, cfg.RuleVersion)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func resolveSchemaPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	candidates := []string{p}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, p), filepath.Join(wd, "..", p), filepath.Join(wd, "../..", p))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return p
}
