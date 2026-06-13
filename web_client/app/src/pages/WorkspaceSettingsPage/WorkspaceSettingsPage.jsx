import { useState } from "react";

import WorkspaceContextMenu from "../../components/WorkspaceContextMenu";

export default function WorkspaceSettingsPage({
  currentWorkspace,
  workspaces,
  overview,
  members,
  invites,
  memberDrafts,
  loading,
  inviteForm,
  onMemberDraftChange,
  onSubmitMemberUpdate,
  onInviteFormChange,
  onSubmitInvite,
  onLeaveWorkspace,
  onDeleteWorkspace,
  onOpenNotes,
  onOpenGraph,
  onOpenCalendar,
  onSelectPersonalWorkspace,
  onSelectWorkspace,
  onCreateWorkspace,
}) {
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const workspace = overview?.workspace || currentWorkspace || {};
  const stats = overview?.stats || {};
  const canInvite = Boolean(overview?.can_invite);
  const canManageMembers = Boolean(overview?.can_manage_members ?? overview?.can_invite);
  const isOwner = overview?.role === "owner";
  const canLeaveWorkspace = !isOwner && overview?.status === "active";

  return (
    <main className="profile-page workspace-settings-page">
      <header className="profile-header page-header">
        <div className="page-header-copy">
          <div className="page-header-leading">
            <WorkspaceContextMenu
              currentWorkspace={currentWorkspace}
              workspaces={workspaces}
              onSelectPersonalWorkspace={onSelectPersonalWorkspace}
              onSelectWorkspace={onSelectWorkspace}
              onCreateWorkspace={onCreateWorkspace}
            />
            <span className="eyebrow">Общее пространство</span>
          </div>
          <h1>{workspace.name || "Пространство"}</h1>
        </div>
        <div className="profile-actions page-header-actions">
          <button className="secondary-button" type="button" onClick={onOpenNotes}>
            Заметки
          </button>
          <button className="secondary-button" type="button" onClick={onOpenGraph}>
            Граф
          </button>
          <button className="secondary-button" type="button" onClick={onOpenCalendar}>
            Календарь
          </button>
        </div>
      </header>

      <section className="profile-grid workspace-summary-grid">
        <article className="profile-card">
          <div className="card-label">О пространстве</div>
          <h2>{workspace.name || "Без названия"}</h2>
          <dl className="profile-details">
            <div>
              <dt>Дата создания</dt>
              <dd>{formatDate(workspace.created_at)}</dd>
            </div>
            <div>
              <dt>Участников</dt>
              <dd>{Number(overview?.members_count || members.length || 0)}</dd>
            </div>
          </dl>
        </article>

        <article className="profile-card">
          <div className="card-label">Моя роль</div>
          <h2>{formatRole(overview?.role)}</h2>
          <dl className="profile-details workspace-role-details">
            <div>
              <dt>Статус</dt>
              <dd>{formatMembershipStatus(overview?.status)}</dd>
            </div>
            <div>
              <dt>Приглашения</dt>
              <dd>{canInvite ? "Можно отправлять" : "Недоступно"}</dd>
            </div>
          </dl>
          <div className="workspace-role-actions">
            {canInvite && (
              <button className="primary-button" type="button" onClick={() => setIsInviteDialogOpen(true)}>
                Пригласить
              </button>
            )}
            {canLeaveWorkspace && (
              <button className="danger-button" type="button" onClick={onLeaveWorkspace} disabled={loading}>
                Выйти из пространства
              </button>
            )}
            {isOwner && (
              <button className="danger-button" type="button" onClick={onDeleteWorkspace} disabled={loading}>
                Удалить пространство
              </button>
            )}
          </div>
        </article>
      </section>

      <section className="stats-grid workspace-stats-grid">
        <StatCard label="Категории" value={stats.categories_count} />
        <StatCard label="Заметки" value={stats.notes_count} />
        <StatCard label="Теги" value={stats.tags_count} />
        <StatCard label="Файлы" value={stats.files_count} />
      </section>

      <section className="profile-card action-history">
        <div className="section-heading">
          <div>
            <div className="card-label">Участники</div>
            <h2>Команда пространства</h2>
          </div>
          {loading && <span className="muted-text">Загрузка...</span>}
        </div>

        <div className="workspace-members-list">
          {members.map((member) => (
            <article className="workspace-member-card" key={member.user_uuid || member.email}>
              <div className="workspace-member-avatar">{initials(member.username || member.email)}</div>
              <div>
                <strong>{member.username || member.email || "Участник"}</strong>
                <span>{member.email || "email не указан"}</span>
                {canManageMembers && member.role !== "owner" && (
                  <div className="workspace-member-controls">
                    <select
                      value={memberDrafts?.[member.user_uuid]?.role || member.role || "viewer"}
                      onChange={(event) => onMemberDraftChange(member.user_uuid, { role: event.target.value })}
                      disabled={loading}
                    >
                      <option value="viewer">Участник</option>
                      <option value="editor">Креатор</option>
                    </select>
                    <select
                      value={memberDrafts?.[member.user_uuid]?.status || member.status || "active"}
                      onChange={(event) => onMemberDraftChange(member.user_uuid, { status: event.target.value })}
                      disabled={loading}
                    >
                      <option value="active">Активен</option>
                      <option value="removed">Удалить из пространства</option>
                    </select>
                    <button
                      className="secondary-button"
                      type="button"
                      onClick={() => onSubmitMemberUpdate(member.user_uuid)}
                      disabled={loading || !hasMemberChanges(member, memberDrafts?.[member.user_uuid])}
                    >
                      Сохранить
                    </button>
                  </div>
                )}
              </div>
              <div className="workspace-member-meta">
                <span className={`workspace-role-chip role-${member.role || "viewer"}`}>{formatRole(member.role)}</span>
                <time>{formatDate(member.joined_at)}</time>
              </div>
            </article>
          ))}
          {!members.length && <div className="empty-copy">В этом пространстве пока нет участников.</div>}
        </div>
      </section>

      {canInvite && (
        <section className="profile-card action-history">
          <div className="section-heading">
            <div>
              <div className="card-label">Приглашения</div>
              <h2>Отправленные приглашения</h2>
            </div>
            {loading && <span className="muted-text">Загрузка...</span>}
          </div>

          <div className="invite-list">
            {invites.map((invite) => (
              <article className="invite-card" key={invite.uuid || invite.id || invite.email}>
                <div className="invite-card-top">
                  <div>
                    <strong>{invite.email || "Без email"}</strong>
                    <span>
                      Отправил {invite.invited_by_username || "участник"} · роль {formatRole(invite.role)} · статус{" "}
                      {formatInviteStatus(invite.status)}
                    </span>
                  </div>
                  <time>{formatDate(invite.created_at)}</time>
                </div>
                <p>{formatInviteTiming(invite)}</p>
              </article>
            ))}
            {!invites.length && <div className="empty-copy">Отправленных приглашений пока нет.</div>}
          </div>
        </section>
      )}

      {isInviteDialogOpen && (
        <div className="dialog-backdrop" onClick={() => setIsInviteDialogOpen(false)}>
          <div className="dialog-card profile-settings-dialog" onClick={(event) => event.stopPropagation()}>
            <div className="section-heading">
              <div>
                <div className="card-label">Приглашение</div>
                <h2>Добавить участника</h2>
              </div>
              <button className="text-button" type="button" onClick={() => setIsInviteDialogOpen(false)}>
                Закрыть
              </button>
            </div>

            <form
              className="compact-form"
              onSubmit={async (event) => {
                event.preventDefault();
                const success = await onSubmitInvite();
                if (success) {
                  setIsInviteDialogOpen(false);
                }
              }}
            >
              <input
                type="email"
                value={inviteForm.email}
                onChange={(event) => onInviteFormChange((current) => ({ ...current, email: event.target.value }))}
                placeholder="Email участника"
              />
              <select
                value={inviteForm.role}
                onChange={(event) => onInviteFormChange((current) => ({ ...current, role: event.target.value }))}
              >
                <option value="viewer">Участник: только просмотр</option>
                <option value="editor">Креатор: может изменять и приглашать</option>
              </select>
              <button className="primary-button" type="submit" disabled={loading}>
                Отправить приглашение
              </button>
            </form>
          </div>
        </div>
      )}
    </main>
  );
}

