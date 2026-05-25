import { useEffect, useState } from "react";

import { fetchProfileActions, fetchProfileSummary } from "../api/profileApi";

export function useProfileData({ token, enabled, setMessage }) {
  const [summary, setSummary] = useState(null);
  const [actions, setActions] = useState([]);
  const [profileLoading, setProfileLoading] = useState(false);

  useEffect(() => {
    setSummary(null);
    setActions([]);
  }, [token]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    if (!token) {
      return;
    }

    void loadProfileData();
  }, [enabled, token]);

  async function loadProfileData() {
    if (!token) {
      return;
    }

    try {
      setProfileLoading(true);
      const [nextSummary, nextActions] = await Promise.all([
        fetchProfileSummary(token),
        fetchProfileActions(token, 50, 0),
      ]);
      setSummary(nextSummary);
      setActions(Array.isArray(nextActions) ? nextActions : []);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось загрузить личный кабинет.",
      });
    } finally {
      setProfileLoading(false);
    }
  }

  return {
    summary,
    actions,
    profileLoading,
    refreshProfile: loadProfileData,
  };
}
