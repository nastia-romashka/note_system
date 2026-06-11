import { authHeaders, request } from "./http";

export function fetchGraph(token, workspaceId = "") {
  return request("/api/graph", {
    headers: authHeaders(token, "application/json", workspaceId),
  });
}

export function createGraphLink(token, payload, workspaceId = "") {
  return request("/api/graph/links", {
    method: "POST",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function deleteGraphLink(token, payload, workspaceId = "") {
  return request("/api/graph/links", {
    method: "DELETE",
    headers: {
      ...authHeaders(token, "application/json", workspaceId),
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}
