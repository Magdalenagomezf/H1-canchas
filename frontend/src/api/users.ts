import apiClient from './client';
import type { User } from '../types';

export interface AdminUser {
  id: number;
  name: string;
  phone: string | null;
  role: 'customer' | 'receptionist' | 'admin';
}

export interface CreateStaffPayload {
  name: string;
  phone: string;
  password: string;
  role: 'receptionist' | 'admin';
}

export const getUsers = (): Promise<AdminUser[]> =>
  apiClient.get<AdminUser[]>('/admin/users').then((r) => r.data);

export const createStaffUser = (payload: CreateStaffPayload): Promise<{ id: number }> =>
  apiClient.post<{ id: number }>('/admin/users', payload).then((r) => r.data);

export const updateUserRole = (id: number, role: string): Promise<void> =>
  apiClient.patch(`/admin/users/${id}/role`, { role }).then(() => undefined);

export const deleteUser = (id: number): Promise<void> =>
  apiClient.delete(`/admin/users/${id}`).then(() => undefined);

/** Fields accepted by PATCH /me. An absent field is left unchanged. */
export interface UpdateMePayload {
  name?: string;
  email?: string;
  phone?: string;
}

export const getMe = (): Promise<User> =>
  apiClient.get<User>('/me').then((r) => r.data);

export const updateMe = (payload: UpdateMePayload): Promise<User> =>
  apiClient.patch<User>('/me', payload).then((r) => r.data);
