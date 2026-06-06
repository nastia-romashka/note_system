import { useEffect, useState } from "react";

import {
  clearStoredSession,
  getStoredToken,
  persistStoredSession,
  SESSION_EVENT,
} from "../api/session";

export function useAuthSession(uiPreview = false) {
  const storedToken = uiPreview ? "" : getStoredToken();

  const [page, setPage] = useState(() => (uiPreview ? "login" : storedToken ? "notes" : "login"));
  const [token, setToken] = useState(() => (uiPreview ? "preview-token" : storedToken));

  useEffect(() => {
    if (uiPreview) {
      return undefined;
    }

    function syncSession(event) {
      const nextToken = event.detail?.token || getStoredToken();
      setToken(nextToken);
      setPage((currentPage) => {
        if (!nextToken) {
          return "login";
        }

        if (currentPage === "login" || currentPage === "signup") {
          return "notes";
        }

        return currentPage;
      });
    }

    window.addEventListener(SESSION_EVENT, syncSession);
    return () => window.removeEventListener(SESSION_EVENT, syncSession);
  }, [uiPreview]);

  function persistSession(nextToken, nextRefreshToken, remember = true) {
    setToken(nextToken);
    setPage("notes");

    if (uiPreview) {
      return;
    }

    persistStoredSession(nextToken, nextRefreshToken, remember);
  }

  function clearSession() {
    setToken("");
    setPage("login");

    if (uiPreview) {
      return;
    }

    clearStoredSession();
  }

  return {
    page,
    setPage,
    token,
    persistSession,
    clearSession,
  };
}
