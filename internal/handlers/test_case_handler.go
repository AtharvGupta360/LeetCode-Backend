package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/service"
)

type TestCaseHandler struct {
	svc *service.TestCaseService
}

func NewTestCaseHandler(svc *service.TestCaseService) *TestCaseHandler {
	return &TestCaseHandler{svc: svc}
}

// --- Request shapes ---

type CreateTestCaseRequest struct {
	Input          string `json:"input"          binding:"required"`
	ExpectedOutput string `json:"expectedOutput" binding:"required"`
	IsHidden       bool   `json:"isHidden"`
	OrderIndex     int    `json:"orderIndex"`
}

type UpdateTestCaseRequest struct {
	Input          string `json:"input"          binding:"required"`
	ExpectedOutput string `json:"expectedOutput" binding:"required"`
	IsHidden       bool   `json:"isHidden"`
	OrderIndex     int    `json:"orderIndex"`
}

// --- Handlers ---

// Create godoc
// POST /api/v1/problems/:problemId/test-cases  (admin only)
func (h *TestCaseHandler) Create(c *gin.Context) {
	problemID := c.Param("problemId")

	var req CreateTestCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	tc, err := h.svc.Create(c.Request.Context(), service.CreateTestCaseInput{
		ProblemID:      problemID,
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
		IsHidden:       req.IsHidden,
		OrderIndex:     req.OrderIndex,
	})
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "CREATE_FAILED")
		return
	}

	common.Success(c, http.StatusCreated, "test case created", tc)
}

// List godoc
// GET /api/v1/problems/:problemId/test-cases
// Admins see all (including hidden); regular users only see visible ones.
func (h *TestCaseHandler) List(c *gin.Context) {
	problemID := c.Param("problemId")
	isAdmin := c.GetString("role") == "admin"

	cases, err := h.svc.List(c.Request.Context(), problemID, isAdmin)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "LIST_FAILED")
		return
	}

	common.Success(c, http.StatusOK, "test cases fetched", cases)
}

// GetByID godoc
// GET /api/v1/problems/:problemId/test-cases/:id
func (h *TestCaseHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	isAdmin := c.GetString("role") == "admin"

	tc, err := h.svc.GetByID(c.Request.Context(), id, isAdmin)
	if err != nil {
		common.Error(c, http.StatusNotFound, "test case not found", "NOT_FOUND")
		return
	}

	common.Success(c, http.StatusOK, "test case fetched", tc)
}

// Update godoc
// PUT /api/v1/problems/:problemId/test-cases/:id  (admin only)
func (h *TestCaseHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTestCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	tc, err := h.svc.Update(c.Request.Context(), id, service.UpdateTestCaseInput{
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
		IsHidden:       req.IsHidden,
		OrderIndex:     req.OrderIndex,
	})
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "UPDATE_FAILED")
		return
	}

	common.Success(c, http.StatusOK, "test case updated", tc)
}

// Delete godoc
// DELETE /api/v1/problems/:problemId/test-cases/:id  (admin only)
func (h *TestCaseHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		common.Error(c, http.StatusNotFound, "test case not found", "NOT_FOUND")
		return
	}

	common.Success(c, http.StatusOK, "test case deleted", nil)
}
