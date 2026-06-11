package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"user_service/internal/apperror"
	handlermodel "user_service/internal/handlers/users"
)

func (s *Storage) FindWorkspacesByUser(userUUID string) ([]handlermodel.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(
		ctx,
		`
			SELECT
				w.id::text,
				w.name,
				w.owner_user_id::text,
				w.visibility,
				w.created_at,
				w.updated_at
			FROM workspaces w
			INNER JOIN workspace_members wm
				ON wm.workspace_id = w.id
			WHERE wm.user_id = $1
				AND wm.status = 'active'
			ORDER BY
				CASE WHEN w.visibility = 'private' THEN 0 ELSE 1 END,
				w.created_at ASC
		`,
		userUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return nil, apperror.BadRequestError("invalid user uuid")
		}
		return nil, fmt.Errorf("find workspaces by user: %w", err)
	}
	defer rows.Close()

	workspaces := make([]handlermodel.Workspace, 0)
	for rows.Next() {
		workspace, scanErr := scanWorkspace(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspaces: %w", err)
	}

	return workspaces, nil
}

func (s *Storage) FindWorkspaceByID(workspaceUUID string) (handlermodel.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return findWorkspaceByID(ctx, s.pool, workspaceUUID)
}

func (s *Storage) FindPersonalWorkspace(userUUID string) (handlermodel.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(
		ctx,
		`
			SELECT
				id::text,
				name,
				owner_user_id::text,
				visibility,
				created_at,
				updated_at
			FROM workspaces
			WHERE owner_user_id = $1
				AND visibility = 'private'
			ORDER BY created_at ASC
			LIMIT 1
		`,
		userUUID,
	)

	workspace, err := scanWorkspace(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handlermodel.Workspace{}, apperror.NotFoundError("personal workspace not found")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.Workspace{}, apperror.BadRequestError("invalid user uuid")
		}
		return handlermodel.Workspace{}, fmt.Errorf("find personal workspace: %w", err)
	}

	return workspace, nil
}

