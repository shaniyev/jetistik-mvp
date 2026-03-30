const API_BASE = import.meta.env.VITE_API_URL || "";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: unknown
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface ApiResponse<T> {
  data: T;
}

interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
  };
}

let accessToken: string | null = null;
let refreshPromise: Promise<boolean> | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

async function refreshAccessToken(): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
      accessToken = null;
      return false;
    }

    const body = (await res.json()) as ApiResponse<{ access_token: string }>;
    accessToken = body.data.access_token;
    return true;
  } catch {
    accessToken = null;
    return false;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  retry = true
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (res.status === 401 && retry && !path.includes("/auth/")) {
    // Deduplicate concurrent refresh calls
    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
    }

    const refreshed = await refreshPromise;
    if (refreshed) {
      return request<T>(path, options, false);
    }
  }

  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as ApiErrorResponse | null;
    throw new ApiError(
      res.status,
      body?.error?.code ?? "UNKNOWN",
      body?.error?.message ?? res.statusText,
      body?.error?.details
    );
  }

  if (res.status === 204) return undefined as T;

  return res.json() as Promise<T>;
}

export const api = {
  get<T>(path: string): Promise<ApiResponse<T>> {
    return request(path);
  },

  post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
    return request(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    });
  },

  patch<T>(path: string, body: unknown): Promise<ApiResponse<T>> {
    return request(path, {
      method: "PATCH",
      body: JSON.stringify(body),
    });
  },

  delete<T>(path: string): Promise<ApiResponse<T>> {
    return request(path, { method: "DELETE" });
  },

  upload<T>(path: string, formData: FormData): Promise<ApiResponse<T>> {
    return request(path, {
      method: "POST",
      body: formData,
      headers: {},
    });
  },
};

export type { ApiResponse, PaginatedResponse };
