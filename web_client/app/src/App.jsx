import { useEffect, useState } from "react";

import { logoutCurrentSession } from "./api/authApi";
import { clearStoredWorkspaceId, getStoredWorkspaceId, persistStoredWorkspaceId } from "./api/session";
import { createWorkspace as createWorkspaceRequest, fetchWorkspaces } from "./api/workspacesApi";
import { useAuthSession } from "./hooks/useAuthSession";
import { useCalendarData } from "./hooks/useCalendarData";
import { useFilesData } from "./hooks/useFilesData";
import { useGraphData } from "./hooks/useGraphData";
import { useNotesData } from "./hooks/useNotesData";
import { useProfileData } from "./hooks/useProfileData";
import { useWorkspaceSettingsData } from "./hooks/useWorkspaceSettingsData";
import AuthPage from "./pages/AuthPage/AuthPage";
import CalendarPage from "./pages/CalendarPage/CalendarPage";
import GraphPage from "./pages/GraphPage/GraphPage";
import { ConfirmDialog, DuplicateNoteDialog, SubcategoryDialog } from "./pages/NotesPage/Dialogs";
import { FlashMessage } from "./pages/NotesPage/Messages";
import NotesPage from "./pages/NotesPage/NotesPage";
import ProfilePage from "./pages/ProfilePage/ProfilePage";
import WorkspaceSettingsPage from "./pages/WorkspaceSettingsPage/WorkspaceSettingsPage";

const UI_PREVIEW = import.meta.env.VITE_UI_PREVIEW === "true";

const PREVIEW_WORKSPACES = [
  { id: "workspace-design", name: "Дизайн-команда", isPersonal: false },
  { id: "workspace-family", name: "Семейные планы", isPersonal: false },
];

