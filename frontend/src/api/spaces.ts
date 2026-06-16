import apiClient from './client';
import type { Space, SpaceSlot } from '../types';

export const getSpaces = (): Promise<Space[]> =>
  apiClient.get<Space[]>('/spaces').then((r) => r.data);

export const getSpace = (id: number): Promise<Space> =>
  apiClient.get<Space>(`/spaces/${id}`).then((r) => r.data);

export const getSlots = (spaceId: number, date?: string): Promise<SpaceSlot[]> => {
  const params = date ? { date } : {};
  return apiClient.get<SpaceSlot[]>(`/spaces/${spaceId}/slots`, { params }).then((r) => r.data);
};
