import { useEffect, useState } from "react";

import {
  clearStoredSession,
  getStoredToken,
  persistStoredSession,
  SESSION_EVENT,
} from "../api/session";
import { refreshSession } from "../api/authApi";

export function useAuthSession(uiPreview = false) {
  const [page, setPage] = useState(() => (uiPreview ? "login" : "login"));
  const [token, setToken] = useState(() => (uiPreview ? "preview-token" : getStoredToken()));
  const [isRestoring, setIsRestoring] = useState(() => !uiPreview);

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

  useEffect(() => {
    if (uiPreview) {
      return;
    }

    let cancelled = false;

    async function restoreSession() {
      try {
        const authData = await refreshSession();
        if (cancelled || !authData?.token) {
          return;
        }

        persistStoredSession(authData.token);
        setToken(authData.token);
        setPage("notes");
      } catch {
        if (!cancelled) {
          clearStoredSession();
          setToken("");
          setPage("login");
        }
      } finally {
        if (!cancelled) {
          setIsRestoring(false);
        }
      }
    }

    void restoreSession();

    return () => {
      cancelled = true;
    };
  }, [uiPreview]);

  function persistSession(nextToken, remember = true) {
    setToken(nextToken);
    setPage("notes");
    setIsRestoring(false);

    if (uiPreview) {
      return;
    }

    persistStoredSession(nextToken, remember);
  }

  function clearSession() {
    setToken("");
    setPage("login");
    setIsRestoring(false);

    if (uiPreview) {
      return;
    }

    clearStoredSession();
  }

  return {
    page,
    setPage,
    token,
    isRestoring,
    persistSession,
    clearSession,
  };
}
