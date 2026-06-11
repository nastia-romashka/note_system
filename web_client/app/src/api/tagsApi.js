import { authHeaders, request } from "./http";

export function fetchTags(token, workspaceId = "") {
  return request("/api/tags", { headers: authHeaders(token, "application/json", workspaceId) }).catch((error) => {
    if (String(error.message).toLowerCase().includes("not found")) {
      return [];
    }
    throw error;
  });
}

export function createTag(token, name, workspaceId = "") {
  return request("/api/tags", {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ name }),
  });
}
