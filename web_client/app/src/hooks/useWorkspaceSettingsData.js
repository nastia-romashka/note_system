import { useEffect, useState } from "react";

import {
  createWorkspaceInvite,
  fetchWorkspaceMembers,
  fetchWorkspaceOverview,
  fetchWorkspaceSentInvites,
  updateWorkspaceMember,
} from "../api/workspacesApi";

export function useWorkspaceSettingsData({
  token,
  workspaceId,
  enabled,
  setMessage,
  uiPreview,
}) {
  const [overview, setOverview] = useState(uiPreview ? createPreviewOverview() : null);
  const [members, setMembers] = useState(uiPreview ? createPreviewMembers() : []);
  const [invites, setInvites] = useState(uiPreview ? createPreviewInvites() : []);
  const [memberDrafts, setMemberDrafts] = useState(() => createMemberDrafts(uiPreview ? createPreviewMembers() : []));
  const [loading, setLoading] = useState(false);
  const [inviteForm, setInviteForm] = useState(createEmptyInviteForm());

  useEffect(() => {
    if (!workspaceId) {
      setOverview(uiPreview ? createPreviewOverview() : null);
      setMembers(uiPreview ? createPreviewMembers() : []);
      setInvites(uiPreview ? createPreviewInvites() : []);
      setMemberDrafts(createMemberDrafts(uiPreview ? createPreviewMembers() : []));
      setInviteForm(createEmptyInviteForm());
      return;
    }

    if (!enabled) {
      return;
    }

    void loadWorkspaceSettings();
  }, [enabled, token, workspaceId, uiPreview]);

  async function loadWorkspaceSettings() {
    if (!workspaceId) {
      return;
    }

    if (uiPreview) {
      const previewMembers = createPreviewMembers();
      setOverview(createPreviewOverview());
      setMembers(previewMembers);
      setInvites(createPreviewInvites());
      setMemberDrafts(createMemberDrafts(previewMembers));
      return;
    }

    if (!token) {
      return;
    }

    try {
      setLoading(true);
      const nextOverview = await fetchWorkspaceOverview(token, workspaceId);
      const [nextMembers, nextInvites] = await Promise.all([
        fetchWorkspaceMembers(token, workspaceId),
        nextOverview?.can_invite ? fetchWorkspaceSentInvites(token, workspaceId) : Promise.resolve([]),
      ]);

      setOverview(nextOverview || null);
      setMembers(Array.isArray(nextMembers) ? nextMembers : []);
      setInvites(Array.isArray(nextInvites) ? nextInvites : []);
      setMemberDrafts(createMemberDrafts(Array.isArray(nextMembers) ? nextMembers : []));
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось загрузить настройки пространства.",
      });
    } finally {
      setLoading(false);
    }
  }

  async function submitInvite() {
    if (uiPreview) {
      const email = inviteForm.email.trim().toLowerCase();
      const role = inviteForm.role || "viewer";
      if (!email) {
        setMessage({ type: "warning", text: "Введите email для приглашения." });
        return false;
      }

      setInvites((current) => [
        {
          uuid: `preview-invite-${Date.now()}`,
          workspace_name: overview?.workspace?.name || "Пространство",
          email,
          role,
          invited_by_username: "you",
          created_at: Math.floor(Date.now() / 1000),
          expires_at: Math.floor(Date.now() / 1000) + 7 * 24 * 60 * 60,
          status: "pending",
        },
        ...current,
      ]);
      setInviteForm(createEmptyInviteForm());
      setMessage({ type: "success", text: `Приглашение для ${email} подготовлено.` });
      return true;
    }

    if (!token || !workspaceId) {
      return false;
    }

    const email = inviteForm.email.trim().toLowerCase();
    const role = inviteForm.role || "viewer";
    if (!email) {
      setMessage({ type: "warning", text: "Введите email для приглашения." });
      return false;
    }

    try {
      setLoading(true);
      const invite = await createWorkspaceInvite(token, workspaceId, { email, role });
      setInvites((current) => [invite, ...current.filter((item) => item.uuid !== invite?.uuid)]);
      setInviteForm(createEmptyInviteForm());
      setMessage({ type: "success", text: `Приглашение отправлено на ${email}.` });
      return true;
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось отправить приглашение.",
      });
      return false;
    } finally {
      setLoading(false);
    }
  }

  function updateMemberDraft(memberUserId, patch) {
    setMemberDrafts((current) => ({
      ...current,
      [memberUserId]: {
        ...(current[memberUserId] || { role: "viewer", status: "active" }),
        ...patch,
      },
    }));
  }

  async function submitMemberUpdate(memberUserId) {
    const draft = memberDrafts[memberUserId];
    const member = members.find((item) => item.user_uuid === memberUserId);

    if (!draft || !member) {
      return false;
    }

    if (uiPreview) {
      setMembers((current) =>
        current
          .map((item) =>
            item.user_uuid === memberUserId
              ? {
                  ...item,
                  role: draft.role || item.role,
                  status: draft.status || item.status,
                }
              : item,
          )
          .filter((item) => item.status !== "removed"),
      );
      setMessage({ type: "success", text: "Статус участника обновлен." });
      return true;
    }

    if (!token || !workspaceId) {
      return false;
    }

    try {
      setLoading(true);
      await updateWorkspaceMember(token, workspaceId, memberUserId, {
        role: draft.role,
        status: draft.status,
      });
      await loadWorkspaceSettings();
      setMessage({ type: "success", text: "Участник пространства обновлен." });
      return true;
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Не удалось обновить участника пространства.",
      });
      return false;
    } finally {
      setLoading(false);
    }
  }

  return {
    overview,
    members,
    invites,
    memberDrafts,
    loading,
    inviteForm,
    setInviteForm,
    updateMemberDraft,
    refreshWorkspaceSettings: loadWorkspaceSettings,
    submitMemberUpdate,
    submitInvite,
  };
}

