const REMEMBER_KEY = "note-system-session-remember";
const WORKSPACE_KEY = "note-system-workspace-id";

export const SESSION_EVENT = "note-system-session-change";

let inMemoryToken = "";

function dispatchSessionEvent(type) {
  window.dispatchEvent(
    new CustomEvent(SESSION_EVENT, {
      detail: {
        type,
        token: getStoredToken(),
        remember: shouldRememberSession(),
      },
    }),
  );
}

export function getStoredToken() {
  return inMemoryToken;
}

export function getStoredWorkspaceId() {
  return localStorage.getItem(WORKSPACE_KEY) || sessionStorage.getItem(WORKSPACE_KEY) || "";
}

export function shouldRememberSession() {
  return localStorage.getItem(REMEMBER_KEY) !== null || sessionStorage.getItem(REMEMBER_KEY) !== null;
}

export function persistStoredSession(token, remember = shouldRememberSession()) {
  inMemoryToken = token;

  if (remember) {
    localStorage.setItem(REMEMBER_KEY, "1");
    sessionStorage.removeItem(REMEMBER_KEY);
  } else {
    sessionStorage.setItem(REMEMBER_KEY, "1");
    localStorage.removeItem(REMEMBER_KEY);
  }

  dispatchSessionEvent("persist");
}

export function persistStoredWorkspaceId(workspaceId, remember = shouldRememberSession()) {
  if (!workspaceId) {
    clearStoredWorkspaceId();
    return;
  }

  if (remember) {
    localStorage.setItem(WORKSPACE_KEY, workspaceId);
    sessionStorage.removeItem(WORKSPACE_KEY);
  } else {
    sessionStorage.setItem(WORKSPACE_KEY, workspaceId);
    localStorage.removeItem(WORKSPACE_KEY);
  }
}

export function clearStoredWorkspaceId() {
  localStorage.removeItem(WORKSPACE_KEY);
  sessionStorage.removeItem(WORKSPACE_KEY);
}

export function clearStoredSession() {
  inMemoryToken = "";
  localStorage.removeItem(REMEMBER_KEY);
  sessionStorage.removeItem(REMEMBER_KEY);
  clearStoredWorkspaceId();
  dispatchSessionEvent("clear");
}
