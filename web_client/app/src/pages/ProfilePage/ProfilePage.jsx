import { useState } from "react";

const ACTION_LABELS = {
  "auth.login": "вошел в аккаунт",
  "profile.updated": "обновил профиль",
  "note.created": "создал заметку",
  "note.updated": "обновил заметку",
  "note.deleted": "удалил заметку",
  "category.created": "создал категорию",
  "category.deleted": "удалил категорию",
  "file.uploaded": "загрузил файл",
  "file.deleted": "удалил файл",
  "tag.created": "создал тег",
  "tag.deleted": "удалил тег",
};

export default function ProfilePage({
  summary,
  actions,
  loading,
  profileForm,
  onProfileFormChange,
  onSubmitProfileUpdate,
  onRefresh,
  onBackToNotes,
  onOpenGraph,
  onOpenCalendar,
  onLogout,
}) {
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const profile = summary?.profile || {};
  const stats = summary?.stats || {};
  const upcomingEvents = Array.isArray(summary?.upcoming_events) ? summary.upcoming_events : [];

  return (
    <main className="profile-page">
      <header className="profile-header">
        <div>
          <span className="eyebrow">Личный кабинет</span>
          <h1>{profile.username || "Профиль"}</h1>
          <p>Профиль, статистика и последние действия в одном месте.</p>
        </div>
        <div className="profile-actions">
          <button className="secondary-button" type="button" onClick={onBackToNotes}>
            К заметкам
          </button>
          <button className="secondary-button" type="button" onClick={onOpenGraph}>
            Граф
          </button>
          <button className="secondary-button" type="button" onClick={onOpenCalendar}>
            Календарь
          </button>
          <button className="secondary-button" type="button" onClick={onRefresh} disabled={loading}>
            Обновить
          </button>
          <button className="secondary-button" type="button" onClick={onLogout}>
            Выйти
          </button>
        </div>
      </header>

      <section className="profile-grid">
        <article className="profile-card profile-main-card profile-main-card-shell">
          <button
            className="profile-settings-trigger"
            type="button"
            aria-label="Открыть настройки профиля"
            onClick={() => setIsSettingsOpen(true)}
          >
            <span />
            <span />
            <span />
          </button>

          <div className="profile-avatar">{initials(profile.username)}</div>
          <div>
            <div className="card-label">Профиль пользователя</div>
            <h2>{profile.username || "Без имени"}</h2>
            <dl className="profile-details">
              <div>
                <dt>Email</dt>
                <dd>{profile.email || "не указан"}</dd>
              </div>
              <div>
                <dt>Дата регистрации</dt>
                <dd>{formatDate(profile.created_at)}</dd>
              </div>
              <div>
                <dt>Последний вход</dt>
                <dd>{formatDate(profile.last_login_at)}</dd>
              </div>
            </dl>
          </div>
        </article>

        <article className="profile-card">
          <div className="card-label">Последняя активность</div>
          <strong className="activity-date">{formatDate(stats.last_activity_at)}</strong>
        </article>
      </section>

      <section className="stats-grid">
        <StatCard label="Категории" value={stats.categories_count} />
        <StatCard label="Заметки" value={stats.notes_count} />
        <StatCard label="Теги" value={stats.tags_count} />
        <StatCard label="Файлы" value={stats.files_count} />
      </section>

      <section className="profile-card action-history">
        <div className="section-heading">
          <div>
            <div className="card-label">Ближайшие события</div>
            <h2>Что запланировано</h2>
          </div>
          {loading && <span className="muted-text">Загрузка...</span>}
        </div>

        <div className="action-list">
          {upcomingEvents.map((event) => (
            <article className="action-item" key={event.uuid || `${event.header}-${event.event?.start_at || 0}`}>
              <div className="action-dot" />
              <div>
                <strong>{event.header || "Без названия"}</strong>
                <span>{event.short_body || event.body || "Запланированная заметка"}</span>
              </div>
              <time>{formatDate(event.event?.start_at)}</time>
            </article>
          ))}
          {!upcomingEvents.length && <div className="empty-copy">Ближайших событий пока нет.</div>}
        </div>
      </section>

      <section className="profile-card action-history">
        <div className="section-heading">
          <div>
            <div className="card-label">История действий</div>
            <h2>Последние события</h2>
          </div>
          {loading && <span className="muted-text">Загрузка...</span>}
        </div>

        <div className="action-list">
          {actions.map((action) => (
            <article className="action-item" key={action.uuid || `${action.action}-${action.created_at}`}>
              <div className="action-dot" />
              <div>
                <strong>{ACTION_LABELS[action.action] || action.action}</strong>
              </div>
              <time>{formatDate(action.created_at)}</time>
            </article>
          ))}
          {!actions.length && <div className="empty-copy">История действий пока пустая.</div>}
        </div>
      </section>

      {isSettingsOpen && (
        <div className="dialog-backdrop" onClick={() => setIsSettingsOpen(false)}>
          <div className="dialog-card profile-settings-dialog" onClick={(event) => event.stopPropagation()}>
            <div className="section-heading">
              <div>
                <div className="card-label">Настройки профиля</div>
                <h2>Имя, email и пароль</h2>
              </div>
              <button className="text-button" type="button" onClick={() => setIsSettingsOpen(false)}>
                Закрыть
              </button>
            </div>

            <form
              className="compact-form"
              onSubmit={(event) => {
                event.preventDefault();
                void onSubmitProfileUpdate();
              }}
            >
              <input
                value={profileForm.username}
                onChange={(event) => onProfileFormChange((current) => ({ ...current, username: event.target.value }))}
                placeholder="Имя пользователя"
              />
              <input
                type="email"
                value={profileForm.email}
                onChange={(event) => onProfileFormChange((current) => ({ ...current, email: event.target.value }))}
                placeholder="Email"
              />
              <input
                type="password"
                value={profileForm.currentPassword}
                onChange={(event) =>
                  onProfileFormChange((current) => ({ ...current, currentPassword: event.target.value }))
                }
                placeholder="Текущий пароль"
              />
              <input
                type="password"
                value={profileForm.newPassword}
                onChange={(event) =>
                  onProfileFormChange((current) => ({ ...current, newPassword: event.target.value }))
                }
                placeholder="Новый пароль"
              />
              <input
                type="password"
                value={profileForm.confirmPassword}
                onChange={(event) =>
                  onProfileFormChange((current) => ({ ...current, confirmPassword: event.target.value }))
                }
                placeholder="Подтвердите новый пароль"
              />
              <button className="primary-button" type="submit" disabled={loading}>
                Сохранить изменения
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

function initials(username = "") {
  const value = username.trim();
  if (!value) {
    return "NS";
  }

  return value.slice(0, 2).toUpperCase();
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
