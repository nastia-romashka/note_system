import { authHeaders, request, requestRaw } from "./http";

export function fetchFiles(token, noteId, workspaceId = "") {
  return request(`/api/notes/${noteId}/files`, { headers: authHeaders(token, "application/json", workspaceId) });
}

export function uploadFile(token, noteId, file, workspaceId = "") {
  const formData = new FormData();
  formData.append("file", file);

  return request(`/api/notes/${noteId}/files`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      ...(workspaceId ? { "X-Workspace-Id": workspaceId } : {}),
    },
    body: formData,
  });
}

export function deleteFile(token, noteId, fileId, workspaceId = "") {
  return request(`/api/notes/${noteId}/files/${fileId}`, {
    method: "DELETE",
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export async function downloadFile(token, noteId, fileId, workspaceId = "") {
  const response = await requestRaw(`/api/notes/${noteId}/files/${fileId}`, {
    headers: authHeaders(token, "*/*", workspaceId),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    throw new Error(errorData?.message || "Не удалось скачать файл.");
  }

  return response.blob();
}
