package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"user_service/internal/apperror"
	"user_service/internal/config"
	handlermodel "user_service/internal/handlers/users"
	"user_service/pkg/logging"
)

type Storage struct {
	pool *pgxpool.Pool
}

const (
	findUserByIDQuery = `
		SELECT id::text, username, email, password_hash, created_at
		FROM users
		WHERE id = $1 AND status = 'active'
		LIMIT 1
	`

	findUserByUsernameQuery = `
		SELECT id::text, username, email, password_hash, created_at
		FROM users
		WHERE username = $1 AND status = 'active'
		LIMIT 1
	`

	findUserProfileQuery = `
		SELECT id::text, username, email, created_at, last_login_at
		FROM users
		WHERE id = $1 AND status = 'active'
		LIMIT 1
	`

	findActiveSessionByRefreshTokenHashQuery = `
		SELECT
			id::text,
			user_id::text,
			refresh_token_hash,
			COALESCE(user_agent, ''),
			COALESCE(ip_address::text, ''),
			created_at,
			expires_at,
			last_used_at
		FROM user_sessions
		WHERE refresh_token_hash = $1
			AND revoked_at IS NULL
			AND (expires_at IS NULL OR expires_at > now())
		LIMIT 1
	`
)

func NewStorage(cfg *config.Config) (storage *Storage, err error) {
	logger := logging.GetLogger().With(
		"layer", "storage",
		"db", "postgres",
		"host", cfg.Postgres.Host,
		"port", cfg.Postgres.Port,
		"database", cfg.Postgres.Database,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(postgresDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("failed to connect postgres", "error", err)
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	storage = &Storage{pool: pool}
	if err = storage.Ping(); err != nil {
		pool.Close()
		logger.Error("postgres ping failed", "error", err)
		return nil, err
	}

	logger.Info("postgres connected")
	return storage, nil
}

func (s *Storage) Ping() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) Create(user handlermodel.User) (userUUID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin create user transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO users (username, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id::text
		`,
		user.Username,
		user.Email,
		user.PasswordHash,
	).Scan(&userUUID)
	if err != nil {
		if isPostgresCode(err, "23505") {
			return "", apperror.BadRequestError("user already exists")
		}
		return "", fmt.Errorf("insert user: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO user_profiles (user_id)
			VALUES ($1)
			ON CONFLICT (user_id) DO NOTHING
		`,
		userUUID,
	)
	if err != nil {
		return "", fmt.Errorf("insert user profile: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO user_preferences (user_id)
			VALUES ($1)
			ON CONFLICT (user_id) DO NOTHING
		`,
		userUUID,
	)
	if err != nil {
		return "", fmt.Errorf("insert user preferences: %w", err)
	}

	var personalWorkspaceUUID string
	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO workspaces (name, owner_user_id, visibility)
			VALUES ('Личное пространство', $1, 'private')
			RETURNING id::text
		`,
		userUUID,
	).Scan(&personalWorkspaceUUID)
	if err != nil {
		return "", fmt.Errorf("insert personal workspace: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO workspace_members (workspace_id, user_id, role, status, joined_at)
			VALUES ($1, $2, 'owner', 'active', now())
		`,
		personalWorkspaceUUID,
		userUUID,
	)
	if err != nil {
		return "", fmt.Errorf("insert personal workspace membership: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit create user transaction: %w", err)
	}

	return userUUID, nil
}

func (s *Storage) FindOne(userUUID string) (user handlermodel.User, err error) {
	return s.findOne(findUserByIDQuery, userUUID)
}

func (s *Storage) FindProfile(userUUID string) (profile handlermodel.UserProfile, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var createdAt time.Time
	var lastLoginAt sql.NullTime
	err = s.pool.QueryRow(ctx, findUserProfileQuery, userUUID).Scan(
		&profile.Uuid,
		&profile.Username,
		&profile.Email,
		&createdAt,
		&lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handlermodel.UserProfile{}, apperror.NotFoundError("user not found")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.UserProfile{}, apperror.BadRequestError("invalid user uuid")
		}
		return handlermodel.UserProfile{}, fmt.Errorf("find user profile: %w", err)
	}

	profile.CreatedAt = createdAt.Unix()
	if lastLoginAt.Valid {
		value := lastLoginAt.Time.Unix()
		profile.LastLoginAt = &value
	}

	return profile, nil
}

func (s *Storage) FindByUsername(username string) (user handlermodel.User, err error) {
	return s.findOne(findUserByUsernameQuery, username)
}

func (s *Storage) UpdateProfile(userUUID, username, email, passwordHash string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE users
		SET
			username = $2,
			email = $3,
			password_hash = CASE WHEN $4 <> '' THEN $4 ELSE password_hash END
		WHERE id = $1 AND status = 'active'
	`

	result, err := s.pool.Exec(ctx, query, userUUID, username, email, passwordHash)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return apperror.BadRequestError("invalid user uuid")
		}
		if isPostgresCode(err, "23505") {
			return apperror.BadRequestError("username or email already exists")
		}
		return fmt.Errorf("update user profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFoundError("user not found")
	}

	return nil
}

func (s *Storage) UpdateLastLogin(userUUID string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.pool.Exec(
		ctx,
		`
			UPDATE users
			SET last_login_at = now()
			WHERE id = $1 AND status = 'active'
		`,
		userUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return apperror.BadRequestError("invalid user uuid")
		}
		return fmt.Errorf("update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFoundError("user not found")
	}

	return nil
}

func (s *Storage) CreateAction(userUUID string, dto handlermodel.CreateUserActionDTO) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata := dto.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal action metadata: %w", err)
	}

	_, err = s.pool.Exec(
		ctx,
		`
			INSERT INTO user_actions (
				user_id,
				action,
				entity_type,
				entity_id,
				status,
				metadata,
				ip_address,
				user_agent
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb, NULLIF($7, '')::inet, NULLIF($8, ''))
		`,
		userUUID,
		dto.Action,
		dto.EntityType,
		dto.EntityId,
		dto.Status,
		string(metadataBytes),
		dto.IPAddress,
		dto.UserAgent,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return apperror.BadRequestError("invalid user uuid or ip address")
		}
		if isPostgresCode(err, "23503") {
			return apperror.NotFoundError("user not found")
		}
		return fmt.Errorf("insert user action: %w", err)
	}

	return nil
}

