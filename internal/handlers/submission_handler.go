package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/service"
)

type SubmissionHandler struct {
	svc *service.SubmissionService
}

func NewSubmissionHandler(svc *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{svc: svc}
}

// --- Request shapes ---

type SubmitRequest struct {
	ProblemID string `json:"problemId" binding:"required"`
	Language  string `json:"language"  binding:"required"`
	Code      string `json:"code"      binding:"required"`
}

// --- Handlers ---

// Submit godoc
// POST /api/v1/submissions  (authenticated user)
func (h *SubmissionHandler) Submit(c *gin.Context) {
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	userID := c.GetString("userID")
	sub, err := h.svc.Submit(c.Request.Context(), service.SubmitInput{
		UserID:    userID,
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
	})
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "SUBMIT_FAILED")
		return
	}

	common.Success(c, http.StatusCreated, "submission created", sub)
}

// GetByID godoc
// GET /api/v1/submissions/:id  (authenticated user — poll for verdict)
func (h *SubmissionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	sub, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Error(c, http.StatusNotFound, "submission not found", "NOT_FOUND")
		return
	}

	common.Success(c, http.StatusOK, "submission fetched", sub)
}

// ListMine godoc
// GET /api/v1/submissions/me?page=1&pageSize=20  (authenticated user — own submissions)
func (h *SubmissionHandler) ListMine(c *gin.Context) {
	userID := c.GetString("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	subs, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "LIST_FAILED")
		return
	}

	common.Success(c, http.StatusOK, "submissions fetched", gin.H{
		"submissions": subs,
		"page":        page,
		"pageSize":    pageSize,
	})
}

// ListByProblem godoc
// GET /api/v1/problems/:problemId/submissions?page=1&pageSize=20  (admin only)
func (h *SubmissionHandler) ListByProblem(c *gin.Context) {
	problemID := c.Param("problemId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	subs, err := h.svc.ListByProblem(c.Request.Context(), problemID, page, pageSize)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "LIST_FAILED")
		return
	}

	common.Success(c, http.StatusOK, "submissions fetched", gin.H{
		"submissions": subs,
		"page":        page,
		"pageSize":    pageSize,
	})
}
