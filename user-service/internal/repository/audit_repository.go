package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhanserikAmangeldi/user-service/internal/models"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) LogAction(ctx context.Context, log *models.AdminAuditLog) error {
	query := `
		INSERT INTO admin_audit_log (admin_id, action, target_type, target_id, details, ip_address::text)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	var detailsJSON []byte
	var err error
	if log.Details != nil {
		detailsJSON, err = json.Marshal(log.Details)
		if err != nil {
			return err
		}
	}

	err = r.db.QueryRow(ctx, query,
		log.AdminID,
		log.Action,
		log.TargetType,
		log.TargetID,
		detailsJSON,
		log.IPAddress,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

func (r *AuditRepository) GetAdminActions(ctx context.Context, limit, offset int) ([]*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_id, action, target_type, target_id, details, ip_address::text, created_at
		FROM admin_audit_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AdminAuditLog
	for rows.Next() {
		log := &models.AdminAuditLog{}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.AdminID,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&detailsJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(detailsJSON) > 0 {
			var details interface{}
			if err := json.Unmarshal(detailsJSON, &details); err == nil {
				log.Details = details
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (r *AuditRepository) GetUserActions(ctx context.Context, userID int64, limit, offset int) ([]*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_id, action, target_type, target_id, details, ip_address::text, created_at
		FROM admin_audit_log
		WHERE target_type = 'user' AND target_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AdminAuditLog
	for rows.Next() {
		log := &models.AdminAuditLog{}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.AdminID,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&detailsJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(detailsJSON) > 0 {
			var details interface{}
			if err := json.Unmarshal(detailsJSON, &details); err == nil {
				log.Details = details
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (r *AuditRepository) GetActionsByAdmin(ctx context.Context, adminID int64, limit, offset int) ([]*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_id, action, target_type, target_id, details, ip_address::text, created_at
		FROM admin_audit_log
		WHERE admin_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, adminID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AdminAuditLog
	for rows.Next() {
		log := &models.AdminAuditLog{}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.AdminID,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&detailsJSON,
			&log.IPAddress,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(detailsJSON) > 0 {
			var details interface{}
			if err := json.Unmarshal(detailsJSON, &details); err == nil {
				log.Details = details
			}
		}

		logs = append(logs, log)
	}

	return logs, nil
}
