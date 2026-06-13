import { authHeaders, readSafeJson, request, requestRaw } from "./http";

export function fetchWorkspaces(token) {
  return request("/api/me/workspaces", {
    headers: authHeaders(token),
  });
}

export function fetchWorkspaceInvites(token) {
  return request("/api/me/workspace-invites", {
    headers: authHeaders(token),
  });
}

export function fetchWorkspaceOverview(token, workspaceId) {
  return request(`/api/workspaces/${workspaceId}`, {
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export function fetchWorkspaceMembers(token, workspaceId) {
  return request(`/api/workspaces/${workspaceId}/members`, {
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export async function updateWorkspaceMember(token, workspaceId, memberUserId, payload) {
  const response = await requestRaw(`/api/workspaces/${workspaceId}/members/${memberUserId}`, {
    method: "PATCH",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось обновить участника пространства.");
  }

  const text = await response.text();
  return text.trim() ? JSON.parse(text) : null;
}

export function fetchWorkspaceSentInvites(token, workspaceId) {
  return request(`/api/workspaces/${workspaceId}/invites`, {
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export async function createWorkspace(token, payload) {
  const response = await requestRaw("/api/workspaces", {
    method: "POST",
    headers: {
      ...authHeaders(token),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось создать пространство.");
  }

  const location = response.headers.get("Location") || "";
  return location.split("/").filter(Boolean).at(-1) || "";
}

export async function createWorkspaceInvite(token, workspaceId, payload) {
  const response = await requestRaw(`/api/workspaces/${workspaceId}/invites`, {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось отправить приглашение.");
  }

  const text = await response.text();
  return text.trim() ? JSON.parse(text) : null;
}

export async function leaveWorkspace(token, workspaceId) {
  const response = await requestRaw(`/api/workspaces/${workspaceId}/leave`, {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({}),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось выйти из пространства.");
  }

  return null;
}

export async function deleteWorkspace(token, workspaceId) {
  const response = await requestRaw(`/api/workspaces/${workspaceId}`, {
    method: "DELETE",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({}),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось удалить пространство.");
  }

  return null;
}

export async function acceptWorkspaceInvite(token, inviteId) {
  const response = await requestRaw(`/api/workspaces/invites/${inviteId}/accept`, {
    method: "POST",
    headers: {
      ...authHeaders(token),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({}),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось принять приглашение.");
  }

  const text = await response.text();
  return text.trim() ? JSON.parse(text) : null;
}

export async function declineWorkspaceInvite(token, inviteId) {
  const response = await requestRaw(`/api/workspaces/invites/${inviteId}/decline`, {
    method: "POST",
    headers: {
      ...authHeaders(token),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({}),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Не удалось отклонить приглашение.");
  }

  return null;
}
