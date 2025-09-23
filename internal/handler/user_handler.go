package handler

import (
	"gpt-load/internal/models"
	"gpt-load/internal/response"
	"gpt-load/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册（仅管理员可用）
func (h *UserHandler) Register(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, "请求参数无效: "+err.Error())
		return
	}

	// 检查当前用户权限
	currentUser := h.getCurrentUser(c)
	if currentUser == nil || currentUser.Role != models.RoleAdmin {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.SuccessWithData(c, user)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, "请求参数无效: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	user, token, err := h.userService.Authenticate(req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.SuccessWithData(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	response.SuccessWithData(c, user)
}

// UpdateProfile 更新当前用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, "请求参数无效: "+err.Error())
		return
	}

	// 非管理员用户不能修改角色和状态
	if user.Role != models.RoleAdmin {
		req.Role = nil
		req.Status = nil
	}

	err := h.userService.UpdateUser(user.ID, req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c)
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, "请求参数无效: "+err.Error())
		return
	}

	err := h.userService.ChangePassword(user.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c)
}

// ListUsers 获取用户列表（仅管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	currentUser := h.getCurrentUser(c)
	if currentUser == nil || currentUser.Role != models.RoleAdmin {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	role := c.Query("role")
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userService.ListUsers(page, pageSize, role, status, search)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.SuccessWithData(c, gin.H{
		"list": users,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUser 获取指定用户信息（仅管理员）
func (h *UserHandler) GetUser(c *gin.Context) {
	currentUser := h.getCurrentUser(c)
	if currentUser == nil || currentUser.Role != models.RoleAdmin {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorWithMessage(c, "无效的用户ID")
		return
	}

	users, _, err := h.userService.ListUsers(1, 1, "", "", "")
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	// 找到指定用户
	var targetUser *models.User
	for _, user := range users {
		if user.ID == uint(userID) {
			targetUser = user
			break
		}
	}

	if targetUser == nil {
		response.ErrorWithMessage(c, "用户不存在")
		return
	}

	response.SuccessWithData(c, targetUser)
}

// UpdateUser 更新指定用户信息（仅管理员）
func (h *UserHandler) UpdateUser(c *gin.Context) {
	currentUser := h.getCurrentUser(c)
	if currentUser == nil || currentUser.Role != models.RoleAdmin {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorWithMessage(c, "无效的用户ID")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, "请求参数无效: "+err.Error())
		return
	}

	err = h.userService.UpdateUser(uint(userID), req)
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c)
}

// DeleteUser 删除用户（仅管理员）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	currentUser := h.getCurrentUser(c)
	if currentUser == nil || currentUser.Role != models.RoleAdmin {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ErrorWithMessage(c, "无效的用户ID")
		return
	}

	// 不能删除自己
	if uint(userID) == currentUser.ID {
		response.ErrorWithMessage(c, "不能删除自己的账户")
		return
	}

	err = h.userService.DeleteUser(uint(userID))
	if err != nil {
		response.ErrorWithMessage(c, err.Error())
		return
	}

	response.Success(c)
}

// Logout 用户退出
func (h *UserHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		h.userService.Logout(token)
	}

	response.Success(c)
}

// 辅助方法
func (h *UserHandler) getCurrentUser(c *gin.Context) *models.User {
	userValue, exists := c.Get("user")
	if !exists {
		return nil
	}

	user, ok := userValue.(*models.User)
	if !ok {
		return nil
	}

	return user
}