func (s *Storage) FindActions(userUUID string, limit, offset int) (actions []handlermodel.UserAction, err error) {
	actions = make([]handlermodel.UserAction, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(
		ctx,
		`
			SELECT
				id::text,
				action,
				entity_type,
				COALESCE(entity_id, ''),
				status,
				metadata,
				COALESCE(ip_address::text, ''),
				COALESCE(user_agent, ''),
				created_at
			FROM user_actions
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`,
		userUUID,
		limit,
		offset,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return nil, apperror.BadRequestError("invalid user uuid")
		}
		return nil, fmt.Errorf("find user actions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action handlermodel.UserAction
		var metadataBytes []byte
		var createdAt time.Time
		if err = rows.Scan(
			&action.Uuid,
			&action.Action,
			&action.EntityType,
			&action.EntityId,
			&action.Status,
			&metadataBytes,
			&action.IPAddress,
			&action.UserAgent,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan user action: %w", err)
		}

		if len(metadataBytes) > 0 {
			if err = json.Unmarshal(metadataBytes, &action.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal action metadata: %w", err)
			}
		}
		if action.Metadata == nil {
			action.Metadata = map[string]any{}
		}
		action.CreatedAt = createdAt.Unix()
		actions = append(actions, action)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user actions: %w", err)
	}

	return actions, nil
}

func (s *Storage) CreateSession(dto handlermodel.CreateUserSessionDTO) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.pool.Exec(
		ctx,
		`
			INSERT INTO user_sessions (
				user_id,
				refresh_token_hash,
				user_agent,
				ip_address,
				expires_at
			)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, '')::inet, to_timestamp($5))
		`,
		dto.UserUUID,
		dto.RefreshTokenHash,
		dto.UserAgent,
		dto.IPAddress,
		dto.ExpiresAt,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return apperror.BadRequestError("invalid user uuid or ip address")
		}
		if isPostgresCode(err, "23503") {
			return apperror.NotFoundError("user not found")
		}
		if isPostgresCode(err, "23505") {
			return apperror.BadRequestError("refresh session already exists")
		}
		return fmt.Errorf("insert user session: %w", err)
	}

	return nil
}

