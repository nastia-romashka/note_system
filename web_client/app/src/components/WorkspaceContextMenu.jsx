import { useEffect, useState } from "react";
import { createPortal } from "react-dom";

export default function WorkspaceContextMenu({
  currentWorkspace,
  workspaces,
  onSelectPersonalWorkspace,
  onSelectWorkspace,
  onCreateWorkspace,
}) {
  const [isWorkspaceMenuOpen, setIsWorkspaceMenuOpen] = useState(false);
  const [isCreateWorkspaceOpen, setIsCreateWorkspaceOpen] = useState(false);
  const [workspaceDraft, setWorkspaceDraft] = useState("");
  const [isPortalReady, setIsPortalReady] = useState(false);

  const isWorkspaceMode = Boolean(currentWorkspace);

  useEffect(() => {
    setIsPortalReady(true);
  }, []);

  useEffect(() => {
    if (!isWorkspaceMenuOpen) {
      return undefined;
    }

    function handleKeyDown(event) {
      if (event.key === "Escape") {
        setIsWorkspaceMenuOpen(false);
        setIsCreateWorkspaceOpen(false);
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isWorkspaceMenuOpen]);

  function closeWorkspaceMenu() {
    setIsWorkspaceMenuOpen(false);
    setIsCreateWorkspaceOpen(false);
  }

  async function handleCreateWorkspace(event) {
    event.preventDefault();
    const wasCreated = await Promise.resolve(onCreateWorkspace(workspaceDraft));
    if (!wasCreated) {
      return;
    }

    setWorkspaceDraft("");
    closeWorkspaceMenu();
  }

  function handleSelectPersonalMode() {
    onSelectPersonalWorkspace();
    closeWorkspaceMenu();
  }

  function handleSelectWorkspaceMode(workspaceId) {
    onSelectWorkspace(workspaceId);
    closeWorkspaceMenu();
  }

  return (
    <div className="workspace-menu-shell">
      <button
        className={`workspace-menu-trigger ${isWorkspaceMenuOpen ? "active" : ""}`}
        type="button"
        aria-label="Меню пространств"
        aria-expanded={isWorkspaceMenuOpen}
        onClick={() => setIsWorkspaceMenuOpen((current) => !current)}
      >
        <span />
        <span />
        <span />
      </button>

      {isWorkspaceMenuOpen && isPortalReady
        ? createPortal(
            <>
              <button
                className="workspace-menu-backdrop"
                type="button"
                aria-label="Закрыть меню пространств"
                onClick={closeWorkspaceMenu}
              />
              <div className="workspace-menu-panel" role="dialog" aria-modal="true" aria-label="Пространства">
                <div className="workspace-menu-header">
                  <div>
                    <div className="card-label">Режим работы</div>
                    <strong>Пространства</strong>
                  </div>
                  <button
                    className="workspace-add-button"
                    type="button"
                    aria-label="Создать пространство"
                    onClick={() => setIsCreateWorkspaceOpen((current) => !current)}
                  >
                    +
                  </button>
                </div>

                {isCreateWorkspaceOpen && (
                  <form className="workspace-create-form" onSubmit={handleCreateWorkspace}>
                    <input
                      value={workspaceDraft}
                      onChange={(event) => setWorkspaceDraft(event.target.value)}
                      placeholder="Название пространства"
                    />
                    <button className="primary-button" type="submit">
                      Создать
                    </button>
                  </form>
                )}

                <div className="workspace-menu-section">
                  <button
                    className={`workspace-option ${isWorkspaceMode ? "" : "active"}`}
                    type="button"
                    onClick={handleSelectPersonalMode}
                  >
                    <span>Личное пространство</span>
                  </button>
                </div>

                <div className="workspace-menu-section">
                  <div className="workspace-section-title">Рабочие пространства</div>
                  {workspaces.length ? (
                    <div className="workspace-option-list">
                      {workspaces.map((workspace) => (
                        <button
                          key={workspace.id}
                          className={`workspace-option ${currentWorkspace?.id === workspace.id ? "active" : ""}`}
                          type="button"
                          onClick={() => handleSelectWorkspaceMode(workspace.id)}
                        >
                          <span>{workspace.name}</span>
                        </button>
                      ))}
                    </div>
                  ) : (
                    <div className="empty-copy workspace-empty-copy">
                      Рабочих пространств пока нет. Нажмите <strong>+</strong>, чтобы создать первое.
                    </div>
                  )}
                </div>
              </div>
            </>,
            document.body,
          )
        : null}
    </div>
  );
}
