export default function CategorySection({
  categories,
  selectedCategoryId,
  onSelectCategory,
  onOpenSubcategoryDialog,
  categoryForm,
  onCategoryFormChange,
  onCreateCategory,
  onDeleteCategory,
  categoryEditor,
  onCategoryEditorChange,
  onCloseCategoryEditor,
  onStartRenameCategory,
  onStartRecolorCategory,
  onSubmitCategoryRename,
  onSubmitCategoryColor,
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
            categoryEditor={categoryEditor}
            onCategoryEditorChange={onCategoryEditorChange}
            onCloseCategoryEditor={onCloseCategoryEditor}
            onStartRenameCategory={onStartRenameCategory}
            onStartRecolorCategory={onStartRecolorCategory}
            onSubmitCategoryRename={onSubmitCategoryRename}
            onSubmitCategoryColor={onSubmitCategoryColor}
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
  categoryEditor,
  onCategoryEditorChange,
  onCloseCategoryEditor,
  onStartRenameCategory,
  onStartRecolorCategory,
  onSubmitCategoryRename,
  onSubmitCategoryColor,
}) {
  const children = Array.isArray(category.children) ? category.children : [];
  const isActive = selectedCategoryId === category.uuid;
  const isEditing = categoryEditor?.uuid === category.uuid;
  const tintedBackground = isActive ? withAlpha(category.color || "#9db8ff", 0.18) : undefined;

  return (
    <div className="category-node">
      <div
        className={`category-item ${isActive ? "active" : ""}`}
        style={{
          marginLeft: `${depth * 16}px`,
          ...(tintedBackground ? { background: tintedBackground } : {}),
        }}
      >
        <button className="category-main" type="button" onClick={() => onSelectCategory(category.uuid)}>
          <span className="category-index">{indexLabel}</span>
          <span className="category-name">{category.name}</span>
        </button>
        <div className="category-actions">
          <button className="category-add-button" type="button" onClick={() => onOpenSubcategoryDialog(category)}>
            +
          </button>
          <div className="category-menu-shell">
            <button
              className="category-menu-button"
              type="button"
              onClick={() =>
                isEditing ? onCloseCategoryEditor() : onCategoryEditorChange({ uuid: category.uuid, mode: "menu", name: category.name, color: category.color || "#9db8ff" })
              }
            >
              ...
            </button>
            {isEditing && (
              <div className="category-menu-popup">
                {categoryEditor.mode === "menu" && (
                  <>
                    <button type="button" onClick={() => onStartRenameCategory(category)}>
                      Переименовать
                    </button>
                    <button type="button" onClick={() => onStartRecolorCategory(category)}>
                      Изменить цвет
                    </button>
                    <button type="button" className="danger" onClick={() => void onDeleteCategory(category.uuid)}>
                      Удалить
                    </button>
                  </>
                )}
                {categoryEditor.mode === "rename" && (
                  <form
                    className="category-inline-form"
                    onSubmit={(event) => {
                      event.preventDefault();
                      void onSubmitCategoryRename();
                    }}
                  >
                    <input
                      autoFocus
                      value={categoryEditor.name}
                      onChange={(event) => onCategoryEditorChange((current) => ({ ...current, name: event.target.value }))}
                      onKeyDown={(event) => {
                        if (event.key === "Escape") {
                          onCloseCategoryEditor();
                        }
                      }}
                      placeholder="Название категории"
                    />
                    <div className="category-inline-actions">
                      <button type="submit">Сохранить</button>
                      <button type="button" onClick={onCloseCategoryEditor}>
                        Отмена
                      </button>
                    </div>
                  </form>
                )}
                {categoryEditor.mode === "color" && (
                  <div className="category-inline-form">
                    <input
                      autoFocus
                      type="color"
                      value={categoryEditor.color}
                      onChange={(event) => onCategoryEditorChange((current) => ({ ...current, color: event.target.value }))}
                    />
                    <div className="category-inline-actions">
                      <button type="button" onClick={() => void onSubmitCategoryColor()}>
                        Сохранить
                      </button>
                      <button type="button" onClick={onCloseCategoryEditor}>
                        Отмена
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
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
              categoryEditor={categoryEditor}
              onCategoryEditorChange={onCategoryEditorChange}
              onCloseCategoryEditor={onCloseCategoryEditor}
              onStartRenameCategory={onStartRenameCategory}
              onStartRecolorCategory={onStartRecolorCategory}
              onSubmitCategoryRename={onSubmitCategoryRename}
              onSubmitCategoryColor={onSubmitCategoryColor}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function withAlpha(hexColor, alpha) {
  const hex = String(hexColor || "").replace("#", "");
  if (hex.length !== 6) {
    return undefined;
  }

  const red = Number.parseInt(hex.slice(0, 2), 16);
  const green = Number.parseInt(hex.slice(2, 4), 16);
  const blue = Number.parseInt(hex.slice(4, 6), 16);

  return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
}
