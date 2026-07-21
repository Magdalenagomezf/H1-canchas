import apiClient from './client';
import type { BookingDetail, BookingBatch, BatchType, CreateBatchResult } from '../types';

export interface CreateBookingPayload {
  space_id: number;
  slot_id: number;
  booking_date: string;
}

export interface CreateManualBookingPayload {
  space_id: number;
  slot_id: number;
  booking_date: string;
  customer_name: string;
  customer_phone: string;
}

export const createBooking = (payload: CreateBookingPayload): Promise<BookingDetail> =>
  apiClient.post<BookingDetail>('/bookings', payload).then((r) => r.data);

export const createManualBooking = (payload: CreateManualBookingPayload): Promise<BookingDetail> =>
  apiClient.post<BookingDetail>('/bookings/manual', payload).then((r) => r.data);

export const getMyBookings = (): Promise<BookingDetail[]> =>
  apiClient.get<BookingDetail[]>('/bookings').then((r) => r.data);

export const getBooking = (id: number): Promise<BookingDetail> =>
  apiClient.get<BookingDetail>(`/bookings/${id}`).then((r) => r.data);

export const getAllBookings = (date?: string): Promise<BookingDetail[]> => {
  const params = date ? { date } : {};
  return apiClient.get<BookingDetail[]>('/admin/bookings', { params }).then((r) => r.data);
};

export const cancelBooking = (id: number): Promise<{ message: string }> =>
  apiClient.patch<{ message: string }>(`/bookings/${id}/cancel`).then((r) => r.data);

// ── Turnos fijos (profesores) y bloqueos por mantenimiento ────────────────────

export interface CreateRecurringBookingPayload {
  space_id: number;
  slot_id: number;
  weekday: number; // 0=domingo .. 6=sábado
  customer_name: string;
  customer_phone: string;
  start_date: string;
  end_date: string;
  note?: string;
}

export interface CreateMaintenanceBlockPayload {
  space_id: number;
  slot_id?: number | null; // null/omitido = todos los turnos activos del espacio
  start_date: string;
  end_date: string;
  reason: string;
}

export const createRecurringBooking = (payload: CreateRecurringBookingPayload): Promise<CreateBatchResult> =>
  apiClient.post<CreateBatchResult>('/bookings/recurring', payload).then((r) => r.data);

export const createMaintenanceBlock = (payload: CreateMaintenanceBlockPayload): Promise<CreateBatchResult> =>
  apiClient.post<CreateBatchResult>('/bookings/block', payload).then((r) => r.data);

export const getBatches = (type?: BatchType): Promise<BookingBatch[]> => {
  const params = type ? { type } : {};
  return apiClient.get<BookingBatch[]>('/bookings/batches', { params }).then((r) => r.data);
};

export const cancelBatch = (id: number): Promise<{ message: string }> =>
  apiClient.patch<{ message: string }>(`/bookings/batches/${id}/cancel`).then((r) => r.data);
