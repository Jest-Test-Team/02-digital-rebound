package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/dto"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/service"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/store"
	"github.com/Jest-Test-Team/02-digital-rebound/src/internal/vo"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/healthz", h.Health)
	v1 := r.Group("/v1/digital-rebound")
	{
		v1.POST("/events", h.IngestEvent)
		v1.GET("/assessments", h.ListAssessments)
		v1.POST("/assessments/analyze", h.Analyze)
		v1.POST("/assessments/:id/annotations", h.Annotate)
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Health())
}

func (h *Handler) IngestEvent(c *gin.Context) {
	var req dto.IngestEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "invalid request", Details: []string{err.Error()}})
		return
	}
	resp, details, err := h.svc.IngestEvent(req)
	if err != nil {
		if len(details) > 0 {
			c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "schema validation failed", Details: details})
			return
		}
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, resp)
}

func (h *Handler) ListAssessments(c *gin.Context) {
	tenantID := service.NormalizeTenantQuery(c.Query("tenant_id"))
	items, err := h.svc.ListAssessments(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}
	if items == nil {
		items = []vo.Assessment{}
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) Analyze(c *gin.Context) {
	var req dto.AnalyzeSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "invalid request", Details: []string{err.Error()}})
		return
	}
	resp, err := h.svc.Analyze(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Annotate(c *gin.Context) {
	tenantID := service.NormalizeTenantQuery(c.Query("tenant_id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "tenant_id query parameter is required"})
		return
	}
	var req dto.AnnotateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.ErrorResponse{Error: "invalid request", Details: []string{err.Error()}})
		return
	}
	a, err := h.svc.Annotate(tenantID, c.Param("id"), req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, vo.ErrorResponse{Error: "assessment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, vo.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}