func (s *Storage) CreateWorkspace(dto handlermodel.CreateWorkspaceDTO) (handlermodel.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return handlermodel.Workspace{}, fmt.Errorf("begin create workspace transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var workspace handlermodel.Workspace
	workspace, err = createWorkspaceTx(ctx, tx, dto)
	if err != nil {
		return handlermodel.Workspace{}, err
	}

	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO workspace_members (workspace_id, user_id, role, status, joined_at)
			VALUES ($1, $2, 'owner', 'active', now())
		`,
		workspace.Uuid,
		dto.OwnerUserUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return handlermodel.Workspace{}, apperror.BadRequestError("invalid user uuid")
		}
		if isPostgresCode(err, "23503") {
			return handlermodel.Workspace{}, apperror.NotFoundError("user not found")
		}
		return handlermodel.Workspace{}, fmt.Errorf("insert workspace membership: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return handlermodel.Workspace{}, fmt.Errorf("commit create workspace transaction: %w", err)
	}

	return workspace, nil
}

func (s *Storage) FindWorkspaceMembers(workspaceUUID string) ([]handlermodel.WorkspaceMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := findWorkspaceByID(ctx, s.pool, workspaceUUID); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(
		ctx,
		`
			SELECT
				wm.workspace_id::text,
				wm.user_id::text,
				u.username,
				u.email,
				wm.role,
				wm.status,
				wm.joined_at,
				COALESCE(wm.invited_by::text, '')
			FROM workspace_members wm
			INNER JOIN users u
				ON u.id = wm.user_id
			WHERE wm.workspace_id = $1
				AND wm.status <> 'removed'
			ORDER BY
				CASE wm.role
					WHEN 'owner' THEN 0
					WHEN 'editor' THEN 1
					ELSE 2
				END,
				u.username ASC
		`,
		workspaceUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return nil, apperror.BadRequestError("invalid workspace uuid")
		}
		return nil, fmt.Errorf("find workspace members: %w", err)
	}
	defer rows.Close()

	members := make([]handlermodel.WorkspaceMember, 0)
	for rows.Next() {
		member, scanErr := scanWorkspaceMember(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspace members: %w", err)
	}

	return members, nil
}

func (s *Storage) UpdateWorkspaceMember(workspaceUUID, memberUserUUID string, dto handlermodel.UpdateWorkspaceMemberDTO) (handlermodel.WorkspaceMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return handlermodel.WorkspaceMember{}, fmt.Errorf("begin update workspace member transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	actorRole, err := findWorkspaceRole(ctx, tx, workspaceUUID, dto.ActorUserUUID)
	if err != nil {
		return handlermodel.WorkspaceMember{}, err
	}
	if actorRole != "owner" && actorRole != "editor" {
		return handlermodel.WorkspaceMember{}, apperror.UnauthorizedError("workspace member update is not allowed")
	}
	if dto.ActorUserUUID == memberUserUUID {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("cannot update your own workspace membership")
	}

	member, err := findWorkspaceMember(ctx, tx, workspaceUUID, memberUserUUID, true)
	if err != nil {
		return handlermodel.WorkspaceMember{}, err
	}
	if member.Role == "owner" {
		return handlermodel.WorkspaceMember{}, apperror.BadRequestError("workspace owner cannot be changed")
	}

	nextRole := member.Role
	if dto.Role != "" {
		nextRole = dto.Role
	}

	nextStatus := member.Status
	if dto.Status != "" {
		nextStatus = dto.Status
	}

	_, err = tx.Exec(
		ctx,
		`
			UPDATE workspace_members
			SET role = $3,
				status = $4,
				joined_at = CASE
					WHEN $4 = 'active' AND joined_at IS NULL THEN now()
					ELSE joined_at
				END
			WHERE workspace_id = $1
				AND user_id = $2
		`,
		workspaceUUID,
		memberUserUUID,
		nextRole,
		nextStatus,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return handlermodel.WorkspaceMember{}, apperror.BadRequestError("invalid workspace uuid or member user uuid")
		}
		return handlermodel.WorkspaceMember{}, fmt.Errorf("update workspace member: %w", err)
	}

	updatedMember, err := findWorkspaceMember(ctx, tx, workspaceUUID, memberUserUUID, false)
	if err != nil {
		return handlermodel.WorkspaceMember{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return handlermodel.WorkspaceMember{}, fmt.Errorf("commit update workspace member transaction: %w", err)
	}

	return updatedMember, nil
}

func (s *Storage) FindWorkspaceInvitesByUser(userUUID string) ([]handlermodel.WorkspaceInvite, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(
		ctx,
		`
			SELECT
				i.id::text,
				i.workspace_id::text,
				w.name,
				i.email,
				i.role,
				i.invited_by::text,
				inviter.username,
				i.expires_at,
				i.accepted_at,
				i.declined_at,
				i.created_at
			FROM workspace_invites i
			INNER JOIN users u
				ON u.id = $1
			INNER JOIN workspaces w
				ON w.id = i.workspace_id
			INNER JOIN users inviter
				ON inviter.id = i.invited_by
			WHERE lower(i.email) = lower(u.email)
				AND i.accepted_at IS NULL
				AND i.declined_at IS NULL
				AND i.expires_at > now()
			ORDER BY i.created_at DESC
		`,
		userUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return nil, apperror.BadRequestError("invalid user uuid")
		}
		return nil, fmt.Errorf("find workspace invites by user: %w", err)
	}
	defer rows.Close()

	invites := make([]handlermodel.WorkspaceInvite, 0)
	for rows.Next() {
		invite, scanErr := scanWorkspaceInvite(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		invites = append(invites, invite)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspace invites: %w", err)
	}

	return invites, nil
}

func (s *Storage) FindWorkspaceInvitesByWorkspace(workspaceUUID, actorUserUUID string) ([]handlermodel.WorkspaceInvite, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin find workspace invites transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	role, err := findWorkspaceRole(ctx, tx, workspaceUUID, actorUserUUID)
	if err != nil {
		return nil, err
	}
	if role != "owner" && role != "editor" {
		return nil, apperror.UnauthorizedError("workspace invites are not available")
	}

	rows, err := tx.Query(
		ctx,
		`
			SELECT
				i.id::text,
				i.workspace_id::text,
				w.name,
				i.email,
				i.role,
				i.invited_by::text,
				inviter.username,
				i.expires_at,
				i.accepted_at,
				i.declined_at,
				i.created_at
			FROM workspace_invites i
			INNER JOIN workspaces w
				ON w.id = i.workspace_id
			INNER JOIN users inviter
				ON inviter.id = i.invited_by
			WHERE i.workspace_id = $1
			ORDER BY i.created_at DESC
		`,
		workspaceUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return nil, apperror.BadRequestError("invalid workspace uuid")
		}
		return nil, fmt.Errorf("find workspace invites by workspace: %w", err)
	}
	defer rows.Close()

	invites := make([]handlermodel.WorkspaceInvite, 0)
	for rows.Next() {
		invite, scanErr := scanWorkspaceInvite(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		invites = append(invites, invite)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspace invites by workspace: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit find workspace invites transaction: %w", err)
	}

	return invites, nil
}

func (s *Storage) CreateWorkspaceInvite(workspaceUUID string, dto handlermodel.CreateWorkspaceInviteDTO) (handlermodel.WorkspaceInvite, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return handlermodel.WorkspaceInvite{}, fmt.Errorf("begin create workspace invite transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	role, err := findWorkspaceRole(ctx, tx, workspaceUUID, dto.InvitedByUserUUID)
	if err != nil {
		return handlermodel.WorkspaceInvite{}, err
	}
	if role != "owner" && role != "editor" {
		return handlermodel.WorkspaceInvite{}, apperror.UnauthorizedError("workspace invite is not allowed")
	}

	workspace, err := findWorkspaceByID(ctx, tx, workspaceUUID)
	if err != nil {
		return handlermodel.WorkspaceInvite{}, err
	}
	if workspace.IsPersonal {
		return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("personal workspace does not support invites")
	}

	if err = ensureWorkspaceInviteTargetAllowed(ctx, tx, workspaceUUID, dto.Email); err != nil {
		return handlermodel.WorkspaceInvite{}, err
	}

	row := tx.QueryRow(
		ctx,
		`
			INSERT INTO workspace_invites (
				workspace_id,
				email,
				role,
				invited_by,
				token_hash,
				expires_at
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				encode(gen_random_bytes(32), 'hex'),
				to_timestamp($5)
			)
			RETURNING
				id::text,
				workspace_id::text,
				(SELECT name FROM workspaces WHERE id = $1),
				email,
				role,
				invited_by::text,
				(SELECT username FROM users WHERE id = $4),
				expires_at,
				accepted_at,
				declined_at,
				created_at
		`,
		workspaceUUID,
		dto.Email,
		dto.Role,
		dto.InvitedByUserUUID,
		dto.ExpiresAt,
	)

	invite, err := scanWorkspaceInvite(row)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("invalid workspace uuid, user uuid, or expires_at")
		}
		if isPostgresCode(err, "23503") {
			return handlermodel.WorkspaceInvite{}, apperror.NotFoundError("workspace or inviter not found")
		}
		return handlermodel.WorkspaceInvite{}, fmt.Errorf("insert workspace invite: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return handlermodel.WorkspaceInvite{}, fmt.Errorf("commit create workspace invite transaction: %w", err)
	}

	return invite, nil
}

func (s *Storage) AcceptWorkspaceInvite(inviteUUID, userUUID string) (handlermodel.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return handlermodel.Workspace{}, fmt.Errorf("begin accept invite transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	invite, err := findInviteForUpdate(ctx, tx, inviteUUID)
	if err != nil {
		return handlermodel.Workspace{}, err
	}

	userEmail, err := findUserEmail(ctx, tx, userUUID)
	if err != nil {
		return handlermodel.Workspace{}, err
	}
	if invite.Email != userEmail {
		return handlermodel.Workspace{}, apperror.UnauthorizedError("workspace invite does not belong to user")
	}
	if invite.AcceptedAt != nil {
		return handlermodel.Workspace{}, apperror.BadRequestError("workspace invite already accepted")
	}
	if invite.DeclinedAt != nil {
		return handlermodel.Workspace{}, apperror.BadRequestError("workspace invite already declined")
	}
	if invite.ExpiresAt <= time.Now().Unix() {
		return handlermodel.Workspace{}, apperror.BadRequestError("workspace invite expired")
	}

	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO workspace_members (workspace_id, user_id, role, status, joined_at, invited_by)
			VALUES ($1, $2, $3, 'active', now(), NULLIF($4, '')::uuid)
			ON CONFLICT (workspace_id, user_id)
			DO UPDATE SET
				role = EXCLUDED.role,
				status = 'active',
				joined_at = COALESCE(workspace_members.joined_at, now()),
				invited_by = COALESCE(workspace_members.invited_by, EXCLUDED.invited_by)
		`,
		invite.WorkspaceUUID,
		userUUID,
		invite.Role,
		invite.InvitedByUserUUID,
	)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return handlermodel.Workspace{}, apperror.BadRequestError("invalid user uuid")
		}
		if isPostgresCode(err, "23503") {
			return handlermodel.Workspace{}, apperror.NotFoundError("workspace or user not found")
		}
		return handlermodel.Workspace{}, fmt.Errorf("upsert workspace membership: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
			UPDATE workspace_invites
			SET accepted_at = now(),
				declined_at = NULL
			WHERE id = $1
		`,
		inviteUUID,
	)
	if err != nil {
		return handlermodel.Workspace{}, fmt.Errorf("mark workspace invite accepted: %w", err)
	}

	workspace, err := findWorkspaceByID(ctx, tx, invite.WorkspaceUUID)
	if err != nil {
		return handlermodel.Workspace{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return handlermodel.Workspace{}, fmt.Errorf("commit accept invite transaction: %w", err)
	}

	return workspace, nil
}

func (s *Storage) DeclineWorkspaceInvite(inviteUUID, userUUID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin decline invite transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	invite, err := findInviteForUpdate(ctx, tx, inviteUUID)
	if err != nil {
		return err
	}

	userEmail, err := findUserEmail(ctx, tx, userUUID)
	if err != nil {
		return err
	}
	if invite.Email != userEmail {
		return apperror.UnauthorizedError("workspace invite does not belong to user")
	}
	if invite.AcceptedAt != nil {
		return apperror.BadRequestError("workspace invite already accepted")
	}
	if invite.DeclinedAt != nil {
		return apperror.BadRequestError("workspace invite already declined")
	}

	_, err = tx.Exec(
		ctx,
		`
			UPDATE workspace_invites
			SET declined_at = now()
			WHERE id = $1
		`,
		inviteUUID,
	)
	if err != nil {
		return fmt.Errorf("decline workspace invite: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit decline invite transaction: %w", err)
	}

	return nil
}

func (s *Storage) FindWorkspaceAccess(workspaceUUID, userUUID string) (handlermodel.WorkspaceAccess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.pool.QueryRow(
		ctx,
		`
			SELECT
				w.id::text,
				w.name,
				w.owner_user_id::text,
				w.visibility,
				w.created_at,
				w.updated_at,
				wm.role,
				wm.status
			FROM workspaces w
			INNER JOIN workspace_members wm
				ON wm.workspace_id = w.id
			WHERE w.id = $1
				AND wm.user_id = $2
				AND wm.status = 'active'
			LIMIT 1
		`,
		workspaceUUID,
		userUUID,
	)

	var access handlermodel.WorkspaceAccess
	var workspaceCreatedAt time.Time
	var workspaceUpdatedAt time.Time
	err := row.Scan(
		&access.Workspace.Uuid,
		&access.Workspace.Name,
		&access.Workspace.OwnerUserUUID,
		&access.Workspace.Visibility,
		&workspaceCreatedAt,
		&workspaceUpdatedAt,
		&access.Role,
		&access.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists, existsErr := workspaceExists(ctx, s.pool, workspaceUUID)
			if existsErr != nil {
				return handlermodel.WorkspaceAccess{}, existsErr
			}
			if !exists {
				return handlermodel.WorkspaceAccess{}, apperror.NotFoundError("workspace not found")
			}
			return handlermodel.WorkspaceAccess{}, apperror.UnauthorizedError("workspace access denied")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.WorkspaceAccess{}, apperror.BadRequestError("invalid workspace uuid or user uuid")
		}
		return handlermodel.WorkspaceAccess{}, fmt.Errorf("find workspace access: %w", err)
	}

	access.UserUUID = userUUID
	access.Allowed = true
	access.Workspace.IsPersonal = access.Workspace.Visibility == "private"
	access.Workspace.CreatedAt = workspaceCreatedAt.Unix()
	access.Workspace.UpdatedAt = workspaceUpdatedAt.Unix()

	return access, nil
}

func createWorkspaceTx(ctx context.Context, tx pgx.Tx, dto handlermodel.CreateWorkspaceDTO) (handlermodel.Workspace, error) {
	row := tx.QueryRow(
		ctx,
		`
			INSERT INTO workspaces (name, owner_user_id, visibility)
			VALUES ($1, $2, $3)
			RETURNING
				id::text,
				name,
				owner_user_id::text,
				visibility,
				created_at,
				updated_at
		`,
		dto.Name,
		dto.OwnerUserUUID,
		dto.Visibility,
	)

	workspace, err := scanWorkspace(row)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return handlermodel.Workspace{}, apperror.BadRequestError("invalid owner user uuid")
		}
		if isPostgresCode(err, "23503") {
			return handlermodel.Workspace{}, apperror.NotFoundError("user not found")
		}
		return handlermodel.Workspace{}, fmt.Errorf("insert workspace: %w", err)
	}

	return workspace, nil
}

func findWorkspaceByID(ctx context.Context, queryRow interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, workspaceUUID string) (handlermodel.Workspace, error) {
	row := queryRow.QueryRow(
		ctx,
		`
			SELECT
				id::text,
				name,
				owner_user_id::text,
				visibility,
				created_at,
				updated_at
			FROM workspaces
			WHERE id = $1
			LIMIT 1
		`,
		workspaceUUID,
	)

	workspace, err := scanWorkspace(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handlermodel.Workspace{}, apperror.NotFoundError("workspace not found")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.Workspace{}, apperror.BadRequestError("invalid workspace uuid")
		}
		return handlermodel.Workspace{}, fmt.Errorf("find workspace: %w", err)
	}

	return workspace, nil
}

func findWorkspaceRole(ctx context.Context, tx pgx.Tx, workspaceUUID, userUUID string) (string, error) {
	var role string
	err := tx.QueryRow(
		ctx,
		`
			SELECT role
			FROM workspace_members
			WHERE workspace_id = $1
				AND user_id = $2
				AND status = 'active'
			LIMIT 1
		`,
		workspaceUUID,
		userUUID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists, existsErr := workspaceExists(ctx, tx, workspaceUUID)
			if existsErr != nil {
				return "", existsErr
			}
			if !exists {
				return "", apperror.NotFoundError("workspace not found")
			}
			return "", apperror.UnauthorizedError("workspace access denied")
		}
		if isPostgresCode(err, "22P02") {
			return "", apperror.BadRequestError("invalid workspace uuid or user uuid")
		}
		return "", fmt.Errorf("find workspace role: %w", err)
	}

	return role, nil
}

func findWorkspaceMember(
	ctx context.Context,
	queryRow interface {
		QueryRow(context.Context, string, ...any) pgx.Row
	},
	workspaceUUID,
	memberUserUUID string,
	forUpdate bool,
) (handlermodel.WorkspaceMember, error) {
	query := `
		SELECT
			wm.workspace_id::text,
			wm.user_id::text,
			u.username,
			u.email,
			wm.role,
			wm.status,
			wm.joined_at,
			COALESCE(wm.invited_by::text, '')
		FROM workspace_members wm
		INNER JOIN users u
			ON u.id = wm.user_id
		WHERE wm.workspace_id = $1
			AND wm.user_id = $2
		LIMIT 1
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	row := queryRow.QueryRow(ctx, query, workspaceUUID, memberUserUUID)

	member, err := scanWorkspaceMember(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists, existsErr := workspaceExists(ctx, queryRow, workspaceUUID)
			if existsErr != nil {
				return handlermodel.WorkspaceMember{}, existsErr
			}
			if !exists {
				return handlermodel.WorkspaceMember{}, apperror.NotFoundError("workspace not found")
			}
			return handlermodel.WorkspaceMember{}, apperror.NotFoundError("workspace member not found")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.WorkspaceMember{}, apperror.BadRequestError("invalid workspace uuid or member user uuid")
		}
		return handlermodel.WorkspaceMember{}, fmt.Errorf("find workspace member: %w", err)
	}

	return member, nil
}

