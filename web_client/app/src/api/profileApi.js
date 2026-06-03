import { authHeaders, request } from "./http";

export function fetchProfileSummary(token) {
  return request("/api/me/summary", {
    headers: authHeaders(token),
  });
}

export function updateProfile(token, payload) {
  return request("/api/me", {
    method: "PATCH",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export function fetchProfileActions(token, limit = 50, offset = 0) {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });

  return request(`/api/me/actions?${params.toString()}`, {
    headers: authHeaders(token),
  });
}
