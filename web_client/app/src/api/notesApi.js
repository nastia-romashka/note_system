import { API_BASE_URL, authHeaders, readSafeJson, request } from "./http";

export function fetchNotes(token, categoryId, workspaceId = "") {
  return request(`/api/notes?category_uuid=${encodeURIComponent(categoryId)}`, {
    headers: authHeaders(token, "application/json", workspaceId),
  }).catch((error) => {
    if (String(error.message).toLowerCase().includes("not found")) {
      return [];
    }
    throw error;
  });
}

export function fetchSearchNotes(token, { query, categoryId, workspaceId = "" }) {
  const params = new URLSearchParams();
  if (query) {
    params.set("q", query);
  }
  if (categoryId) {
    params.set("category_uuid", categoryId);
  }

  return request(`/api/search/notes?${params.toString()}`, {
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export function fetchCalendarNotes(token, { from, to, workspaceId = "" }) {
  const params = new URLSearchParams({
    from: String(from),
    to: String(to),
  });

  return request(`/api/calendar?${params.toString()}`, {
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export function createNote(token, payload, workspaceId = "") {
  return fetch(`${API_BASE_URL}/api/notes`, {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
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

export function updateNote(token, noteId, payload, workspaceId = "") {
  return request(`/api/notes/${noteId}`, {
    method: "PATCH",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function deleteNote(token, noteId, workspaceId = "") {
  return request(`/api/notes/${noteId}`, {
    method: "DELETE",
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export function duplicateNote(token, noteId, payload, workspaceId = "") {
  return request(`/api/notes/${noteId}/duplicate`, {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}