func ensureWorkspaceInviteTargetAllowed(ctx context.Context, tx pgx.Tx, workspaceUUID, email string) error {
	var activeMemberCount int
	err := tx.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM workspace_members wm
			INNER JOIN users u
				ON u.id = wm.user_id
			WHERE wm.workspace_id = $1
				AND wm.status = 'active'
				AND lower(u.email) = lower($2)
		`,
		workspaceUUID,
		email,
	).Scan(&activeMemberCount)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return apperror.BadRequestError("invalid workspace uuid")
		}
		return fmt.Errorf("count active workspace members by email: %w", err)
	}
	if activeMemberCount > 0 {
		return apperror.BadRequestError("user is already a workspace member")
	}

	var pendingInviteCount int
	err = tx.QueryRow(
		ctx,
		`
			SELECT count(*)
			FROM workspace_invites
			WHERE workspace_id = $1
				AND lower(email) = lower($2)
				AND accepted_at IS NULL
				AND declined_at IS NULL
				AND expires_at > now()
		`,
		workspaceUUID,
		email,
	).Scan(&pendingInviteCount)
	if err != nil {
		return fmt.Errorf("count pending workspace invites: %w", err)
	}
	if pendingInviteCount > 0 {
		return apperror.BadRequestError("active workspace invite already exists")
	}

	return nil
}

func findInviteForUpdate(ctx context.Context, tx pgx.Tx, inviteUUID string) (handlermodel.WorkspaceInvite, error) {
	row := tx.QueryRow(
		ctx,
		`
			SELECT
				i.id::text,
				i.workspace_id::text,
				w.name,
				i.email,
				i.role,
				i.invited_by::text,
				inviter.username,
				i.expires_at,
				i.accepted_at,
				i.declined_at,
				i.created_at
			FROM workspace_invites i
			INNER JOIN workspaces w
				ON w.id = i.workspace_id
			INNER JOIN users inviter
				ON inviter.id = i.invited_by
			WHERE i.id = $1
			FOR UPDATE
		`,
		inviteUUID,
	)

	invite, err := scanWorkspaceInvite(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handlermodel.WorkspaceInvite{}, apperror.NotFoundError("workspace invite not found")
		}
		if isPostgresCode(err, "22P02") {
			return handlermodel.WorkspaceInvite{}, apperror.BadRequestError("invalid invite uuid")
		}
		return handlermodel.WorkspaceInvite{}, fmt.Errorf("find workspace invite: %w", err)
	}

	return invite, nil
}

func findUserEmail(ctx context.Context, tx pgx.Tx, userUUID string) (string, error) {
	var email string
	err := tx.QueryRow(
		ctx,
		`
			SELECT lower(email)
			FROM users
			WHERE id = $1
				AND status = 'active'
			LIMIT 1
		`,
		userUUID,
	).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperror.NotFoundError("user not found")
		}
		if isPostgresCode(err, "22P02") {
			return "", apperror.BadRequestError("invalid user uuid")
		}
		return "", fmt.Errorf("find user email: %w", err)
	}

	return email, nil
}

func workspaceExists(ctx context.Context, queryRow interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, workspaceUUID string) (bool, error) {
	var exists bool
	err := queryRow.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM workspaces WHERE id = $1)`,
		workspaceUUID,
	).Scan(&exists)
	if err != nil {
		if isPostgresCode(err, "22P02") {
			return false, apperror.BadRequestError("invalid workspace uuid")
		}
		return false, fmt.Errorf("check workspace exists: %w", err)
	}

	return exists, nil
}

