import apiClient from './client';
import type { BookingDetail } from '../types';

export interface CreateBookingPayload {
  space_id: number;
  slot_id: number;
  booking_date: string; // formato: "2026-06-15"
}

export const createBooking = (payload: CreateBookingPayload): Promise<BookingDetail> =>
  apiClient.post<BookingDetail>('/bookings', payload).then((r) => r.data);

export const getMyBookings = (): Promise<BookingDetail[]> =>
  apiClient.get<BookingDetail[]>('/bookings').then((r) => r.data);

export const cancelBooking = (id: number): Promise<{ message: string }> =>
  apiClient.patch<{ message: string }>(`/bookings/${id}/cancel`).then((r) => r.data);
