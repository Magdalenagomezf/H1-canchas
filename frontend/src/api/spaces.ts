import apiClient from './client';
import type { Space, SpaceSlot, SpaceType } from '../types';

export const getSpaces = (): Promise<Space[]> =>
  apiClient.get<Space[]>('/spaces').then((r) => r.data);

export const getSpace = (id: number): Promise<Space> =>
  apiClient.get<Space>(`/spaces/${id}`).then((r) => r.data);

export const getSlots = (spaceId: number, date?: string): Promise<SpaceSlot[]> => {
  const params = date ? { date } : {};
  return apiClient.get<SpaceSlot[]>(`/spaces/${spaceId}/slots`, { params }).then((r) => r.data);
};

export interface CreateSpacePayload {
  name: string;
  type: SpaceType;
  description?: string;
  price_per_slot: number;
}

export const createSpace = (payload: CreateSpacePayload): Promise<Space> =>
  apiClient.post<Space>('/spaces', payload).then((r) => r.data);

export interface UpdateSpacePayload {
  name: string;
  description?: string;
  price_per_slot: number;
}

export const updateSpace = (id: number, payload: UpdateSpacePayload): Promise<void> =>
  apiClient.put(`/spaces/${id}`, payload).then(() => undefined);

export const deleteSpace = (id: number): Promise<void> =>
  apiClient.delete(`/admin/spaces/${id}`).then(() => undefined);
