import { authHeaders, request } from "./http";

export function fetchTags(token) {
  return request("/api/tags", { headers: authHeaders(token) }).catch((error) => {
    if (String(error.message).toLowerCase().includes("not found")) {
      return [];
    }
    throw error;
  });
}

export function createTag(token, name) {
  return request("/api/tags", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ name }),
  });
}
