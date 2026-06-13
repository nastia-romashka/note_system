import { request } from "./http";
import { shouldRememberSession } from "./session";

export function login(username, password, remember = true) {
  return request("/api/auth", {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password, remember }),
    skipAuthRefresh: true,
  });
}

export function signup(payload) {
  return request("/api/signup", {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
    skipAuthRefresh: true,
  });
}

export function refreshSession() {
  return request("/api/auth", {
    method: "PUT",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ remember: shouldRememberSession() }),
    skipAuthRefresh: true,
  });
}

export async function logoutCurrentSession() {
  return request("/api/auth", {
    method: "DELETE",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({}),
    skipAuthRefresh: true,
  });
}
