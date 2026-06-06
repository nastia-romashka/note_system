import {
  clearStoredSession,
  getStoredRefreshToken,
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

export function authHeaders(token, accept = "application/json") {
  return {
    Accept: accept,
    Authorization: `Bearer ${token}`,
  };
}

export async function request(path, options = {}) {
  const response = await requestRaw(path, options);
  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new APIRequestError(
      errorData?.developer_message || errorData?.message || "Р—Р°РїСЂРѕСЃ Р·Р°РІРµСЂС€РёР»СЃСЏ СЃ РѕС€РёР±РєРѕР№.",
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
  return fetch(`${API_BASE_URL}${path}`, requestOptions);
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

  return getStoredRefreshToken() !== "";
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

  const refreshToken = getStoredRefreshToken();
  if (!refreshToken) {
    throw new Error("missing refresh token");
  }

  refreshPromise = (async () => {
    const response = await fetch(`${API_BASE_URL}/api/auth`, {
      method: "PUT",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      throw new Error("refresh failed");
    }

    const data = await response.json();
    persistStoredSession(data.token, data.refresh_token, shouldRememberSession());
    return data.token;
  })();

  try {
    return await refreshPromise;
  } finally {
    refreshPromise = null;
  }
}
