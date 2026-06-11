import { CalendarCreateDialog } from "../NotesPage/Dialogs";
import WorkspaceContextMenu from "../../components/WorkspaceContextMenu";

const WEEKDAY_LABELS = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

export default function CalendarPage({
  currentWorkspace,
  workspaces,
  currentMonth,
  selectedDay,
  notesByDay,
  selectedDayNotes,
  categories,
  createDialog,
  onCreateDialogChange,
  onOpenCreateDialog,
  onCloseCreateDialog,
  onConfirmCreateDialog,
  onSelectDay,
  onChangeMonth,
  onToday,
  onOpenNote,
  onOpenGraph,
  onOpenNotes,
  onOpenProfile,
  onSelectPersonalWorkspace,
  onSelectWorkspace,
  onCreateWorkspace,
}) {
  const monthDays = buildCalendarDays(currentMonth);
  const categoryNames = buildCategoryNameMap(categories);
  const isWorkspaceMode = Boolean(currentWorkspace);
  const pageTitle = isWorkspaceMode
    ? `${currentWorkspace.name}: Календарь`
    : "Календарь";
  const contextButtonLabel = isWorkspaceMode ? "Настройки пространства" : "Личный кабинет";
  const selectedLabel = selectedDay.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
  });

  return (
    <main className="calendar-page">
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
            <span className="eyebrow">{isWorkspaceMode ? "Общий режим" : "Личный режим"}</span>
          </div>
          <h1>{pageTitle}</h1>
          <p>{currentMonth.toLocaleDateString("ru-RU", { month: "long", year: "numeric" })}</p>
        </div>
        <div className="profile-actions page-header-actions">
          <button className="secondary-button" type="button" onClick={onOpenGraph}>
            Граф
          </button>
          <button className="secondary-button" type="button" onClick={onOpenNotes}>
            Заметки
          </button>
          <button className="secondary-button" type="button" onClick={onOpenProfile}>
            {contextButtonLabel}
          </button>
        </div>
      </header>

      <section className="calendar-layout">
        <div className="calendar-board">
          <div className="calendar-board-toolbar">
            <button className="secondary-button" type="button" onClick={() => onChangeMonth(-1)}>
              Назад
            </button>
            <button className="secondary-button" type="button" onClick={onToday}>
              Сегодня
            </button>
            <button className="secondary-button" type="button" onClick={() => onChangeMonth(1)}>
              Вперед
            </button>
          </div>

          <div className="calendar-grid">
            {WEEKDAY_LABELS.map((label) => (
              <div key={label} className="calendar-weekday">
                {label}
              </div>
            ))}

            {monthDays.map((day) => {
              const key = formatDayKey(day.date);
              const hasNotes = (notesByDay.get(key) || []).length > 0;
              const isSelected = formatDayKey(selectedDay) === key;
              const isCurrentMonth = day.date.getMonth() === currentMonth.getMonth();

              return (
                <button
                  key={key}
                  type="button"
                  className={`calendar-day ${isSelected ? "selected" : ""} ${isCurrentMonth ? "" : "outside"}`}
                  onClick={() => onSelectDay(day.date)}
                >
                  <span>{day.date.getDate()}</span>
                  {hasNotes && <span className="calendar-day-dot" aria-hidden="true" />}
                </button>
              );
            })}
          </div>
        </div>

        <aside className="calendar-side-panel">
          <div className="calendar-side-header">
            <div>
              <div className="panel-title">День</div>
              <p>{selectedLabel}</p>
            </div>
            <button className="primary-button" type="button" onClick={onOpenCreateDialog}>
              Создать заметку
            </button>
          </div>

          <div className="calendar-day-list">
            {selectedDayNotes.length ? (
              selectedDayNotes.map((note) => (
                <button
                  key={note.uuid}
                  type="button"
                  className="calendar-note-card"
                  onClick={() => onOpenNote(note)}
                >
                  <div className="calendar-note-time">{formatTimeRange(note.event)}</div>
                  <strong>{note.header || "Без заголовка"}</strong>
                  <span>{resolveCategoryName(categoryNames, note.category_uuid)}</span>
                </button>
              ))
            ) : (
              <div className="calendar-empty-state">
                <strong>На этот день ничего не запланировано</strong>
                <span>Создайте заметку на выбранную дату и задайте ей удобное время.</span>
              </div>
            )}
          </div>
        </aside>
      </section>

      {createDialog && (
        <CalendarCreateDialog
          categories={categories}
          value={createDialog}
          onChange={onCreateDialogChange}
          onConfirm={onConfirmCreateDialog}
          onCancel={onCloseCreateDialog}
        />
      )}
    </main>
  );
}

function buildCalendarDays(currentMonth) {
  const firstDay = new Date(currentMonth.getFullYear(), currentMonth.getMonth(), 1);
  const lastDay = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 0);
  const offset = (firstDay.getDay() + 6) % 7;
  const start = new Date(firstDay);
  start.setDate(firstDay.getDate() - offset);

  const total = Math.ceil((offset + lastDay.getDate()) / 7) * 7;

  return Array.from({ length: total }, (_, index) => {
    const date = new Date(start);
    date.setDate(start.getDate() + index);
    return { date };
  });
}

function buildCategoryNameMap(categories, map = new Map()) {
  for (const category of categories || []) {
    if (category?.uuid) {
      map.set(String(category.uuid), category.name || "Без названия");
    }

    if (Array.isArray(category?.children) && category.children.length > 0) {
      buildCategoryNameMap(category.children, map);
    }
  }

  return map;
}

function formatDayKey(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatTimeRange(event) {
  if (!event?.start_at || !event?.end_at) {
    return "Без времени";
  }

  const start = new Date(event.start_at * 1000).toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
  const end = new Date(event.end_at * 1000).toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });

  return `${start} - ${end}`;
}

function resolveCategoryName(categoryNames, categoryId) {
  if (!categoryId) {
    return "Категория не указана";
  }

  return categoryNames.get(String(categoryId)) || "Категория не найдена";
}
