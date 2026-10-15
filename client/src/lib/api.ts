import { getClientId } from "./client-id";
import { getApiKey } from "./api-key";

export class ApiError extends Error {
  status: number;
  constructor(path: string, status: number, detail = "") {
    super(detail ? `${path} ${status} ${detail}` : `${path} ${status}`);
    this.status = status;
  }
}

/** Absolute API origin in dev (`VITE_API_URL`). Empty keeps same-origin `/api` paths. */
export const apiUrl = (path: string): string => {
  const base = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");
  return `${base}${path}`;
};

export const baseHeaders = (): HeadersInit => {
  const headers: Record<string, string> = {
    "X-Client-Id": getClientId(),
    "Content-Type": "application/json",
  };
  const key = getApiKey();
  if (key) headers.Authorization = `Bearer ${key}`;
  return headers;
};

export const apiGet = async <T>(path: string): Promise<T> => {
  const r = await fetch(apiUrl(path), { headers: baseHeaders() });
  if (!r.ok) throw new ApiError(path, r.status);
  return r.json() as Promise<T>;
};

export const apiPost = async <T>(path: string, body: unknown): Promise<T> => {
  const r = await fetch(apiUrl(path), { method: "POST", headers: baseHeaders(), body: JSON.stringify(body) });
  if (!r.ok) {
    const text = await r.text().catch(() => "");
    throw new ApiError(path, r.status, text);
  }
  return r.json() as Promise<T>;
};

export const apiPostVoid = async (path: string): Promise<void> => {
  const r = await fetch(apiUrl(path), { method: "POST", headers: baseHeaders(), body: "{}" });
  if (!r.ok) throw new ApiError(path, r.status);
};

export const apiDelete = async (path: string): Promise<void> => {
  const r = await fetch(apiUrl(path), { method: "DELETE", headers: baseHeaders() });
  if (!r.ok) throw new ApiError(path, r.status);
};
