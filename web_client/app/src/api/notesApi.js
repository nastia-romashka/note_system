import { authHeaders, request } from "./http";

export function fetchNotes(token, categoryId) {
  return request(`/api/notes?category_uuid=${encodeURIComponent(categoryId)}`, {
    headers: authHeaders(token),
  }).catch((error) => {
    if (String(error.message).toLowerCase().includes("not found")) {
      return [];
    }
    throw error;
  });
}

export function createNote(token, payload) {
  return request("/api/notes", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export function updateNote(token, noteId, payload) {
  return request(`/api/notes/${noteId}`, {
    method: "PATCH",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export function deleteNote(token, noteId) {
  return request(`/api/notes/${noteId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}
