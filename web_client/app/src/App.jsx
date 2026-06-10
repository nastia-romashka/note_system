import { useState } from "react";

import { logoutCurrentSession } from "./api/authApi";
import { useAuthSession } from "./hooks/useAuthSession";
import { useCalendarData } from "./hooks/useCalendarData";
import { useFilesData } from "./hooks/useFilesData";
import { useGraphData } from "./hooks/useGraphData";
import { useNotesData } from "./hooks/useNotesData";
import { useProfileData } from "./hooks/useProfileData";
import AuthPage from "./pages/AuthPage/AuthPage";
import CalendarPage from "./pages/CalendarPage/CalendarPage";
import GraphPage from "./pages/GraphPage/GraphPage";
import { ConfirmDialog, DuplicateNoteDialog, SubcategoryDialog } from "./pages/NotesPage/Dialogs";
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
  const calendarData = useCalendarData({
    token,
    categories: notesData.categories,
    uiPreview: UI_PREVIEW,
    setMessage,
    setLoading,
    onOpenNote: (note) => {
      notesData.handleOpenNote(note);
      setPage("notes");
    },
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

  function openTodayInCalendar() {
    const today = new Date();
    calendarData.setCurrentMonth(new Date(today.getFullYear(), today.getMonth(), 1));
    calendarData.setSelectedDay(new Date(today.getFullYear(), today.getMonth(), today.getDate()));
  }

  async function logout() {
    if (UI_PREVIEW) {
      clearSession();
      setMessage({ type: "info", text: "Preview mode: возврат на экран входа." });
      return;
    }

    try {
      await logoutCurrentSession();
    } catch {
      // Clear the local session even if the server-side revoke call fails.
    }

    clearSession();
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
            <button type="button" className={page === "calendar" ? "active" : ""} onClick={() => setPage("calendar")}>
              Календарь
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
      {notesData.duplicateDialog && (
        <DuplicateNoteDialog
          categories={notesData.categories}
          categoryUuid={notesData.duplicateDialog.categoryUuid}
          onCategoryChange={(value) =>
            notesData.setDuplicateDialog((current) => (current ? { ...current, categoryUuid: value } : current))
          }
          header={notesData.duplicateDialog.header}
          onHeaderChange={(value) =>
            notesData.setDuplicateDialog((current) => (current ? { ...current, header: value } : current))
          }
          shortBody={notesData.duplicateDialog.shortBody}
          onConfirm={() => void notesData.confirmDuplicateNote()}
          onCancel={notesData.closeDuplicateDialog}
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
          selectedCategory={notesData.selectedCategory}
          selectedCategoryId={notesData.selectedCategoryId}
          onSelectCategory={notesData.setSelectedCategoryId}
          onOpenSubcategoryDialog={notesData.openSubcategoryDialog}
          notes={notesData.filteredNotes}
          selectedNote={notesData.selectedNote}
          selectedNoteId={notesData.selectedNoteId}
          onSelectNote={notesData.handleSelectNote}
          onOpenProfile={() => setPage("profile")}
          onOpenGraph={() => setPage("graph")}
          onOpenCalendar={() => setPage("calendar")}
          search={notesData.search}
          onSearch={notesData.setSearch}
          searchResults={notesData.searchResults}
          onSelectSearchResult={notesData.handleSelectSearchResult}
          categoryForm={notesData.categoryForm}
          onCategoryFormChange={notesData.setCategoryForm}
          onCreateCategory={notesData.handleCreateCategory}
          onDeleteCategory={notesData.openDeleteCategoryDialog}
          categoryEditor={notesData.categoryEditor}
          onCategoryEditorChange={notesData.setCategoryEditor}
          onCloseCategoryEditor={notesData.closeCategoryEditor}
          onStartRenameCategory={notesData.startRenameCategory}
          onStartRecolorCategory={notesData.startRecolorCategory}
          onSubmitCategoryRename={notesData.submitCategoryRename}
          onSubmitCategoryColor={notesData.submitCategoryColor}
          noteForm={notesData.noteForm}
          onNoteFormChange={notesData.setNoteForm}
          noteEditorForm={notesData.noteEditorForm}
          onNoteEditorFormChange={notesData.setNoteEditorForm}
          parsedEditorTags={notesData.parsedEditorTags}
          onCreateNote={notesData.handleCreateNote}
          onUpdateNote={notesData.handleUpdateNote}
          onDeleteNote={notesData.handleDeleteNote}
          onOpenDuplicateDialog={notesData.openDuplicateDialog}
          files={filesData.files}
          onDownloadFile={filesData.handleDownloadFile}
          onDeleteFile={filesData.handleDeleteFile}
          onUploadFile={filesData.handleUploadFile}
          onFilePick={filesData.setFileDraft}
          fileInputRef={filesData.fileInputRef}
          fileDraft={filesData.fileDraft}
        />
      )}

      {page === "calendar" && (
        <CalendarPage
          currentMonth={calendarData.currentMonth}
          selectedDay={calendarData.selectedDay}
          notesByDay={calendarData.notesByDay}
          selectedDayNotes={calendarData.selectedDayNotes}
          categories={notesData.categories}
          createDialog={calendarData.createDialog}
          onCreateDialogChange={calendarData.setCreateDialog}
          onOpenCreateDialog={calendarData.openCreateDialog}
          onCloseCreateDialog={calendarData.closeCreateDialog}
          onConfirmCreateDialog={() => void calendarData.confirmCreateFromCalendar()}
          onSelectDay={calendarData.setSelectedDay}
          onChangeMonth={(offset) =>
            calendarData.setCurrentMonth(
              new Date(
                calendarData.currentMonth.getFullYear(),
                calendarData.currentMonth.getMonth() + offset,
                1,
              ),
            )
          }
          onToday={openTodayInCalendar}
          onOpenNote={calendarData.openNote}
          onOpenGraph={() => setPage("graph")}
          onOpenNotes={() => setPage("notes")}
          onOpenProfile={() => setPage("profile")}
        />
      )}

      {page === "profile" && (
        <ProfilePage
          summary={profileData.summary}
          actions={profileData.actions}
          loading={profileData.profileLoading}
          profileForm={profileData.profileForm}
          onProfileFormChange={profileData.setProfileForm}
          onSubmitProfileUpdate={() => void profileData.submitProfileUpdate()}
          onRefresh={profileData.refreshProfile}
          onBackToNotes={() => setPage("notes")}
          onOpenGraph={() => setPage("graph")}
          onOpenCalendar={() => setPage("calendar")}
          onLogout={logout}
        />
      )}

      {page === "graph" && (
        <GraphPage
          graph={graphData.graph}
          loading={graphData.graphLoading}
          onCreateGraphLink={graphData.createUserGraphLink}
          onDeleteGraphLink={graphData.deleteUserGraphLink}
          onBackToNotes={() => setPage("notes")}
          onOpenCalendar={() => setPage("calendar")}
          onOpenGraphNode={(node) => {
            notesData.handleOpenGraphNode(node);
            setPage("notes");
          }}
          onOpenProfile={() => setPage("profile")}
        />
      )}
    </div>
  );
}

export default App;
