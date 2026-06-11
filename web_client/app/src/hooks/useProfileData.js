import { useEffect, useState } from "react";

import { fetchProfileActions, fetchProfileSummary, updateProfile } from "../api/profileApi";
import {
  acceptWorkspaceInvite as acceptWorkspaceInviteRequest,
  declineWorkspaceInvite as declineWorkspaceInviteRequest,
  fetchWorkspaceInvites,
} from "../api/workspacesApi";

export function useProfileData({ token, enabled, setMessage, uiPreview, onWorkspaceMembershipChange }) {
  const [summary, setSummary] = useState(null);
  const [actions, setActions] = useState([]);
  const [workspaceInvites, setWorkspaceInvites] = useState(() => (uiPreview ? createPreviewWorkspaceInvites() : []));
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileForm, setProfileForm] = useState(createEmptyProfileForm());

  useEffect(() => {
    setSummary(null);
    setActions([]);
    setWorkspaceInvites(uiPreview ? createPreviewWorkspaceInvites() : []);
    setProfileForm(createEmptyProfileForm());
  }, [token, uiPreview]);

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
      const [nextSummary, nextActions, nextInvites] = await Promise.all([
        fetchProfileSummary(token),
        fetchProfileActions(token, 50, 0),
        uiPreview ? Promise.resolve(createPreviewWorkspaceInvites()) : fetchWorkspaceInvites(token),
      ]);

      setSummary(nextSummary);
      setActions(Array.isArray(nextActions) ? nextActions : []);
      setWorkspaceInvites(Array.isArray(nextInvites) ? nextInvites.map(normalizeWorkspaceInvite) : []);
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
      await loadProfileData();
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

  async function acceptWorkspaceInvite(inviteId) {
    const invite = workspaceInvites.find((item) => item.id === inviteId || item.uuid === inviteId);
    if (!invite) {
      return;
    }

    if (uiPreview) {
      setWorkspaceInvites((current) => current.filter((item) => (item.id || item.uuid) !== inviteId));
      setMessage({
        type: "success",
        text: `Приглашение в пространство "${invite.workspace_name}" принято.`,
      });
      return;
    }

    if (!token) {
      return;
    }

    try {
      setProfileLoading(true);
      const workspace = await acceptWorkspaceInviteRequest(token, invite.uuid || invite.id);
      setWorkspaceInvites((current) => current.filter((item) => (item.id || item.uuid) !== inviteId));
      onWorkspaceMembershipChange?.(workspace?.uuid || "");
      setMessage({
        type: "success",
        text: `Приглашение в пространство "${invite.workspace_name}" принято.`,
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось принять приглашение.",
      });
    } finally {
      setProfileLoading(false);
    }
  }

  async function declineWorkspaceInvite(inviteId) {
    const invite = workspaceInvites.find((item) => item.id === inviteId || item.uuid === inviteId);
    if (!invite) {
      return;
    }

    if (uiPreview) {
      setWorkspaceInvites((current) => current.filter((item) => (item.id || item.uuid) !== inviteId));
      setMessage({
        type: "info",
        text: `Приглашение в пространство "${invite.workspace_name}" отклонено.`,
      });
      return;
    }

    if (!token) {
      return;
    }

    try {
      setProfileLoading(true);
      await declineWorkspaceInviteRequest(token, invite.uuid || invite.id);
      setWorkspaceInvites((current) => current.filter((item) => (item.id || item.uuid) !== inviteId));
      setMessage({
        type: "info",
        text: `Приглашение в пространство "${invite.workspace_name}" отклонено.`,
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось отклонить приглашение.",
      });
    } finally {
      setProfileLoading(false);
    }
  }

  return {
    summary,
    actions,
    workspaceInvites,
    profileLoading,
    profileForm,
    setProfileForm,
    refreshProfile: loadProfileData,
    submitProfileUpdate,
    acceptWorkspaceInvite,
    declineWorkspaceInvite,
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

function createPreviewWorkspaceInvites() {
  return [
    {
      id: "invite-design-team",
      uuid: "invite-design-team",
      workspace_name: "Дизайн-команда",
      invited_by: "alisa",
      role: "editor",
      created_at: Math.floor(Date.now() / 1000) - 3600 * 5,
      message: "Нужна совместная работа над структурой заметок и шаблонами.",
    },
    {
      id: "invite-family-plans",
      uuid: "invite-family-plans",
      workspace_name: "Семейные планы",
      invited_by: "maria",
      role: "member",
      created_at: Math.floor(Date.now() / 1000) - 3600 * 28,
      message: "Добавили общее пространство для календаря, покупок и маршрутов.",
    },
  ];
}

function normalizeWorkspaceInvite(invite) {
  return {
    id: invite?.id || invite?.uuid || "",
    uuid: invite?.uuid || invite?.id || "",
    workspace_name: invite?.workspace_name || "Пространство",
    invited_by: invite?.invited_by || invite?.invited_by_username || "",
    role: invite?.role || "member",
    created_at: invite?.created_at || 0,
    message: invite?.message || "",
  };
}
