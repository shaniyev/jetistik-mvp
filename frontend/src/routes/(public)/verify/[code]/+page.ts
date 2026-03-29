import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params, fetch }) => {
  const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

  try {
    const res = await fetch(`${API_BASE}/api/v1/verify/${params.code}`);
    if (!res.ok) {
      return { result: null, code: params.code };
    }
    const body = await res.json();
    return { result: body.data, code: params.code };
  } catch {
    return { result: null, code: params.code };
  }
};
