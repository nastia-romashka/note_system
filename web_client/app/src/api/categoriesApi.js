import { authHeaders, request } from "./http";

export function fetchCategories(token) {
  return request("/api/categories", { headers: authHeaders(token) }).catch((error) => {
    if (String(error.message).toLowerCase().includes("not found")) {
      return [];
    }
    throw error;
  });
}

export function createCategory(token, payload) {
  return request("/api/categories", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export function updateCategory(token, categoryId, payload) {
  return request(`/api/categories/${categoryId}`, {
    method: "PATCH",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export function deleteCategory(token, categoryId) {
  return request(`/api/categories/${categoryId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}