function createEmptyInviteForm() {
  return {
    email: "",
    role: "viewer",
  };
}

function createPreviewOverview() {
  return {
    workspace: {
      uuid: "workspace-design",
      name: "Дизайн-команда",
      visibility: "invite_only",
      created_at: Math.floor(Date.now() / 1000) - 15 * 24 * 60 * 60,
    },
    role: "owner",
    status: "active",
    can_invite: true,
    members_count: 4,
    stats: {
      categories_count: 6,
      notes_count: 18,
      tags_count: 11,
      files_count: 9,
      last_activity_at: null,
    },
    upcoming_events: [],
  };
}

function createPreviewMembers() {
  const now = Math.floor(Date.now() / 1000);
  return [
    {
      user_uuid: "owner-1",
      username: "nastia",
      email: "nastia@example.com",
      role: "owner",
      status: "active",
      joined_at: now - 15 * 24 * 60 * 60,
    },
    {
      user_uuid: "editor-1",
      username: "alisa",
      email: "alisa@example.com",
      role: "editor",
      status: "active",
      joined_at: now - 12 * 24 * 60 * 60,
    },
    {
      user_uuid: "viewer-1",
      username: "maria",
      email: "maria@example.com",
      role: "viewer",
      status: "active",
      joined_at: now - 8 * 24 * 60 * 60,
    },
  ];
}

function createPreviewInvites() {
  const now = Math.floor(Date.now() / 1000);
  return [
    {
      uuid: "invite-pending",
      workspace_name: "Дизайн-команда",
      email: "new.user@example.com",
      role: "viewer",
      invited_by_username: "nastia",
      created_at: now - 3600,
      expires_at: now + 6 * 24 * 60 * 60,
      status: "pending",
    },
    {
      uuid: "invite-accepted",
      workspace_name: "Дизайн-команда",
      email: "teammate@example.com",
      role: "editor",
      invited_by_username: "nastia",
      created_at: now - 4 * 24 * 60 * 60,
      expires_at: now + 3 * 24 * 60 * 60,
      accepted_at: now - 3 * 24 * 60 * 60,
      status: "accepted",
    },
  ];
}

function createMemberDrafts(members) {
  return (Array.isArray(members) ? members : []).reduce((accumulator, member) => {
    const key = member?.user_uuid;
    if (!key) {
      return accumulator;
    }

    accumulator[key] = {
      role: member.role || "viewer",
      status: member.status || "active",
    };
    return accumulator;
  }, {});
}
