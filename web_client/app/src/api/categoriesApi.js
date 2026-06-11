import { authHeaders, request } from "./http";

export function fetchCategories(token, workspaceId = "") {
  return request("/api/categories", { headers: authHeaders(token, "application/json", workspaceId) }).catch((error) => {
    if (String(error.message).toLowerCase().includes("not found")) {
      return [];
    }
    throw error;
  });
}

export function createCategory(token, payload, workspaceId = "") {
  return request("/api/categories", {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function updateCategory(token, categoryId, payload, workspaceId = "") {
  return request(`/api/categories/${categoryId}`, {
    method: "PATCH",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function deleteCategory(token, categoryId, workspaceId = "") {
  return request(`/api/categories/${categoryId}`, {
    method: "DELETE",
    headers: authHeaders(token, "application/json", workspaceId),
  });
}
