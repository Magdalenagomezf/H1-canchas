import apiClient from './client';
import type { PaymentKind, PaymentPreference } from '../types';

export const generatePreference = (
  bookingId: number,
  kind: PaymentKind
): Promise<PaymentPreference> =>
  apiClient
    .post<PaymentPreference>(`/bookings/${bookingId}/payments`, { kind })
    .then((r) => r.data);
