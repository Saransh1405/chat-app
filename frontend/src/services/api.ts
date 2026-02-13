// Base API client with JWT interceptor
// TODO: Replace BASE_URL with your actual backend URL

const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";

class ApiClient {
  private getToken(): string | null {
    return localStorage.getItem("access_token");
  }

  private getHeaders(contentType = "application/json"): HeadersInit {
    const headers: HeadersInit = {};
    if (contentType) headers["Content-Type"] = contentType;

    const token = this.getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;

    return headers;
  }

  async get<T>(endpoint: string): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      headers: this.getHeaders(),
    });
    if (!res.ok) throw new Error(`GET ${endpoint} failed: ${res.statusText}`);
    return res.json();
  }

  async post<T>(endpoint: string, body?: unknown): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      method: "POST",
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!res.ok) throw new Error(`POST ${endpoint} failed: ${res.statusText}`);
    return res.json();
  }

  async put<T>(endpoint: string, body?: unknown): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      method: "PUT",
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!res.ok) throw new Error(`PUT ${endpoint} failed: ${res.statusText}`);
    return res.json();
  }

  async patch<T>(endpoint: string, body?: unknown): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      method: "PATCH",
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    });
    if (!res.ok) throw new Error(`PATCH ${endpoint} failed: ${res.statusText}`);
    return res.json();
  }

  async delete<T>(endpoint: string): Promise<T> {
    const res = await fetch(`${BASE_URL}${endpoint}`, {
      method: "DELETE",
      headers: this.getHeaders(),
    });
    if (!res.ok) throw new Error(`DELETE ${endpoint} failed: ${res.statusText}`);
    return res.json();
  }

  async upload<T>(endpoint: string, file: File): Promise<T> {
    const formData = new FormData();
    formData.append("file", file);

    const res = await fetch(`${BASE_URL}${endpoint}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${this.getToken()}` },
      body: formData,
    });
    if (!res.ok) throw new Error(`UPLOAD ${endpoint} failed: ${res.statusText}`);
    return res.json();
  }
}

export const api = new ApiClient();
