import apiClient from './client';
import type { AuthResponse } from '../types';

export const login = (phone: string, password: string): Promise<AuthResponse> =>
  apiClient.post<AuthResponse>('/auth/login', { phone, password }).then((r) => r.data);

export const register = (name: string, phone: string, password: string): Promise<AuthResponse> =>
  apiClient.post<AuthResponse>('/auth/register', { name, phone, password }).then((r) => r.data);
