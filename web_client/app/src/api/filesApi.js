import { API_BASE_URL, authHeaders, readSafeJson, request } from "./http";

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
  const response = await fetch(`${API_BASE_URL}/api/notes/${noteId}/files/${fileId}`, {
    headers: authHeaders(token, "*/*"),
  });

  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.message || "Не удалось скачать файл.");
  }

  return response.blob();
}
