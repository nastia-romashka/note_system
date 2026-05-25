export default function CategorySection({
  categories,
  selectedCategoryId,
  onSelectCategory,
  onOpenSubcategoryDialog,
  categoryForm,
  onCategoryFormChange,
  onCreateCategory,
  onDeleteCategory,
}) {
  return (
    <aside className="sidebar-panel">
      <div className="panel-title">Категории</div>
      <form className="compact-form" onSubmit={onCreateCategory}>
        <input
          value={categoryForm.name}
          onChange={(event) => onCategoryFormChange((current) => ({ ...current, name: event.target.value }))}
          placeholder="Новая категория"
        />
        <div className="compact-row">
          <input
            type="color"
            value={categoryForm.color}
            onChange={(event) => onCategoryFormChange((current) => ({ ...current, color: event.target.value }))}
          />
          <button className="secondary-button" type="submit">
            Добавить
          </button>
        </div>
      </form>
      <div className="category-list">
        {categories.map((category, index) => (
          <CategoryTreeItem
            key={category.uuid}
            category={category}
            depth={0}
            indexLabel={`${index + 1}.`}
            selectedCategoryId={selectedCategoryId}
            onSelectCategory={onSelectCategory}
            onOpenSubcategoryDialog={onOpenSubcategoryDialog}
            onDeleteCategory={onDeleteCategory}
          />
        ))}
        {!categories.length && <div className="empty-copy">Категории пока не созданы.</div>}
      </div>
    </aside>
  );
}

function CategoryTreeItem({
  category,
  depth,
  indexLabel,
  selectedCategoryId,
  onSelectCategory,
  onOpenSubcategoryDialog,
  onDeleteCategory,
}) {
  const children = Array.isArray(category.children) ? category.children : [];

  return (
    <div className="category-node">
      <div
        className={`category-item ${selectedCategoryId === category.uuid ? "active" : ""}`}
        style={{ marginLeft: `${depth * 16}px` }}
      >
        <button className="category-main" type="button" onClick={() => onSelectCategory(category.uuid)}>
          <span className="category-index">{indexLabel}</span>
          <span>{category.name}</span>
        </button>
        <div className="category-actions">
          <button className="category-add-button" type="button" onClick={() => onOpenSubcategoryDialog(category)}>
            +
          </button>
          <button className="text-button danger" type="button" onClick={() => void onDeleteCategory(category.uuid)}>
            удалить
          </button>
        </div>
      </div>
      {children.length > 0 && (
        <div className="category-children">
          {children.map((child, index) => (
            <CategoryTreeItem
              key={child.uuid}
              category={child}
              depth={depth + 1}
              indexLabel={`${indexLabel}${index + 1}.`}
              selectedCategoryId={selectedCategoryId}
              onSelectCategory={onSelectCategory}
              onOpenSubcategoryDialog={onOpenSubcategoryDialog}
              onDeleteCategory={onDeleteCategory}
            />
          ))}
        </div>
      )}
    </div>
  );
}
