import apiClient from './client';

export interface AdminUser {
  id: number;
  name: string;
  phone: string;
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
