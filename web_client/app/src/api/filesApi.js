import { authHeaders, request, requestRaw } from "./http";

export function fetchFiles(token, noteId) {
  return request(`/api/notes/${noteId}/files`, { headers: authHeaders(token) });
}

export function uploadFile(token, noteId, file) {
  const formData = new FormData();
  formData.append("file", file);

  return request(`/api/notes/${noteId}/files`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: formData,
  });
}

export function deleteFile(token, noteId, fileId) {
  return request(`/api/notes/${noteId}/files/${fileId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export async function downloadFile(token, noteId, fileId) {
  const response = await requestRaw(`/api/notes/${noteId}/files/${fileId}`, {
    headers: authHeaders(token, "*/*"),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    throw new Error(errorData?.message || "РќРµ СѓРґР°Р»РѕСЃСЊ СЃРєР°С‡Р°С‚СЊ С„Р°Р№Р».");
  }

  return response.blob();
}
