export default function NotesSection({
  loading,
  notes,
  selectedCategory,
  selectedNote,
  selectedNoteId,
  onSelectNote,
  noteForm,
  onNoteFormChange,
  noteEditorForm,
  onNoteEditorFormChange,
  parsedEditorTags,
  onCreateNote,
  onUpdateNote,
  onDeleteNote,
  onOpenDuplicateDialog,
  files,
  onDownloadFile,
  onDeleteFile,
  onUploadFile,
  onFilePick,
  fileInputRef,
  fileDraft,
}) {
  return (
    <section className="content-panel">
      <div className="content-split">
        <div className="editor-column">
          {selectedNote ? (
            <div className="note-card-sheet">
              <header className="sheet-header">
                <h2>Редактирование заметки</h2>
                <div className="sheet-header-actions">
                  <span>{loading ? "Сохранение..." : formatUpdatedAt(selectedNote.updated_at || selectedNote.created_date)}</span>
                </div>
              </header>
              <form className="note-creator" onSubmit={onUpdateNote}>
                <input
                  value={noteEditorForm.header}
                  onChange={(event) => onNoteEditorFormChange((current) => ({ ...current, header: event.target.value }))}
                  placeholder="Название заметки"
                />
                <textarea
                  rows={10}
                  value={noteEditorForm.body}
                  onChange={(event) => onNoteEditorFormChange((current) => ({ ...current, body: event.target.value }))}
                  placeholder="Текст заметки"
                />

                <div className="schedule-card">
                  <div className="schedule-header">
                    <div className="side-title">Дата и время</div>
                    <label className="schedule-toggle">
                      <input
                        type="checkbox"
                        checked={noteEditorForm.scheduled}
                        onChange={(event) =>
                          onNoteEditorFormChange((current) => ({ ...current, scheduled: event.target.checked }))
                        }
                      />
                      <span>Запланировать заметку</span>
                    </label>
                  </div>

                  {noteEditorForm.scheduled ? (
                    <div className="schedule-fields">
                      <label className="schedule-field">
                        <span>Дата</span>
                        <input
                          type="date"
                          value={noteEditorForm.eventDate}
                          onChange={(event) =>
                            onNoteEditorFormChange((current) => ({ ...current, eventDate: event.target.value }))
                          }
                        />
                      </label>
                      <label className="schedule-field">
                        <span>Начало</span>
                        <input
                          type="time"
                          value={noteEditorForm.eventStartTime}
                          onChange={(event) =>
                            onNoteEditorFormChange((current) => ({ ...current, eventStartTime: event.target.value }))
                          }
                        />
                      </label>
                      <label className="schedule-field">
                        <span>Конец</span>
                        <input
                          type="time"
                          value={noteEditorForm.eventEndTime}
                          onChange={(event) =>
                            onNoteEditorFormChange((current) => ({ ...current, eventEndTime: event.target.value }))
                          }
                        />
                      </label>
                      <label className="schedule-toggle secondary">
                        <input
                          type="checkbox"
                          checked={noteEditorForm.eventEnabled}
                          onChange={(event) =>
                            onNoteEditorFormChange((current) => ({ ...current, eventEnabled: event.target.checked }))
                          }
                        />
                        <span>Показывать в календаре</span>
                      </label>
                    </div>
                  ) : (
                    <div className="empty-copy">Эта заметка пока не привязана к дате и времени.</div>
                  )}
                </div>

                <div className="tag-section">
                  <div className="side-title">Теги</div>
                  <div className="tag-list">
                    {parsedEditorTags.length ? (
                      parsedEditorTags.map((tagName) => <span key={tagName}>#{tagName}</span>)
                    ) : (
                      <span>Тегов пока нет</span>
                    )}
                  </div>
                  <input
                    value={noteEditorForm.tags}
                    onChange={(event) => onNoteEditorFormChange((current) => ({ ...current, tags: event.target.value }))}
                    placeholder="Теги через запятую"
                  />
                </div>
                <div className="editor-actions">
                  <button className="primary-button" type="submit" disabled={loading}>
                    Сохранить
                  </button>
                  <button
                    className="secondary-button danger-button"
                    type="button"
                    onClick={() => void onDeleteNote(selectedNote.uuid)}
                    disabled={loading}
                  >
                    Удалить заметку
                  </button>
                  <button className="secondary-button" type="button" onClick={onOpenDuplicateDialog} disabled={loading}>
                    Дублировать
                  </button>
                </div>
              </form>
              <div className="attachments">
                <div className="side-title">Вложения</div>
                <form className="attachment-form" onSubmit={onUploadFile}>
                  <input ref={fileInputRef} type="file" onChange={(event) => onFilePick(event.target.files?.[0] || null)} />
                  <button className="secondary-button" type="submit">
                    {fileDraft ? "Прикрепить" : "Загрузить"}
                  </button>
                </form>
                <div className="attachment-list">
                  {files.map((file) => (
                    <div className="attachment-item" key={file.id}>
                      <div>
                        <strong>{file.name}</strong>
                        <span>{formatSize(file.size)}</span>
                      </div>
                      <div className="attachment-actions">
                        <button className="text-button" type="button" onClick={() => onDownloadFile(file.id, file.name)}>
                          Скачать
                        </button>
                        <button className="text-button danger" type="button" onClick={() => onDeleteFile(file.id)}>
                          удалить
                        </button>
                      </div>
                    </div>
                  ))}
                  {!files.length && <div className="empty-copy">Вложений пока нет.</div>}
                </div>
              </div>
            </div>
          ) : (
            <form className="note-creator" onSubmit={onCreateNote}>
              <input
                value={noteForm.header}
                onChange={(event) => onNoteFormChange((current) => ({ ...current, header: event.target.value }))}
                placeholder="Название заметки"
              />
              <textarea
                rows={8}
                value={noteForm.body}
                onChange={(event) => onNoteFormChange((current) => ({ ...current, body: event.target.value }))}
                placeholder="Текст заметки"
              />
              <input
                value={noteForm.tags}
                onChange={(event) => onNoteFormChange((current) => ({ ...current, tags: event.target.value }))}
                placeholder="Теги через запятую"
              />
              <button className="primary-button" type="submit">
                Создать заметку
              </button>
            </form>
          )}
        </div>
        <aside className="detail-column notes-column">
          <div className="notes-side-header">
            <div>
              <div className="panel-title">Заметки</div>
              {selectedCategory && (
                <div className="category-context">
                  <span className="category-context-marker" style={{ backgroundColor: selectedCategory.color || "#9db8ff" }} />
                  <span>{selectedCategory.name}</span>
                </div>
              )}
            </div>
            <button className="secondary-button" type="button" onClick={() => onSelectNote("")}>
              Новая
            </button>
          </div>
          <div className="notes-list">
            {notes.map((note) => (
              <article
                key={`side-${note.uuid}`}
                className={`note-preview ${selectedNoteId === note.uuid ? "active" : ""}`}
                onClick={() => onSelectNote(note.uuid)}
              >
                <div className="note-preview-top">
                  <h2>{note.header}</h2>
                  <button
                    className="text-button danger"
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      void onDeleteNote(note.uuid);
                    }}
                  >
                    удалить
                  </button>
                </div>
                {note.event?.start_at && (
                  <div className="note-preview-meta">
                    {formatScheduledAt(note.event.start_at)}
                    {note.event.enabled === false ? " • скрыта из календаря" : ""}
                  </div>
                )}
                <div className="note-preview-line" />
                <p>{note.short_body || note.body || "Без описания"}</p>
              </article>
            ))}
            {!notes.length && <div className="empty-copy">Для выбранной категории заметок пока нет.</div>}
          </div>
        </aside>
      </div>
    </section>
  );
}

function formatSize(size) {
  if (!size) {
    return "0 B";
  }
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function formatUpdatedAt(unixSeconds) {
  if (!unixSeconds) {
    return "Без даты";
  }

  return `Обновлено ${new Date(unixSeconds * 1000).toLocaleDateString("ru-RU")}`;
}

function formatScheduledAt(unixSeconds) {
  if (!unixSeconds) {
    return "";
  }

  return new Date(unixSeconds * 1000).toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}
