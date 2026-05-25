import { useState } from "react";

const TOKEN_KEY = "note-system-token";
const REFRESH_TOKEN_KEY = "note-system-refresh-token";

export function useAuthSession(uiPreview = false) {
  const storedToken = uiPreview ? "" : localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY) || "";

  const [page, setPage] = useState(() => (uiPreview ? "login" : storedToken ? "notes" : "login"));
  const [token, setToken] = useState(() => (uiPreview ? "preview-token" : storedToken));

  function persistSession(nextToken, nextRefreshToken, remember = true) {
    setToken(nextToken);
    setPage("notes");

    if (uiPreview) {
      return;
    }

    if (remember) {
      localStorage.setItem(TOKEN_KEY, nextToken);
      localStorage.setItem(REFRESH_TOKEN_KEY, nextRefreshToken);
      sessionStorage.removeItem(TOKEN_KEY);
      sessionStorage.removeItem(REFRESH_TOKEN_KEY);
      return;
    }

    sessionStorage.setItem(TOKEN_KEY, nextToken);
    sessionStorage.setItem(REFRESH_TOKEN_KEY, nextRefreshToken);
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  }

  function clearSession() {
    setPage("login");

    if (uiPreview) {
      return;
    }

    setToken("");
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    sessionStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
  }

  return {
    page,
    setPage,
    token,
    persistSession,
    clearSession,
  };
}
