import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowRight, Trophy, CircleDot, Goal, Volleyball, UtensilsCrossed, type LucideIcon } from 'lucide-react';
import { motion } from 'framer-motion';
import type { Space, SpaceType } from '../types';
import { cn } from '@/lib/utils';

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

export function SpaceCard({ space, index }: { space: Space; index: number }) {
  const navigate = useNavigate();
  const Icon = SPACE_ICONS[space.type];
  // Fallback to no images so a type without photos renders without crashing
  const images = SPACE_IMAGES[space.type] ?? [];

  const [imgIndex, setImgIndex] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const isTouch = typeof window !== 'undefined' && window.matchMedia('(hover: none)').matches;

  // On touch devices, auto-cycle images since there's no hover
  useEffect(() => {
    if (!isTouch || images.length <= 1) return;
    intervalRef.current = setInterval(() => {
      setImgIndex((prev) => (prev + 1) % images.length);
    }, 2000);
    return () => { if (intervalRef.current) clearInterval(intervalRef.current); };
  }, [isTouch, images.length]);

  const handleMouseEnter = () => {
    if (isTouch) return;
    intervalRef.current = setInterval(() => {
      setImgIndex((prev) => (prev + 1) % images.length);
    }, 900);
  };

  const handleMouseLeave = () => {
    if (isTouch) return;
    if (intervalRef.current) clearInterval(intervalRef.current);
    intervalRef.current = null;
    setImgIndex(0);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ delay: index * 0.08, duration: 0.45 }}
      className="group bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm overflow-hidden cursor-pointer transition-all duration-normal ease-smooth hover:-translate-y-[6px] hover:shadow-lg"
      onClick={() => navigate(`/canchas/${space.id}`)}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {/* Image area — bigger than before */}
      <div className="relative h-60 overflow-hidden">
        {images.map((src, i) => (
          <img
            key={src}
            src={src}
            alt={space.name}
            className={cn('absolute inset-0 w-full h-full object-cover transition-opacity duration-700', i === imgIndex ? 'opacity-100' : 'opacity-0')}
          />
        ))}

        {/* Gradient overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/35 to-transparent" />

        {/* Type badge */}
        <span className="absolute top-3 left-3 inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 bg-primary text-white text-2xs font-bold uppercase tracking-wide">
          <Icon size={11} />
          {SPACE_LABELS[space.type]}
        </span>

        {/* Dot indicators */}
        {images.length > 1 && (
          <div className="absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-1.5">
            {images.map((_, i) => (
              <span
                key={i}
                className={cn('block rounded-full transition-all duration-300', i === imgIndex ? 'w-4 h-1.5 bg-white' : 'w-1.5 h-1.5 bg-white/50')}
              />
            ))}
          </div>
        )}
      </div>

      <div className="p-5 flex flex-col gap-2">
        <h3 className="text-base font-bold text-ink">{space.name}</h3>
        {space.description && (
          <p className="text-sm text-ink-2 leading-relaxed line-clamp-2">{space.description}</p>
        )}
        <div className="flex items-center justify-between mt-2">
          <div>
            <span className="font-serif text-2xl text-ink">
              ${space.price_per_slot.toLocaleString('es-AR')}
            </span>
            <span className="text-xs text-ink-2 ml-1">/ turno</span>
          </div>
          <button
            onClick={(e) => { e.stopPropagation(); navigate(`/canchas/${space.id}`); }}
            className="bg-primary text-white font-bold rounded-lg px-4 py-2 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:scale-[1.02] hover:shadow-glow active:scale-[0.98] flex items-center gap-1.5"
          >
            Reservar <ArrowRight size={14} />
          </button>
        </div>
      </div>
    </motion.div>
  );
}

export function SpaceCardSkeleton() {
  return (
    <div className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm overflow-hidden">
      <div className="h-60 bg-surface animate-pulse" />
      <div className="p-5 flex flex-col gap-3">
        <div className="h-3 w-1/3 bg-surface-2 rounded-full animate-pulse" />
        <div className="h-5 w-2/3 bg-surface-2 rounded-full animate-pulse" />
        <div className="h-3 w-full bg-surface-2 rounded-full animate-pulse" />
        <div className="flex justify-between items-center mt-1">
          <div className="h-7 w-20 bg-surface-2 rounded-full animate-pulse" />
          <div className="h-9 w-24 bg-surface-2 rounded-lg animate-pulse" />
        </div>
      </div>
    </div>
  );
}