func scanWorkspace(scanner interface {
	Scan(...any) error
}) (handlermodel.Workspace, error) {
	var workspace handlermodel.Workspace
	var createdAt time.Time
	var updatedAt time.Time

	err := scanner.Scan(
		&workspace.Uuid,
		&workspace.Name,
		&workspace.OwnerUserUUID,
		&workspace.Visibility,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return handlermodel.Workspace{}, err
	}

	workspace.IsPersonal = workspace.Visibility == "private"
	workspace.CreatedAt = createdAt.Unix()
	workspace.UpdatedAt = updatedAt.Unix()

	return workspace, nil
}

func scanWorkspaceMember(scanner interface {
	Scan(...any) error
}) (handlermodel.WorkspaceMember, error) {
	var member handlermodel.WorkspaceMember
	var joinedAt sql.NullTime

	err := scanner.Scan(
		&member.WorkspaceUUID,
		&member.UserUUID,
		&member.Username,
		&member.Email,
		&member.Role,
		&member.Status,
		&joinedAt,
		&member.InvitedBy,
	)
	if err != nil {
		return handlermodel.WorkspaceMember{}, fmt.Errorf("scan workspace member: %w", err)
	}

	if joinedAt.Valid {
		value := joinedAt.Time.Unix()
		member.JoinedAt = &value
	}

	return member, nil
}

