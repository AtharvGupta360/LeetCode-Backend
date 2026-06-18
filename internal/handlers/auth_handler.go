package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/auth"
	"github.com/gupta/leetcode-judge/internal/common"
)

type AuthHandler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	user, token, err := h.authService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		common.Error(c, http.StatusConflict, "username or email already exists", "DUPLICATE_USER")
		return
	}

	common.Success(c, http.StatusCreated, "user registered successfully", gin.H{
		"user":  user,
		"token": token,
	})
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		common.Error(c, http.StatusUnauthorized, "invalid credentials", "INVALID_CREDENTIALS")
		return
	}

	common.Success(c, http.StatusOK, "login successful", gin.H{
		"user":  user,
		"token": token,
	})
}

