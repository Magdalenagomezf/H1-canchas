export type SpaceType = 'cancha_padel' | 'cancha_futbol' | 'cancha_padbol' | 'cancha_beach_voley' | 'quincho';

export interface Space {
  id: number;
  name: string;
  type: SpaceType;
  description: string | null;
  price_per_slot: number;
  created_at: string;
}

export interface SpaceSlot {
  id: number;
  label: string;
  description: string | null;
  start_time: string | null;
  end_time: string | null;
  available?: boolean;
}

export type BookingStatus = 'pending' | 'confirmed' | 'cancelled' | 'completed';

export interface BookingSpaceInfo {
  id: number;
  name: string;
  type: SpaceType;
}

export interface BookingSlotInfo {
  id: number;
  label: string;
  start_time: string | null;
  end_time: string | null;
}

export interface BookingCustomerInfo {
  user_id: number | null;
  name: string;
  phone: string;
  is_manual: boolean;
}

export type PaymentStatus = 'unpaid' | 'paid';

export interface BookingDetail {
  id: number;
  booking_date: string;
  status: BookingStatus;
  total_price: number;
  created_at: string;
  updated_at: string;

  deposit_amount: number;
  deposit_status: PaymentStatus;
  deposit_method?: string;
  expires_at?: string;
  balance_amount: number;
  balance_status: PaymentStatus;
  balance_method?: string;

  space: BookingSpaceInfo;
  slot: BookingSlotInfo;
  customer: BookingCustomerInfo;
  batch?: BatchInfo;
}

export type BatchType = 'recurring_teacher' | 'maintenance';
export type BatchStatus = 'active' | 'cancelled';

export interface BatchInfo {
  id: number;
  type: BatchType;
  reason: string;
}

export interface BookingBatch {
  id: number;
  type: BatchType;
  weekday?: number; // 0=domingo .. 6=sábado, solo recurring_teacher
  start_date: string;
  end_date: string;
  reason: string;
  status: BatchStatus;
  created_at: string;
  space: { id: number; name: string; type?: SpaceType | '' };
  slot?: { id: number; label?: string };
}

export interface CreateBatchResult {
  batch: BookingBatch;
  created_dates: string[];
  skipped_dates: string[];
}

export type PaymentKind = 'deposit' | 'balance';

export interface PaymentPreference {
  preference_id: string;
  init_point: string;
  sandbox_init_point: string;
}

export interface User {
  id: number;
  name: string;
  phone: string;
  email?: string | null;
  role: 'customer' | 'receptionist' | 'admin';
}

export interface AuthResponse {
  token: string;
  user: User;
}
