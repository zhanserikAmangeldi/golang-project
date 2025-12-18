package service

import (
	"context"
	"errors"
	"time"

	"github.com/zhanserikAmangeldi/user-service/internal/models"
	"github.com/zhanserikAmangeldi/user-service/internal/repository"
)

type AdminService struct {
	userRepo  *repository.UserRepository
	banRepo   *repository.BanRepository
	auditRepo *repository.AuditRepository
}

func NewAdminService(
	userRepo *repository.UserRepository,
	banRepo *repository.BanRepository,
	auditRepo *repository.AuditRepository,
) *AdminService {
	return &AdminService{
		userRepo:  userRepo,
		banRepo:   banRepo,
		auditRepo: auditRepo,
	}
}

func (s *AdminService) BanUser(ctx context.Context, adminID, userID int64, reason string, duration *time.Duration, ipAddress *string) error {
	if adminID == userID {
		return errors.New("cannot ban yourself")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IsAdmin() {
		return errors.New("cannot ban admin users")
	}

	ban := &models.UserBan{
		UserID:   userID,
		BannedBy: adminID,
		Reason:   reason,
	}

	if duration != nil {
		expiresAt := time.Now().Add(*duration)
		ban.ExpiresAt = &expiresAt
		ban.IsPermanent = false
	} else {
		ban.IsPermanent = true
	}

	err = s.banRepo.BanUser(ctx, ban)
	if err != nil {
		return err
	}

	details := map[string]interface{}{
		"reason":       reason,
		"is_permanent": ban.IsPermanent,
	}
	if ban.ExpiresAt != nil {
		details["expires_at"] = ban.ExpiresAt
	}

	_ = s.auditRepo.LogAction(ctx, &models.AdminAuditLog{
		AdminID:    adminID,
		Action:     "ban_user",
		TargetType: "user",
		TargetID:   &userID,
		Details:    details,
		IPAddress:  ipAddress,
	})

	return nil
}

func (s *AdminService) UnbanUser(ctx context.Context, adminID, userID int64, ipAddress *string) error {
	err := s.banRepo.UnbanUser(ctx, userID, adminID)
	if err != nil {
		return err
	}

	_ = s.auditRepo.LogAction(ctx, &models.AdminAuditLog{
		AdminID:    adminID,
		Action:     "unban_user",
		TargetType: "user",
		TargetID:   &userID,
		IPAddress:  ipAddress,
	})

	return nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, adminID, userID int64, newRole string, ipAddress *string) error {
	validRoles := []string{models.RoleUser, models.RoleAdmin, models.RoleModerator}
	isValid := false
	for _, role := range validRoles {
		if newRole == role {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("invalid role")
	}

	if adminID == userID {
		return errors.New("cannot change your own role")
	}

	err := s.userRepo.UpdateRole(ctx, userID, newRole)
	if err != nil {
		return err
	}

	_ = s.auditRepo.LogAction(ctx, &models.AdminAuditLog{
		AdminID:    adminID,
		Action:     "update_user_role",
		TargetType: "user",
		TargetID:   &userID,
		Details:    map[string]interface{}{"new_role": newRole},
		IPAddress:  ipAddress,
	})

	return nil
}

func (s *AdminService) DeleteUser(ctx context.Context, adminID, userID int64, ipAddress *string) error {
	if adminID == userID {
		return errors.New("cannot delete yourself")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IsAdmin() {
		return errors.New("cannot delete admin users")
	}

	err = s.userRepo.SoftDeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	_ = s.auditRepo.LogAction(ctx, &models.AdminAuditLog{
		AdminID:    adminID,
		Action:     "delete_user",
		TargetType: "user",
		TargetID:   &userID,
		IPAddress:  ipAddress,
	})

	return nil
}

func (s *AdminService) GetAllUsers(ctx context.Context, limit, offset int, role string) ([]*models.User, int64, error) {
	users, err := s.userRepo.GetAllUsers(ctx, limit, offset, role)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.CountUsers(ctx, role)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *AdminService) GetUserBanHistory(ctx context.Context, userID int64) ([]*models.UserBan, error) {
	return s.banRepo.GetUserBanHistory(ctx, userID)
}

func (s *AdminService) GetAuditLogs(ctx context.Context, limit, offset int) ([]*models.AdminAuditLog, error) {
	return s.auditRepo.GetAdminActions(ctx, limit, offset)
}

func (s *AdminService) GetUserAuditLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.AdminAuditLog, error) {
	return s.auditRepo.GetUserActions(ctx, userID, limit, offset)
}

func (s *AdminService) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	return s.userRepo.GetUserStats(ctx)
}
