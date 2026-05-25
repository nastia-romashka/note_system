import { request } from "./http";

export function login(username, password) {
  return request("/api/auth", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export function signup(payload) {
  return request("/api/signup", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
