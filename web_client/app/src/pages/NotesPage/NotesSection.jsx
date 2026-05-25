export default function NotesSection({
  loading,
  notes,
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
                  <span>{loading ? "Сохранение..." : new Date().toLocaleDateString("ru-RU")}</span>
                  <button className="secondary-button" type="button" onClick={() => onSelectNote("")}>
                    К созданию
                  </button>
                </div>
              </header>
              <form className="note-creator" onSubmit={onUpdateNote}>
                <input
                  value={noteEditorForm.header}
                  onChange={(event) => onNoteEditorFormChange((current) => ({ ...current, header: event.target.value }))}
                  placeholder="Имя заметки"
                />
                <textarea
                  rows={10}
                  value={noteEditorForm.body}
                  onChange={(event) => onNoteEditorFormChange((current) => ({ ...current, body: event.target.value }))}
                  placeholder="Текст заметки"
                />
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
                placeholder="Имя заметки"
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
            <div className="panel-title">Заметки</div>
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
                <div className="note-preview-line" />
                <p>{note.short_body || note.body}</p>
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
