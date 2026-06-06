import { request } from "./http";
import { getStoredRefreshToken } from "./session";

export function login(username, password) {
  return request("/api/auth", {
    method: "POST",
    body: JSON.stringify({ username, password }),
    skipAuthRefresh: true,
  });
}

export function signup(payload) {
  return request("/api/signup", {
    method: "POST",
    body: JSON.stringify(payload),
    skipAuthRefresh: true,
  });
}

export function refreshSession() {
  return request("/api/auth", {
    method: "PUT",
    body: JSON.stringify({ refresh_token: getStoredRefreshToken() }),
    skipAuthRefresh: true,
  });
}

export async function logoutCurrentSession() {
  const refreshToken = getStoredRefreshToken();
  if (!refreshToken) {
    return null;
  }

  return request("/api/auth", {
    method: "DELETE",
    body: JSON.stringify({ refresh_token: refreshToken }),
    skipAuthRefresh: true,
  });
}