function App() {
  const [message, setMessage] = useState(null);
  const [loading, setLoading] = useState(false);
  const [workspaces, setWorkspaces] = useState(() => (UI_PREVIEW ? PREVIEW_WORKSPACES : []));
  const [currentWorkspaceId, setCurrentWorkspaceId] = useState(() =>
    UI_PREVIEW ? PREVIEW_WORKSPACES[0]?.id || null : getStoredWorkspaceId() || null,
  );
  const { page, setPage, token, persistSession, clearSession } = useAuthSession(UI_PREVIEW);
  const activeWorkspaceId = currentWorkspaceId || "";

  const notesData = useNotesData({
    token,
    workspaceId: activeWorkspaceId,
    uiPreview: UI_PREVIEW,
    setMessage,
    setLoading,
  });
  const calendarData = useCalendarData({
    token,
    workspaceId: activeWorkspaceId,
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
    workspaceId: activeWorkspaceId,
    selectedNoteId: notesData.selectedNoteId,
    uiPreview: UI_PREVIEW,
    setMessage,
    setLoading,
  });
  const profileData = useProfileData({
    token,
    enabled: page === "profile",
    setMessage,
    uiPreview: UI_PREVIEW,
    onWorkspaceMembershipChange: (workspaceId) => {
      void loadWorkspaces(workspaceId);
    },
  });
  const graphData = useGraphData({
    token,
    workspaceId: activeWorkspaceId,
    enabled: page === "graph",
    setMessage,
  });
  const workspaceSettingsData = useWorkspaceSettingsData({
    token,
    workspaceId: activeWorkspaceId,
    enabled: page === "workspace-settings",
    setMessage,
    uiPreview: UI_PREVIEW,
  });

  const currentWorkspace = workspaces.find((workspace) => workspace.id === currentWorkspaceId) || null;

  useEffect(() => {
    if (UI_PREVIEW) {
      return;
    }

    if (!token) {
      setWorkspaces([]);
      setCurrentWorkspaceId(null);
      clearStoredWorkspaceId();
      return;
    }

    void loadWorkspaces(getStoredWorkspaceId() || currentWorkspaceId || "");
  }, [token]);

  useEffect(() => {
    if (UI_PREVIEW) {
      return;
    }

    if (!currentWorkspaceId) {
      clearStoredWorkspaceId();
      return;
    }

    persistStoredWorkspaceId(currentWorkspaceId);
  }, [currentWorkspaceId]);

  async function loadWorkspaces(preferredWorkspaceId = "") {
    if (UI_PREVIEW || !token) {
      return [];
    }

    try {
      const list = await fetchWorkspaces(token);
      const normalized = Array.isArray(list) ? list.map(normalizeWorkspace).filter((workspace) => !workspace.isPersonal) : [];
      setWorkspaces(normalized);

      const requestedWorkspaceId = preferredWorkspaceId || currentWorkspaceId || getStoredWorkspaceId();
      if (requestedWorkspaceId && normalized.some((workspace) => workspace.id === requestedWorkspaceId)) {
        setCurrentWorkspaceId(requestedWorkspaceId);
      } else {
        setCurrentWorkspaceId(null);
      }

      return normalized;
    } catch (error) {
      setWorkspaces([]);
      setCurrentWorkspaceId(null);
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось загрузить пространства.",
      });
      return [];
    }
  }

  function openTodayInCalendar() {
    const today = new Date();
    calendarData.setCurrentMonth(new Date(today.getFullYear(), today.getMonth(), 1));
    calendarData.setSelectedDay(new Date(today.getFullYear(), today.getMonth(), today.getDate()));
  }

  function selectPersonalWorkspace() {
    notesData.resetNotesState();
    filesData.resetFilesState();
    setCurrentWorkspaceId(null);
  }

  function selectWorkspace(workspaceId) {
    notesData.resetNotesState();
    filesData.resetFilesState();
    setCurrentWorkspaceId(workspaceId);
  }

  async function createWorkspace(name) {
    const trimmedName = name.trim();
    if (!trimmedName) {
      setMessage({ type: "error", text: "Введите название нового пространства." });
      return false;
    }

    if (UI_PREVIEW) {
      const nextWorkspace = {
        id: `workspace-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`,
        name: trimmedName,
        isPersonal: false,
      };

      setWorkspaces((current) => [...current, nextWorkspace]);
      setCurrentWorkspaceId(nextWorkspace.id);
      setMessage({ type: "success", text: `Пространство "${trimmedName}" создано.` });
      return true;
    }

    if (!token) {
      setMessage({ type: "error", text: "Сессия недоступна. Войдите заново." });
      return false;
    }

    try {
      setLoading(true);
      setMessage(null);

      const createdWorkspaceId = await createWorkspaceRequest(token, {
        name: trimmedName,
        visibility: "invite_only",
      });
      const refreshed = await loadWorkspaces(createdWorkspaceId);
      const selectedWorkspaceId =
        createdWorkspaceId ||
        refreshed.at(-1)?.id ||
        "";

      notesData.resetNotesState();
      filesData.resetFilesState();
      setCurrentWorkspaceId(selectedWorkspaceId || null);
      setPage("notes");
      setMessage({ type: "success", text: `Пространство "${trimmedName}" создано.` });
      return true;
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось создать пространство.",
      });
      return false;
    } finally {
      setLoading(false);
    }
  }

  function openContextSettings() {
    if (!currentWorkspace) {
      setPage("profile");
      return;
    }

    setPage("workspace-settings");
    return;

    setMessage({
      type: "info",
      text: `Настройки пространства "${currentWorkspace.name}" добавим следующим этапом.`,
    });
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
    setWorkspaces([]);
    setCurrentWorkspaceId(null);
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
          currentWorkspace={currentWorkspace}
          workspaces={workspaces}
          categories={notesData.categories}
          selectedCategory={notesData.selectedCategory}
          selectedCategoryId={notesData.selectedCategoryId}
          onSelectCategory={notesData.setSelectedCategoryId}
          onOpenSubcategoryDialog={notesData.openSubcategoryDialog}
          notes={notesData.filteredNotes}
          selectedNote={notesData.selectedNote}
          selectedNoteId={notesData.selectedNoteId}
          onSelectNote={notesData.handleSelectNote}
          onOpenProfile={openContextSettings}
          onOpenGraph={() => setPage("graph")}
          onOpenCalendar={() => setPage("calendar")}
          onSelectPersonalWorkspace={selectPersonalWorkspace}
          onSelectWorkspace={selectWorkspace}
          onCreateWorkspace={createWorkspace}
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
          currentWorkspace={currentWorkspace}
          workspaces={workspaces}
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
          onOpenProfile={openContextSettings}
          onSelectPersonalWorkspace={selectPersonalWorkspace}
          onSelectWorkspace={selectWorkspace}
          onCreateWorkspace={createWorkspace}
        />
      )}

      {page === "profile" && (
        <ProfilePage
          summary={profileData.summary}
          actions={profileData.actions}
          workspaceInvites={profileData.workspaceInvites}
          loading={profileData.profileLoading}
          profileForm={profileData.profileForm}
          onProfileFormChange={profileData.setProfileForm}
          onSubmitProfileUpdate={() => void profileData.submitProfileUpdate()}
          onAcceptWorkspaceInvite={(inviteId) => void profileData.acceptWorkspaceInvite(inviteId)}
          onDeclineWorkspaceInvite={(inviteId) => void profileData.declineWorkspaceInvite(inviteId)}
          onRefresh={profileData.refreshProfile}
          onBackToNotes={() => setPage("notes")}
          onOpenGraph={() => setPage("graph")}
          onOpenCalendar={() => setPage("calendar")}
          onLogout={logout}
        />
      )}

      {page === "graph" && (
        <GraphPage
          currentWorkspace={currentWorkspace}
          workspaces={workspaces}
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
          onOpenProfile={openContextSettings}
          onSelectPersonalWorkspace={selectPersonalWorkspace}
          onSelectWorkspace={selectWorkspace}
          onCreateWorkspace={createWorkspace}
        />
      )}

      {page === "workspace-settings" && (
        <WorkspaceSettingsPage
          currentWorkspace={currentWorkspace}
          workspaces={workspaces}
          overview={workspaceSettingsData.overview}
          members={workspaceSettingsData.members}
          invites={workspaceSettingsData.invites}
          memberDrafts={workspaceSettingsData.memberDrafts}
          loading={workspaceSettingsData.loading}
          inviteForm={workspaceSettingsData.inviteForm}
          onInviteFormChange={workspaceSettingsData.setInviteForm}
          onMemberDraftChange={workspaceSettingsData.updateMemberDraft}
          onSubmitMemberUpdate={(memberUserId) => workspaceSettingsData.submitMemberUpdate(memberUserId)}
          onSubmitInvite={() => workspaceSettingsData.submitInvite()}
          onRefresh={workspaceSettingsData.refreshWorkspaceSettings}
          onOpenNotes={() => setPage("notes")}
          onOpenGraph={() => setPage("graph")}
          onOpenCalendar={() => setPage("calendar")}
          onSelectPersonalWorkspace={selectPersonalWorkspace}
          onSelectWorkspace={selectWorkspace}
          onCreateWorkspace={createWorkspace}
        />
      )}
    </div>
  );
}

function normalizeWorkspace(workspace) {
  return {
    id: workspace?.uuid || workspace?.id || "",
    name: workspace?.name || "Пространство",
    isPersonal: Boolean(workspace?.is_personal),
    visibility: workspace?.visibility || "",
  };
}

export default App;
