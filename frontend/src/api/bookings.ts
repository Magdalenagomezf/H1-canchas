import apiClient from './client';
import type { BookingDetail } from '../types';

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

export const getAllBookings = (date?: string): Promise<BookingDetail[]> => {
  const params = date ? { date } : {};
  return apiClient.get<BookingDetail[]>('/admin/bookings', { params }).then((r) => r.data);
};

export const cancelBooking = (id: number): Promise<{ message: string }> =>
  apiClient.patch<{ message: string }>(`/bookings/${id}/cancel`).then((r) => r.data);
