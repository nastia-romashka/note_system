export function ConfirmDialog({ title, description, confirmLabel, cancelLabel, onConfirm, onCancel, loading }) {
  return (
    <div className="dialog-backdrop" role="presentation">
      <div className="dialog-card" role="dialog" aria-modal="true" aria-label={title}>
        <h2>{title}</h2>
        <p>{description}</p>
        <div className="dialog-actions">
          <button className="secondary-button" type="button" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </button>
          <button className="primary-button danger-button" type="button" onClick={onConfirm} disabled={loading}>
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

export function SubcategoryDialog({
  parentCategoryName,
  value,
  onChange,
  onConfirm,
  onCancel,
  loading,
}) {
  return (
    <div className="dialog-backdrop" role="presentation">
      <div className="dialog-card" role="dialog" aria-modal="true" aria-label="Создание подкатегории">
        <h2>Новая подкатегория</h2>
        <p>Родительская категория: {parentCategoryName}</p>
        <div className="dialog-form">
          <input
            autoFocus
            value={value}
            onChange={(event) => onChange(event.target.value)}
            placeholder="Название подкатегории"
          />
        </div>
        <div className="dialog-actions">
          <button className="secondary-button" type="button" onClick={onCancel} disabled={loading}>
            Отмена
          </button>
          <button className="primary-button" type="button" onClick={onConfirm} disabled={loading || !value.trim()}>
            Создать
          </button>
        </div>
      </div>
    </div>
  );
}

export function DuplicateNoteDialog({
  categories,
  categoryUuid,
  onCategoryChange,
  header,
  onHeaderChange,
  shortBody,
  onConfirm,
  onCancel,
  loading,
}) {
  const categoryOptions = flattenCategories(categories);

  return (
    <div className="dialog-backdrop" role="presentation">
      <div className="dialog-card" role="dialog" aria-modal="true" aria-label="Дублирование заметки">
        <h2>Дублировать заметку</h2>
        <div className="dialog-form compact-form">
          <label>
            <span>Категория</span>
            <select value={categoryUuid} onChange={(event) => onCategoryChange(event.target.value)}>
              {categoryOptions.map((category) => (
                <option key={category.uuid} value={category.uuid}>
                  {category.label}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Заголовок</span>
            <input autoFocus value={header} onChange={(event) => onHeaderChange(event.target.value)} />
          </label>
          <label>
            <span>Короткое тело заметки</span>
            <textarea rows={4} value={shortBody} readOnly />
          </label>
        </div>
        <div className="dialog-actions">
          <button className="secondary-button" type="button" onClick={onCancel} disabled={loading}>
            Отмена
          </button>
          <button
            className="primary-button"
            type="button"
            onClick={onConfirm}
            disabled={loading || !header.trim() || !categoryUuid}
          >
            Дублировать
          </button>
        </div>
      </div>
    </div>
  );
}

export function CalendarCreateDialog({ categories, value, onChange, onConfirm, onCancel }) {
  const categoryOptions = flattenCategories(categories);

  return (
    <div className="dialog-backdrop" role="presentation">
      <div className="dialog-card" role="dialog" aria-modal="true" aria-label="Создание заметки из календаря">
        <h2>Новая заметка</h2>
        <p>Выберите категорию, задайте заголовок и время, чтобы сразу отметить заметку в календаре.</p>
        <div className="dialog-form compact-form">
          <label>
            <span>Категория</span>
            <select value={value.categoryUuid} onChange={(event) => onChange((current) => ({ ...current, categoryUuid: event.target.value }))}>
              {categoryOptions.map((category) => (
                <option key={category.uuid} value={category.uuid}>
                  {category.label}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Заголовок</span>
            <input
              autoFocus
              value={value.header}
              onChange={(event) => onChange((current) => ({ ...current, header: event.target.value }))}
              placeholder="Например, Подготовить демо"
            />
          </label>
          <label>
            <span>Дата</span>
            <input
              type="date"
              value={value.date}
              onChange={(event) => onChange((current) => ({ ...current, date: event.target.value }))}
            />
          </label>
          <div className="compact-row">
            <label>
              <span>Начало</span>
              <input
                type="time"
                value={value.startTime}
                onChange={(event) => onChange((current) => ({ ...current, startTime: event.target.value }))}
              />
            </label>
            <label>
              <span>Конец</span>
              <input
                type="time"
                value={value.endTime}
                onChange={(event) => onChange((current) => ({ ...current, endTime: event.target.value }))}
              />
            </label>
          </div>
        </div>
        <div className="dialog-actions">
          <button className="secondary-button" type="button" onClick={onCancel}>
            Отмена
          </button>
          <button className="primary-button" type="button" onClick={onConfirm}>
            Создать
          </button>
        </div>
      </div>
    </div>
  );
}

function flattenCategories(categories, depth = 0) {
  return (categories || []).flatMap((category) => {
    const prefix = depth > 0 ? `${"  ".repeat(depth)}• ` : "";
    const current = {
      uuid: category.uuid,
      label: `${prefix}${category.name}`,
    };
    const children = flattenCategories(category.children || [], depth + 1);
    return [current, ...children];
  });
}