func scanWorkspaceInvite(scanner interface {
	Scan(...any) error
}) (handlermodel.WorkspaceInvite, error) {
	var invite handlermodel.WorkspaceInvite
	var expiresAt time.Time
	var acceptedAt sql.NullTime
	var declinedAt sql.NullTime
	var createdAt time.Time

	err := scanner.Scan(
		&invite.Uuid,
		&invite.WorkspaceUUID,
		&invite.WorkspaceName,
		&invite.Email,
		&invite.Role,
		&invite.InvitedByUserUUID,
		&invite.InvitedByUsername,
		&expiresAt,
		&acceptedAt,
		&declinedAt,
		&createdAt,
	)
	if err != nil {
		return handlermodel.WorkspaceInvite{}, fmt.Errorf("scan workspace invite: %w", err)
	}

	invite.ExpiresAt = expiresAt.Unix()
	invite.CreatedAt = createdAt.Unix()
	if acceptedAt.Valid {
		value := acceptedAt.Time.Unix()
		invite.AcceptedAt = &value
	}
	if declinedAt.Valid {
		value := declinedAt.Time.Unix()
		invite.DeclinedAt = &value
	}
	switch {
	case invite.AcceptedAt != nil:
		invite.Status = "accepted"
	case invite.DeclinedAt != nil:
		invite.Status = "declined"
	case invite.ExpiresAt <= time.Now().Unix():
		invite.Status = "expired"
	default:
		invite.Status = "pending"
	}

	return invite, nil
}