func (s *Storage) RotateSession(dto handlermodel.RotateUserSessionDTO) (handlermodel.UserSession, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return handlermodel.UserSession{}, fmt.Errorf("begin rotate session transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	currentSession, err := findSessionForUpdate(ctx, tx, dto.RefreshTokenHash)
	if err != nil {
		return handlermodel.UserSession{}, err
	}

	_, err = tx.Exec(
		ctx,
		`
			UPDATE user_sessions
			SET revoked_at = now(), last_used_at = now()
			WHERE id = $1
		`,
		currentSession.Uuid,
	)
	if err != nil {
		return handlermodel.UserSession{}, fmt.Errorf("revoke current session: %w", err)
	}

	var newSession handlermodel.UserSession
	var createdAt time.Time
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	err = tx.QueryRow(
		ctx,
		`
			INSERT INTO user_sessions (
				user_id,
				refresh_token_hash,
				user_agent,
				ip_address,
				expires_at,
				last_used_at
			)
			VALUES (
				$1,
				$2,
				NULLIF($3, ''),
				NULLIF($4, '')::inet,
				to_timestamp($5),
				now()
			)
			RETURNING id::text, user_id::text, COALESCE(user_agent, ''), COALESCE(ip_address::text, ''), created_at, expires_at, last_used_at
		`,
		currentSession.UserUUID,
		dto.NewRefreshTokenHash,
		dto.UserAgent,
		dto.IPAddress,
		dto.ExpiresAt,
	).Scan(
		&newSession.Uuid,
		&newSession.UserUUID,
		&newSession.UserAgent,
		&newSession.IPAddress,
		&createdAt,
		&expiresAt,
		&lastUsedAt,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return handlermodel.UserSession{}, apperror.BadRequestError("invalid ip address")
		}
		if isPostgresCode(err, "23505") {
			return handlermodel.UserSession{}, apperror.BadRequestError("refresh session already exists")
		}
		return handlermodel.UserSession{}, fmt.Errorf("insert rotated user session: %w", err)
	}

	newSession.CreatedAt = createdAt.Unix()
	if expiresAt.Valid {
		newSession.ExpiresAt = expiresAt.Time.Unix()
	}
	if lastUsedAt.Valid {
		value := lastUsedAt.Time.Unix()
		newSession.LastUsedAt = &value
	}

	if err = tx.Commit(ctx); err != nil {
		return handlermodel.UserSession{}, fmt.Errorf("commit rotate session transaction: %w", err)
	}

	return newSession, nil
}

func (s *Storage) RevokeSession(refreshTokenHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.pool.Exec(
		ctx,
		`
			UPDATE user_sessions
			SET revoked_at = now(), last_used_at = now()
			WHERE refresh_token_hash = $1 AND revoked_at IS NULL
		`,
		refreshTokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke user session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.NotFoundError("session not found")
	}

	return nil
}

func (s *Storage) findOne(query string, arg any) (user handlermodel.User, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var createdAt time.Time
	err = s.pool.QueryRow(ctx, query, arg).Scan(
		&user.Uuid,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handlermodel.User{}, apperror.NotFoundError("user not found")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.User{}, apperror.BadRequestError("invalid user uuid")
		}
		return handlermodel.User{}, fmt.Errorf("find user: %w", err)
	}

	user.CreatedDate = createdAt.Unix()
	return user, nil
}

func findSessionForUpdate(ctx context.Context, tx pgx.Tx, refreshTokenHash string) (handlermodel.UserSession, error) {
	var session handlermodel.UserSession
	var createdAt time.Time
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := tx.QueryRow(
		ctx,
		findActiveSessionByRefreshTokenHashQuery+` FOR UPDATE`,
		refreshTokenHash,
	).Scan(
		&session.Uuid,
		&session.UserUUID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&createdAt,
		&expiresAt,
		&lastUsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handlermodel.UserSession{}, apperror.NotFoundError("session not found")
		}
		return handlermodel.UserSession{}, fmt.Errorf("find session: %w", err)
	}

	session.CreatedAt = createdAt.Unix()
	if expiresAt.Valid {
		session.ExpiresAt = expiresAt.Time.Unix()
	}
	if lastUsedAt.Valid {
		value := lastUsedAt.Time.Unix()
		session.LastUsedAt = &value
	}

	return session, nil
}

func postgresDSN(cfg *config.Config) string {
	postgresURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Postgres.Username, cfg.Postgres.Password),
		Host:   net.JoinHostPort(cfg.Postgres.Host, cfg.Postgres.Port),
		Path:   cfg.Postgres.Database,
	}

	query := postgresURL.Query()
	query.Set("sslmode", cfg.Postgres.SSLMode)
	postgresURL.RawQuery = query.Encode()

	return postgresURL.String()
}

func isPostgresCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
