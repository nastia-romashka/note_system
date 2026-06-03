import { authHeaders, request } from "./http";

export function fetchGraph(token) {
  return request("/api/graph", {
    headers: authHeaders(token),
  });
}

export function createGraphLink(token, payload) {
  return request("/api/graph/links", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export function deleteGraphLink(token, payload) {
  return request("/api/graph/links", {
    method: "DELETE",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}
