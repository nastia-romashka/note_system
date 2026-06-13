import {
  clearStoredSession,
  persistStoredSession,
  shouldRememberSession,
} from "./session";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

let refreshPromise = null;

class APIRequestError extends Error {
  constructor(message, status) {
    super(message);
    this.name = "APIRequestError";
    this.status = status;
  }
}

export function authHeaders(token, accept = "application/json", workspaceId = "") {
  const headers = {
    Accept: accept,
    Authorization: `Bearer ${token}`,
  };

  if (workspaceId) {
    headers["X-Workspace-Id"] = workspaceId;
  }

  return headers;
}

export async function request(path, options = {}) {
  const response = await requestRaw(path, options);
  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new APIRequestError(
      errorData?.developer_message || errorData?.message || "Запрос завершился с ошибкой.",
      response.status,
    );
  }
  if (response.status === 204) {
    return null;
  }

  const contentType = response.headers.get("Content-Type") || "";
  const responseText = await response.text();
  if (!responseText.trim()) {
    return null;
  }

  if (contentType.includes("application/json")) {
    return JSON.parse(responseText);
  }

  return responseText;
}

export async function requestRaw(path, options = {}) {
  return fetchWithAutoRefresh(path, options);
}

export async function readSafeJson(response) {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

async function fetchWithAutoRefresh(path, options = {}) {
  const initialResponse = await performFetch(path, options);
  if (!shouldRefreshRequest(initialResponse, options)) {
    return initialResponse;
  }

  try {
    const nextToken = await refreshAccessToken();
    return performFetch(path, withAuthorization(options, nextToken));
  } catch {
    clearStoredSession();
    return initialResponse;
  }
}

async function performFetch(path, options = {}) {
  const { skipAuthRefresh: _skipAuthRefresh, ...requestOptions } = options;
  return fetch(`${API_BASE_URL}${path}`, {
    credentials: "include",
    ...requestOptions,
  });
}

function shouldRefreshRequest(response, options = {}) {
  if (response.status !== 401 || options.skipAuthRefresh) {
    return false;
  }

  const headers = new Headers(options.headers || {});
  const authorization = headers.get("Authorization") || "";
  if (!authorization.startsWith("Bearer ")) {
    return false;
  }

  return true;
}

function withAuthorization(options, token) {
  const headers = new Headers(options.headers || {});
  headers.set("Authorization", `Bearer ${token}`);

  return {
    ...options,
    headers,
    skipAuthRefresh: true,
  };
}

async function refreshAccessToken() {
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    const response = await fetch(`${API_BASE_URL}/api/auth`, {
      method: "PUT",
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ remember: shouldRememberSession() }),
    });

    if (!response.ok) {
      throw new Error("refresh failed");
    }

    const data = await response.json();
    persistStoredSession(data.token, shouldRememberSession());
    return data.token;
  })();

  try {
    return await refreshPromise;
  } finally {
    refreshPromise = null;
  }
}
