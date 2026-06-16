export type SpaceType = 'cancha_padel' | 'cancha_futbol' | 'quincho';

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

export interface BookingDetail {
  id: number;
  booking_date: string;
  status: BookingStatus;
  total_price: number;
  created_at: string;
  updated_at: string;
  space: BookingSpaceInfo;
  slot: BookingSlotInfo;
  customer: BookingCustomerInfo;
}

export interface User {
  id: number;
  name: string;
  phone: string;
  role: 'customer' | 'receptionist' | 'admin';
}

export interface AuthResponse {
  token: string;
  user: User;
}
