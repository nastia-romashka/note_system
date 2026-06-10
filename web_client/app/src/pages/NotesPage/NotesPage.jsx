import CategorySection from "./CategorySection";
import NotesSection from "./NotesSection";

export default function NotesPage({
  loading,
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
  return (
    <main className="notes-page">
      <header className="notes-header">
        <h1>Заметки</h1>
        <div className="notes-toolbar">
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
            Личный кабинет
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
