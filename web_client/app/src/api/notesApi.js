import { API_BASE_URL, authHeaders, readSafeJson, request } from "./http";

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

export function fetchSearchNotes(token, { query, categoryId }) {
  const params = new URLSearchParams();
  if (query) {
    params.set("q", query);
  }
  if (categoryId) {
    params.set("category_uuid", categoryId);
  }

  return request(`/api/search/notes?${params.toString()}`, {
    headers: authHeaders(token),
  });
}

export function fetchCalendarNotes(token, { from, to }) {
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  });

  return request(`/api/calendar?${params.toString()}`, {
    headers: authHeaders(token),
  });
}

export function createNote(token, payload) {
  return fetch(`${API_BASE_URL}/api/notes`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  }).then(async (response) => {
    if (!response.ok) {
      const errorData = await readSafeJson(response);
      throw new Error(errorData?.developer_message || errorData?.message || "Запрос завершился с ошибкой.");
    }

    const location = response.headers.get("Location") || "";
    const uuid = location.split("/").filter(Boolean).at(-1) || "";
    return {
      uuid,
      category_uuid: payload.category_uuid,
      ...payload,
    };
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

export function duplicateNote(token, noteId, payload) {
  return request(`/api/notes/${noteId}/duplicate`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}
