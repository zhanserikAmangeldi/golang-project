package handler

import (
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/zhanserikAmangeldi/user-service/internal/dto"
    "github.com/zhanserikAmangeldi/user-service/internal/middleware"
    "github.com/zhanserikAmangeldi/user-service/internal/service"
)

type AdminHandler struct {
    adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
    return &AdminHandler{adminService: adminService}
}

// BanUser godoc
// @Summary Забанить пользователя
// @Description Блокирует доступ пользователя к системе на определенный срок или навсегда
// @Tags admin
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.BanUserRequest true "Параметры бана"
// @Success 200 {object} map[string]string "message: user banned successfully"
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/admin/bans [post]
func (h *AdminHandler) BanUser(c *gin.Context) {
    adminID := middleware.GetUserID(c)

    var req dto.BanUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    var duration *time.Duration
    if req.DurationMinutes > 0 {
        d := time.Duration(req.DurationMinutes) * time.Minute
        duration = &d
    }

    ip := c.ClientIP()
    err := h.adminService.BanUser(c.Request.Context(), adminID, req.UserID, req.Reason, duration, &ip)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "ban_failed",
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "user banned successfully",
    })
}

// UnbanUser godoc
// @Summary Разбанить пользователя
// @Description Снимает активную блокировку с пользователя
// @Tags admin
// @Security ApiKeyAuth
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} map[string]string "message: user unbanned successfully"
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/admin/bans/{user_id} [delete]
func (h *AdminHandler) UnbanUser(c *gin.Context) {
    adminID := middleware.GetUserID(c)

    userIDStr := c.Param("user_id")
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "invalid_user_id",
            Message: "User ID must be a number",
        })
        return
    }

    ip := c.ClientIP()
    err = h.adminService.UnbanUser(c.Request.Context(), adminID, userID, &ip)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "unban_failed",
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "user unbanned successfully",
    })
}

// UpdateUserRole godoc
// @Summary Изменить роль пользователя
// @Description Назначает пользователю новую роль (admin, moderator, user)
// @Tags admin
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateRoleRequest true "Данные для смены роли"
// @Success 200 {object} map[string]string "message: user role updated successfully"
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
    adminID := middleware.GetUserID(c)

    var req dto.UpdateRoleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "validation_error",
            Message: err.Error(),
        })
        return
    }

    ip := c.ClientIP()
    err := h.adminService.UpdateUserRole(c.Request.Context(), adminID, req.UserID, req.Role, &ip)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "role_update_failed",
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "user role updated successfully",
    })
}

// DeleteUser godoc
// @Summary Удалить пользователя
// @Description Полное удаление аккаунта пользователя из системы
// @Tags admin
// @Security ApiKeyAuth
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} map[string]string "message: user deleted successfully"
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{user_id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
    adminID := middleware.GetUserID(c)

    userIDStr := c.Param("user_id")
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "invalid_user_id",
            Message: "User ID must be a number",
        })
        return
    }

    ip := c.ClientIP()
    err = h.adminService.DeleteUser(c.Request.Context(), adminID, userID, &ip)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error:   "delete_failed",
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "user deleted successfully",
    })
}

// GetAllUsers godoc
// @Summary Список всех пользователей
// @Description Возвращает список пользователей с поддержкой пагинации и фильтрации по роли
// @Tags admin, moderator
// @Security ApiKeyAuth
// @Produce json
// @Param limit query int false "Количество (default 20)"
// @Param offset query int false "Смещение (default 0)"
// @Param role query string false "Фильтр по роли"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
    limitStr := c.DefaultQuery("limit", "20")
    offsetStr := c.DefaultQuery("offset", "0")
    role := c.Query("role")

    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    users, total, err := h.adminService.GetAllUsers(c.Request.Context(), limit, offset, role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Error: "failed_to_fetch_users",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "users":  users,
        "total":  total,
        "limit":  limit,
        "offset": offset,
    })
}

// GetUserBanHistory godoc
// @Summary История банов пользователя
// @Description Возвращает все записи о блокировках конкретного пользователя
// @Tags admin, moderator
// @Security ApiKeyAuth
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/bans/{user_id}/history [get]
func (h *AdminHandler) GetUserBanHistory(c *gin.Context) {
    userIDStr := c.Param("user_id")
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error: "invalid_user_id",
        })
        return
    }

    bans, err := h.adminService.GetUserBanHistory(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Error: "failed_to_fetch_ban_history",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "bans": bans,
    })
}

// GetAuditLogs godoc
// @Summary Общий лог аудита
// @Description Возвращает историю действий администраторов и модераторов
// @Tags admin
// @Security ApiKeyAuth
// @Produce json
// @Param limit query int false "Количество (default 50)"
// @Param offset query int false "Смещение (default 0)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit-logs [get]
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
    limitStr := c.DefaultQuery("limit", "50")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    logs, err := h.adminService.GetAuditLogs(c.Request.Context(), limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Error: fmt.Sprintf("failed to fetch audit logs with error %v", err.Error()),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "logs":   logs,
        "limit":  limit,
        "offset": offset,
    })
}

// GetUserAuditLogs godoc
// @Summary Лог аудита по конкретному пользователю
// @Description Возвращает все действия, совершенные над указанным пользователем
// @Tags admin
// @Security ApiKeyAuth
// @Produce json
// @Param user_id path int true "ID целевого пользователя"
// @Param limit query int false "Количество"
// @Param offset query int false "Смещение"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/audit-logs/user/{user_id} [get]
func (h *AdminHandler) GetUserAuditLogs(c *gin.Context) {
    userIDStr := c.Param("user_id")
    userID, err := strconv.ParseInt(userIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Error: "invalid_user_id",
        })
        return
    }

    limitStr := c.DefaultQuery("limit", "50")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    logs, err := h.adminService.GetUserAuditLogs(c.Request.Context(), userID, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Error: "failed_to_fetch_user_audit_logs",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "logs":   logs,
        "limit":  limit,
        "offset": offset,
    })
}

// GetDashboardStats godoc
// @Summary Статистика системы
// @Description Возвращает общую статистику по пользователям (всего, онлайн, забаненные и т.д.)
// @Tags admin
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} interface{}
// @Router /api/v1/admin/stats [get]
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
    stats, err := h.adminService.GetUserStats(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Error: "failed_to_fetch_stats",
        })
        return
    }

    c.JSON(http.StatusOK, stats)
}