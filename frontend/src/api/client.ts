import axios from 'axios';

/**
 * Error thrown by every API call. Keeps the HTTP status and the optional
 * machine-readable `code` from the backend (e.g. "phone_required"), so the UI
 * never has to branch on the message text.
 */
export class ApiError extends Error {
  readonly status?: number;
  readonly code?: string;

  constructor(message: string, status?: number, code?: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
  headers: { 'Content-Type': 'application/json' },
});

apiClient.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.error ?? error.message;
    return Promise.reject(new ApiError(message, error.response?.status, error.response?.data?.code));
  }
);

export default apiClient;
