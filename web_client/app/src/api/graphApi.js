import { authHeaders, request } from "./http";

export function fetchGraph(token) {
  return request("/api/graph", {
    headers: authHeaders(token),
  });
}
