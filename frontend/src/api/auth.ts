import apiClient from './client';
import type { AuthResponse } from '../types';

export const login = (phone: string, password: string): Promise<AuthResponse> =>
  apiClient.post<AuthResponse>('/auth/login', { phone, password }).then((r) => r.data);

export const register = (name: string, phone: string, password: string, email: string): Promise<AuthResponse> =>
  apiClient.post<AuthResponse>('/auth/register', { name, phone, password, email }).then((r) => r.data);

/** Sends the Google ID token (`credential` from <GoogleLogin />) to the backend. */
export const loginWithGoogle = (credential: string): Promise<AuthResponse> =>
  apiClient.post<AuthResponse>('/auth/google', { credential }).then((r) => r.data);

/** Asks for a reset link. The backend answers the same message whether the email has an account or not. */
export const requestPasswordReset = (email: string): Promise<string> =>
  apiClient.post<{ message: string }>('/auth/password-reset', { email }).then((r) => r.data.message);

/** Sets a new password with the token from the email link. Fails with code "invalid_reset_token" when the link no longer works. */
export const confirmPasswordReset = (token: string, password: string): Promise<void> =>
  apiClient.post('/auth/password-reset/confirm', { token, password }).then(() => undefined);
