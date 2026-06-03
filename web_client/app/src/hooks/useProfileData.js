import { useEffect, useState } from "react";

import { fetchProfileActions, fetchProfileSummary, updateProfile } from "../api/profileApi";

export function useProfileData({ token, enabled, setMessage }) {
  const [summary, setSummary] = useState(null);
  const [actions, setActions] = useState([]);
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileForm, setProfileForm] = useState(createEmptyProfileForm());

  useEffect(() => {
    setSummary(null);
    setActions([]);
    setProfileForm(createEmptyProfileForm());
  }, [token]);

  useEffect(() => {
    if (!summary?.profile) {
      return;
    }

    setProfileForm((current) => ({
      ...current,
      username: summary.profile.username || "",
      email: summary.profile.email || "",
    }));
  }, [summary]);

  useEffect(() => {
    if (!enabled || !token) {
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

  async function submitProfileUpdate() {
    if (!token) {
      return;
    }

    const username = profileForm.username.trim();
    const email = profileForm.email.trim();
    const currentPassword = profileForm.currentPassword.trim();
    const newPassword = profileForm.newPassword.trim();
    const confirmPassword = profileForm.confirmPassword.trim();

    if (!username) {
      setMessage({ type: "warning", text: "Введите имя пользователя." });
      return;
    }

    if (!email) {
      setMessage({ type: "warning", text: "Введите email." });
      return;
    }

    if (newPassword && !currentPassword) {
      setMessage({
        type: "warning",
        text: "Введите текущий пароль, чтобы задать новый.",
      });
      return;
    }

    if (newPassword && newPassword !== confirmPassword) {
      setMessage({
        type: "warning",
        text: "Подтверждение нового пароля не совпадает.",
      });
      return;
    }

    try {
      setProfileLoading(true);
      await updateProfile(token, {
        username,
        email,
        current_password: currentPassword,
        new_password: newPassword,
      });

      setMessage({ type: "success", text: "Профиль обновлен." });

      const [nextSummary, nextActions] = await Promise.all([
        fetchProfileSummary(token),
        fetchProfileActions(token, 50, 0),
      ]);

      setSummary(nextSummary);
      setActions(Array.isArray(nextActions) ? nextActions : []);
      setProfileForm((current) => ({
        ...current,
        username,
        email,
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
      }));
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось обновить профиль.",
      });
    } finally {
      setProfileLoading(false);
    }
  }

  return {
    summary,
    actions,
    profileLoading,
    profileForm,
    setProfileForm,
    refreshProfile: loadProfileData,
    submitProfileUpdate,
  };
}

function createEmptyProfileForm() {
  return {
    username: "",
    email: "",
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  };
}
