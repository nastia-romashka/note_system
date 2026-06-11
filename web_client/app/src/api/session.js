const TOKEN_KEY = "note-system-token";
const REFRESH_TOKEN_KEY = "note-system-refresh-token";
const WORKSPACE_KEY = "note-system-workspace-id";

export const SESSION_EVENT = "note-system-session-change";

function dispatchSessionEvent(type) {
  window.dispatchEvent(
    new CustomEvent(SESSION_EVENT, {
      detail: {
        type,
        token: getStoredToken(),
        refreshToken: getStoredRefreshToken(),
      },
    }),
  );
}

export function getStoredToken() {
  return localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY) || "";
}

export function getStoredRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY) || sessionStorage.getItem(REFRESH_TOKEN_KEY) || "";
}

export function getStoredWorkspaceId() {
  return localStorage.getItem(WORKSPACE_KEY) || sessionStorage.getItem(WORKSPACE_KEY) || "";
}

export function shouldRememberSession() {
  return localStorage.getItem(TOKEN_KEY) !== null || localStorage.getItem(REFRESH_TOKEN_KEY) !== null;
}

export function persistStoredSession(token, refreshToken, remember = true) {
  if (remember) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    sessionStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
  } else {
    sessionStorage.setItem(TOKEN_KEY, token);
    sessionStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
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
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  sessionStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(REFRESH_TOKEN_KEY);
  clearStoredWorkspaceId();
  dispatchSessionEvent("clear");
}
