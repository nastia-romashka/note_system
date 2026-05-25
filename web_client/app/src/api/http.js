export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export function authHeaders(token, accept = "application/json") {
  return {
    Accept: accept,
    Authorization: `Bearer ${token}`,
  };
}

export async function request(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, options);
  if (!response.ok) {
    const errorData = await readSafeJson(response);
    throw new Error(errorData?.developer_message || errorData?.message || "Запрос завершился с ошибкой.");
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

export async function readSafeJson(response) {
  try {
    return await response.json();
  } catch {
    return null;
  }
}
