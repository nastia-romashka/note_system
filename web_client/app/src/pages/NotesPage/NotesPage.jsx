import CategorySection from "./CategorySection";
import NotesSection from "./NotesSection";

export default function NotesPage({
  loading,
  categories,
  selectedCategoryId,
  onSelectCategory,
  onOpenSubcategoryDialog,
  notes,
  selectedNote,
  selectedNoteId,
  onSelectNote,
  onOpenProfile,
  onOpenGraph,
  onLogout,
  search,
  onSearch,
  categoryForm,
  onCategoryFormChange,
  onCreateCategory,
  onDeleteCategory,
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
    <main className="notes-page">
      <header className="notes-header">
        <h1>Заметки</h1>
        <div className="notes-toolbar">
          <label className="search-box">
            <span>⌕</span>
            <input value={search} onChange={(event) => onSearch(event.target.value)} placeholder="Поиск..." />
          </label>
          <button className="secondary-button" onClick={onOpenProfile} type="button">
            Личный кабинет
          </button>
          <button className="secondary-button" onClick={onOpenGraph} type="button">
            Граф
          </button>
          <button className="secondary-button" onClick={onLogout} type="button">
            Выйти
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
        />
        <NotesSection
          loading={loading}
          notes={notes}
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