function StatCard({ label, value }) {
  return (
    <article className="stat-card">
      <span>{label}</span>
      <strong>{Number(value || 0)}</strong>
    </article>
  );
}

function initials(value = "") {
  const normalized = value.trim();
  if (!normalized) {
    return "WS";
  }

  return normalized.slice(0, 2).toUpperCase();
}

function formatDate(value) {
  if (!value) {
    return "нет данных";
  }

  return new Date(value * 1000).toLocaleString("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatRole(value) {
  if (value === "owner") {
    return "Владелец";
  }

  if (value === "editor" || value === "creator") {
    return "Креатор";
  }

  return "Участник";
}

function formatMembershipStatus(value) {
  if (value === "active") {
    return "Активен";
  }

  if (value === "pending") {
    return "Ожидает";
  }

  if (value === "removed") {
    return "Удален";
  }

  return value || "не указан";
}

function formatInviteStatus(value) {
  if (value === "accepted") {
    return "принято";
  }

  if (value === "declined") {
    return "отклонено";
  }

  if (value === "expired") {
    return "истекло";
  }

  return "ожидает";
}

function formatInviteTiming(invite) {
  if (invite?.status === "accepted") {
    return `Принято ${formatDate(invite.accepted_at)}.`;
  }

  if (invite?.status === "declined") {
    return `Отклонено ${formatDate(invite.declined_at)}.`;
  }

  if (invite?.status === "expired") {
    return `Срок приглашения истек ${formatDate(invite.expires_at)}.`;
  }

  return `Действует до ${formatDate(invite.expires_at)}.`;
}

function hasMemberChanges(member, draft) {
  if (!member || !draft) {
    return false;
  }

  return (draft.role || member.role) !== member.role || (draft.status || member.status) !== member.status;
}
