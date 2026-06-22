package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/gupta/leetcode-judge/internal/service"
)

type ProblemHandler struct {
	svc *service.ProblemService
}

func NewProblemHandler(svc *service.ProblemService) *ProblemHandler {
	return &ProblemHandler{svc: svc}
}

// --- Request shapes ---

type CreateProblemRequest struct {
	Title       string             `json:"title"       binding:"required,min=3,max=255"`
	Description string             `json:"description" binding:"required"`
	Difficulty  models.Difficulty  `json:"difficulty"  binding:"required"`
	Tags        []string           `json:"tags"`
	Examples    string             `json:"examples"`
	Constraints string             `json:"constraints"`
	IsPublished bool               `json:"isPublished"`
}

type UpdateProblemRequest struct {
	Title       string             `json:"title"       binding:"required,min=3,max=255"`
	Description string             `json:"description" binding:"required"`
	Difficulty  models.Difficulty  `json:"difficulty"  binding:"required"`
	Tags        []string           `json:"tags"`
	Examples    string             `json:"examples"`
	Constraints string             `json:"constraints"`
	IsPublished bool               `json:"isPublished"`
}

// --- Handlers ---

// Create godoc
// POST /api/v1/problems  (admin only)
func (h *ProblemHandler) Create(c *gin.Context) {
	var req CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	createdBy := c.GetString("userID")
	problem, err := h.svc.Create(c.Request.Context(), service.CreateProblemInput{
		Title:       req.Title,
		Description: req.Description,
		Difficulty:  req.Difficulty,
		Tags:        req.Tags,
		Examples:    req.Examples,
		Constraints: req.Constraints,
		IsPublished: req.IsPublished,
		CreatedBy:   createdBy,
	})
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "CREATE_FAILED")
		return
	}

	common.Success(c, http.StatusCreated, "problem created", problem)
}

// List godoc
// GET /api/v1/problems?difficulty=easy&page=1&pageSize=20
func (h *ProblemHandler) List(c *gin.Context) {
	difficulty := c.Query("difficulty")
	page, _     := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// Admins see unpublished problems too
	role    := c.GetString("role")
	isAdmin := role == "admin"

	problems, err := h.svc.List(c.Request.Context(), difficulty, isAdmin, page, pageSize)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "LIST_FAILED")
		return
	}

	common.Success(c, http.StatusOK, "problems fetched", gin.H{
		"problems": problems,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetByID godoc
// GET /api/v1/problems/:id
func (h *ProblemHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	problem, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Error(c, http.StatusNotFound, "problem not found", "NOT_FOUND")
		return
	}

	common.Success(c, http.StatusOK, "problem fetched", problem)
}

// Update godoc
// PUT /api/v1/problems/:id  (admin only)
func (h *ProblemHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	problem, err := h.svc.Update(c.Request.Context(), id, service.UpdateProblemInput{
		Title:       req.Title,
		Description: req.Description,
		Difficulty:  req.Difficulty,
		Tags:        req.Tags,
		Examples:    req.Examples,
		Constraints: req.Constraints,
		IsPublished: req.IsPublished,
	})
	if err != nil {
		common.Error(c, http.StatusInternalServerError, err.Error(), "UPDATE_FAILED")
		return
	}

	common.Success(c, http.StatusOK, "problem updated", problem)
}

// Delete godoc
// DELETE /api/v1/problems/:id  (admin only)
func (h *ProblemHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		common.Error(c, http.StatusNotFound, "problem not found", "NOT_FOUND")
		return
	}

	common.Success(c, http.StatusOK, "problem deleted", nil)
}
