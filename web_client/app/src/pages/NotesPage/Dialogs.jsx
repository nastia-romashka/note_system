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
