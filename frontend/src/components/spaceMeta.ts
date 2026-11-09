import { Trophy, CircleDot, Goal, Volleyball, UtensilsCrossed, type LucideIcon } from 'lucide-react';
import type { SpaceType } from '../types';

// Replace these arrays with your own uploaded photos — one array per space type
export const SPACE_IMAGES: Record<SpaceType, string[]> = {
  cancha_padel: [
    '/spaces/padel.jpeg',
    '/spaces/padel2.jpeg',
    '/spaces/padel3.jpeg',
  ],
  cancha_futbol: [
    '/spaces/futbol1.jpeg',
    '/spaces/futbol2.jpeg',
    '/spaces/fubtol3.jpeg',
  ],
  // TODO: add more photos to public/spaces/ (one per type currently)
  cancha_padbol: ['/spaces/padbol.jpeg'],
  cancha_beach_voley: ['/spaces/beach.jpeg'],
  quincho: [
    '/spaces/quincho1.jpeg',
    '/spaces/quincho2.jpeg',
  ],
};

export const SPACE_ICONS: Record<SpaceType, LucideIcon> = {
  cancha_padel: Trophy,
  cancha_futbol: CircleDot,
  cancha_padbol: Goal,
  cancha_beach_voley: Volleyball,
  quincho: UtensilsCrossed,
};

export const SPACE_LABELS: Record<SpaceType, string> = {
  cancha_padel: 'Pádel',
  cancha_futbol: 'Fútbol',
  cancha_padbol: 'Padbol',
  cancha_beach_voley: 'Beach vóley',
  quincho: 'Quincho / Salón',
};
