import { useState } from "react";

import { useAuthSession } from "./hooks/useAuthSession";
import { useFilesData } from "./hooks/useFilesData";
import { useGraphData } from "./hooks/useGraphData";
import { useNotesData } from "./hooks/useNotesData";
import { useProfileData } from "./hooks/useProfileData";
import AuthPage from "./pages/AuthPage/AuthPage";
import GraphPage from "./pages/GraphPage/GraphPage";
import { ConfirmDialog, SubcategoryDialog } from "./pages/NotesPage/Dialogs";
import { FlashMessage } from "./pages/NotesPage/Messages";
import NotesPage from "./pages/NotesPage/NotesPage";
import ProfilePage from "./pages/ProfilePage/ProfilePage";

const UI_PREVIEW = import.meta.env.VITE_UI_PREVIEW === "true";

function App() {
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(false);
  const { page, setPage, token, persistSession, clearSession } = useAuthSession(UI_PREVIEW);

  const notesData = useNotesData({
    token,
    uiPreview: UI_PREVIEW,
    setMessage,
    setLoading,
  });
  const filesData = useFilesData({
    token,
    selectedNoteId: notesData.selectedNoteId,
    uiPreview: UI_PREVIEW,
    setMessage,
    setLoading,
  });
  const profileData = useProfileData({
    token,
    enabled: page === "profile",
    setMessage,
  });
  const graphData = useGraphData({
    token,
    enabled: page === "graph",
    setMessage,
  });

  function logout() {
    clearSession();
    if (UI_PREVIEW) {
      setMessage({ type: "info", text: "Режим предпросмотра: возврат на экран входа." });
      return;
    }

    notesData.resetNotesState();
    filesData.resetFilesState();
    setMessage({ type: "info", text: "Сессия завершена." });
  }

  return (
    <div className="site-shell">
      <div className="page-glow page-glow-left" />
      <div className="page-glow page-glow-right" />

      {UI_PREVIEW && (
        <div className="preview-banner">
          <span>Preview mode</span>
          <div className="preview-switch">
            <button type="button" className={page === "login" ? "active" : ""} onClick={() => setPage("login")}>
              Вход
            </button>
            <button type="button" className={page === "signup" ? "active" : ""} onClick={() => setPage("signup")}>
              Регистрация
            </button>
            <button type="button" className={page === "notes" ? "active" : ""} onClick={() => setPage("notes")}>
              Заметки
            </button>
            <button type="button" className={page === "profile" ? "active" : ""} onClick={() => setPage("profile")}>
              ЛК
            </button>
            <button type="button" className={page === "graph" ? "active" : ""} onClick={() => setPage("graph")}>
              Граф
            </button>
          </div>
        </div>
      )}

      {message && <FlashMessage message={message} onClose={() => setMessage(null)} />}
      {notesData.pendingCategoryDelete && (
        <ConfirmDialog
          title="Удалить категорию?"
          description={`Категория "${notesData.pendingCategoryDelete.name}" будет удалена.`}
          confirmLabel="Удалить"
          cancelLabel="Отмена"
          onConfirm={() => void notesData.confirmDeleteCategory()}
          onCancel={() => notesData.setPendingCategoryDelete(null)}
          loading={loading}
        />
      )}
      {notesData.subcategoryParent && (
        <SubcategoryDialog
          parentCategoryName={notesData.subcategoryParent.name}
          value={notesData.subcategoryDraft}
          onChange={notesData.setSubcategoryDraft}
          onConfirm={() => void notesData.confirmCreateSubcategory()}
          onCancel={() => {
            if (loading) {
              return;
            }
            notesData.closeSubcategoryDialog();
          }}
          loading={loading}
        />
      )}

      {(page === "login" || page === "signup") && (
        <AuthPage
          page={page}
          onPageChange={setPage}
          onSessionReady={persistSession}
          onMessage={setMessage}
          uiPreview={UI_PREVIEW}
        />
      )}

      {page === "notes" && (
        <NotesPage
          loading={loading}
          categories={notesData.categories}
          selectedCategoryId={notesData.selectedCategoryId}
          onSelectCategory={notesData.setSelectedCategoryId}
          onOpenSubcategoryDialog={notesData.openSubcategoryDialog}
          notes={notesData.filteredNotes}
          selectedNote={notesData.selectedNote}
          selectedNoteId={notesData.selectedNoteId}
          onSelectNote={notesData.setSelectedNoteId}
          onOpenProfile={() => setPage("profile")}
          onOpenGraph={() => setPage("graph")}
          onLogout={logout}
          search={notesData.search}
          onSearch={notesData.setSearch}
          categoryForm={notesData.categoryForm}
          onCategoryFormChange={notesData.setCategoryForm}
          onCreateCategory={notesData.handleCreateCategory}
          onDeleteCategory={notesData.openDeleteCategoryDialog}
          noteForm={notesData.noteForm}
          onNoteFormChange={notesData.setNoteForm}
          noteEditorForm={notesData.noteEditorForm}
          onNoteEditorFormChange={notesData.setNoteEditorForm}
          parsedEditorTags={notesData.parsedEditorTags}
          onCreateNote={notesData.handleCreateNote}
          onUpdateNote={notesData.handleUpdateNote}
          onDeleteNote={notesData.handleDeleteNote}
          files={filesData.files}
          onDownloadFile={filesData.handleDownloadFile}
          onDeleteFile={filesData.handleDeleteFile}
          onUploadFile={filesData.handleUploadFile}
          onFilePick={filesData.setFileDraft}
          fileInputRef={filesData.fileInputRef}
          fileDraft={filesData.fileDraft}
        />
      )}

      {page === "profile" && (
        <ProfilePage
          summary={profileData.summary}
          actions={profileData.actions}
          loading={profileData.profileLoading}
          onRefresh={profileData.refreshProfile}
          onBackToNotes={() => setPage("notes")}
          onOpenGraph={() => setPage("graph")}
          onLogout={logout}
        />
      )}

      {page === "graph" && (
        <GraphPage
          graph={graphData.graph}
          loading={graphData.graphLoading}
          onRefresh={graphData.refreshGraph}
          onBackToNotes={() => setPage("notes")}
          onOpenProfile={() => setPage("profile")}
          onLogout={logout}
        />
      )}
    </div>
  );
}

export default App;
