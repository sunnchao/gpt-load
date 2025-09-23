package middleware

import (
	"gpt-load/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserAuth 用户认证中间件
func UserAuth(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "缺少认证令牌",
			})
			c.Abort()
			return
		}

		// 检查Bearer格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "认证令牌为空",
			})
			c.Abort()
			return
		}

		// 验证令牌
		user, err := userService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		allowed := false
		for _, allowedRole := range roles {
			if role == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 需要管理员权限
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}