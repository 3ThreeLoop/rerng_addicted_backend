package utils

import (
	"fmt"
	"os"
	"time"

	custom_log "rerng_addicted_api/pkg/logs"
	postgres "rerng_addicted_api/pkg/postgres"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func AddUserAuditLog(
	userID int,
	auditContext string,
	auditDesc string,
	auditTypeID int,
	userAgent string,
	userName string,
	ip string,
	createdBy int,
	dbPool *sqlx.DB,
) (*bool, error) {

	// 1. Get next sequence value
	seqName := "tbl_users_audits_id_seq"
	seqVal, err := postgres.GetSeqNextVal(seqName, dbPool)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch next sequence value: %w", err)
	}

	// 2. Load timezone
	appTimezone := os.Getenv("APP_TIMEZONE")
	location, err := time.LoadLocation(appTimezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	localNow := time.Now().In(location)

	// 3. Generate UUID
	auditUUID := uuid.New().String()

	// 4. Build and execute insert query
	query := `INSERT INTO tbl_users_audits (
		id, user_audit_uuid, user_id, user_audit_context, user_audit_desc,
		audit_type_id, user_agent, operator, ip, status_id,
		"order", created_by, created_at
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10,
		$11, $12, $13
	)`

	_, err = dbPool.Exec(
		query,
		*seqVal,
		auditUUID,
		userID,
		auditContext,
		auditDesc,
		auditTypeID,
		userAgent,
		userName,
		ip,
		1,       // status_id
		*seqVal, // order same as ID for now
		createdBy,
		localNow,
	)
	if err != nil {
		custom_log.NewCustomLog("user_audit_create_failed", err.Error(), "error")
		return nil, err
	}

	success := true
	return &success, nil
}
