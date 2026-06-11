import WorkspaceContextMenu from "../../components/WorkspaceContextMenu";
import CategorySection from "./CategorySection";
import NotesSection from "./NotesSection";

export default function NotesPage({
  loading,
  currentWorkspace,
  workspaces,
  categories,
  selectedCategory,
  selectedCategoryId,
  onSelectCategory,
  onOpenSubcategoryDialog,
  notes,
  selectedNote,
  selectedNoteId,
  onSelectNote,
  onOpenProfile,
  onOpenGraph,
  onOpenCalendar,
  onSelectPersonalWorkspace,
  onSelectWorkspace,
  onCreateWorkspace,
  search,
  onSearch,
  searchResults,
  onSelectSearchResult,
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
  const isWorkspaceMode = Boolean(currentWorkspace);
  const pageTitle = isWorkspaceMode
    ? `${currentWorkspace.name}: Заметки`
    : "Заметки";
  const contextButtonLabel = isWorkspaceMode ? "Настройки пространства" : "Личный кабинет";

  return (
    <main className="notes-page">
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
          <p>Категории, поиск и заметки в одном рабочем контексте.</p>
        </div>
        <div className="profile-actions page-header-actions">
          <div className="search-shell">
            <label className="search-box">
              <span>⌕</span>
              <input value={search} onChange={(event) => onSearch(event.target.value)} placeholder="Поиск..." />
            </label>
            {!!searchResults.length && (
              <div className="search-results-dropdown">
                {searchResults.map((note) => (
                  <button
                    key={`search-${note.uuid}`}
                    className="search-result-item"
                    type="button"
                    onClick={() => onSelectSearchResult(note)}
                  >
                    <strong>{note.header}</strong>
                    <span>{note.short_body || note.body}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
          <button className="secondary-button" onClick={onOpenGraph} type="button">
            Граф
          </button>
          <button className="secondary-button" onClick={onOpenCalendar} type="button">
            Календарь
          </button>
          <button className="secondary-button" onClick={onOpenProfile} type="button">
            {contextButtonLabel}
          </button>
        </div>
      </header>

      <section className="notes-layout">
        <CategorySection
          categories={categories}
          selectedCategoryId={selectedCategoryId}
          onSelectCategory={onSelectCategory}
          onOpenSubcategoryDialog={onOpenSubcategoryDialog}
          categoryForm={categoryForm}
          onCategoryFormChange={onCategoryFormChange}
          onCreateCategory={onCreateCategory}
          onDeleteCategory={onDeleteCategory}
          categoryEditor={categoryEditor}
          onCategoryEditorChange={onCategoryEditorChange}
          onCloseCategoryEditor={onCloseCategoryEditor}
          onStartRenameCategory={onStartRenameCategory}
          onStartRecolorCategory={onStartRecolorCategory}
          onSubmitCategoryRename={onSubmitCategoryRename}
          onSubmitCategoryColor={onSubmitCategoryColor}
        />
        <NotesSection
          loading={loading}
          notes={notes}
          selectedCategory={selectedCategory}
          selectedNote={selectedNote}
          selectedNoteId={selectedNoteId}
          onSelectNote={onSelectNote}
          noteForm={noteForm}
          onNoteFormChange={onNoteFormChange}
          noteEditorForm={noteEditorForm}
          onNoteEditorFormChange={onNoteEditorFormChange}
          parsedEditorTags={parsedEditorTags}
          onCreateNote={onCreateNote}
          onUpdateNote={onUpdateNote}
          onDeleteNote={onDeleteNote}
          onOpenDuplicateDialog={onOpenDuplicateDialog}
          files={files}
          onDownloadFile={onDownloadFile}
          onDeleteFile={onDeleteFile}
          onUploadFile={onUploadFile}
          onFilePick={onFilePick}
          fileInputRef={fileInputRef}
          fileDraft={fileDraft}
        />
      </section>
    </main>
  );
}
