import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Trophy, CircleDot, UtensilsCrossed, ArrowLeft, CheckCircle2, Clock } from 'lucide-react';
import { motion } from 'framer-motion';
import { getSpace, getSlots } from '../api/spaces';
import { createBooking } from '../api/bookings';
import { SPACE_IMAGES, SPACE_LABELS } from '../components/SpaceCard';
import { useAuth } from '../hooks/useAuth';
import { cn } from '@/lib/utils';
import type { SpaceType } from '../types';

const SPACE_ICONS: Record<SpaceType, typeof Trophy> = {
  cancha_padel: Trophy,
  cancha_futbol: CircleDot,
  quincho: UtensilsCrossed,
};

function getNextDays(n: number) {
  return Array.from({ length: n }, (_, i) => {
    const d = new Date();
    d.setDate(d.getDate() + i);
    return d;
  });
}

const DAYS = getNextDays(14);
const today = DAYS[0].toISOString().split('T')[0];

export default function SpaceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated } = useAuth();

  const [date, setDate] = useState(today);
  const [selectedSlot, setSelectedSlot] = useState<number | null>(null);
  const [booked, setBooked] = useState(false);
  const [imgIndex, setImgIndex] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const { data: space, isLoading: spaceLoading } = useQuery({
    queryKey: ['space', id],
    queryFn: () => getSpace(Number(id)),
    enabled: !!id,
  });

  const { data: slots, isLoading: slotsLoading } = useQuery({
    queryKey: ['slots', id, date],
    queryFn: () => getSlots(Number(id), date),
    enabled: !!id && !!date,
  });

  const bookingMutation = useMutation({
    mutationFn: () =>
      createBooking({ space_id: Number(id), slot_id: selectedSlot!, booking_date: date }),
    onSuccess: () => { setBooked(true); setSelectedSlot(null); },
  });

  const handleConfirm = () => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: location } });
      return;
    }
    bookingMutation.mutate();
  };

  const images = space ? SPACE_IMAGES[space.type] : [];

  useEffect(() => {
    if (images.length <= 1) return;
    intervalRef.current = setInterval(() => {
      setImgIndex((prev) => (prev + 1) % images.length);
    }, 3000);
    return () => { if (intervalRef.current) clearInterval(intervalRef.current); };
  }, [images.length]);

  const handleDotClick = (i: number) => {
    setImgIndex(i);
    if (intervalRef.current) clearInterval(intervalRef.current);
    intervalRef.current = setInterval(() => {
      setImgIndex((prev) => (prev + 1) % images.length);
    }, 3000);
  };

  if (spaceLoading) {
    return (
      <div className="animate-fade-up min-h-[calc(100vh-58px)] bg-bg">
        <div className="h-[55vh] bg-surface animate-pulse" />
        <div className="max-w-[1140px] mx-auto px-6 py-8 flex flex-col gap-4">
          <div className="h-8 w-1/3 bg-surface rounded-full animate-pulse" />
          <div className="h-4 w-2/3 bg-surface rounded-full animate-pulse" />
          <div className="h-24 bg-surface rounded-xl animate-pulse mt-4" />
        </div>
      </div>
    );
  }

  if (!space) return null;

  const Icon = SPACE_ICONS[space.type];

  return (
    <div className="animate-fade-up min-h-[calc(100vh-58px)] bg-bg">

      {/* ── Hero image gallery ── */}
      <div className="relative h-[55vh] bg-dark overflow-hidden">
        {images.map((src, i) => (
          <img
            key={src}
            src={src}
            alt={space.name}
            className={cn('absolute inset-0 w-full h-full object-cover transition-opacity duration-700', i === imgIndex ? 'opacity-100' : 'opacity-0')}
          />
        ))}
        <div className="absolute inset-0 bg-gradient-to-t from-black/65 via-black/10 to-transparent" />

        <button
          onClick={() => navigate('/canchas')}
          className="absolute top-5 left-5 flex items-center gap-2 text-white/80 text-sm font-semibold hover:text-white transition-colors bg-black/25 backdrop-blur-sm rounded-lg px-3 py-2 active:scale-[0.98]"
        >
          <ArrowLeft size={15} /> Canchas
        </button>

        {images.length > 1 && (
          <div className="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-2">
            {images.map((_, i) => (
              <button
                key={i}
                onClick={() => handleDotClick(i)}
                className={cn(
                  'rounded-full transition-all duration-300',
                  i === imgIndex ? 'w-5 h-2 bg-white' : 'w-2 h-2 bg-white/50 hover:bg-white/80',
                )}
              />
            ))}
          </div>
        )}

        <div className="absolute bottom-0 left-0 right-0 px-6 pb-8">
          <div className="max-w-[1140px] mx-auto">
            <span className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 bg-primary text-white text-2xs font-bold uppercase tracking-wide mb-3">
              <Icon size={11} />
              {SPACE_LABELS[space.type]}
            </span>
            <h1 className="font-serif text-4xl text-white tracking-tight">{space.name}</h1>
          </div>
        </div>
      </div>

      {/* ── Content ── */}
      <div className="max-w-[1140px] mx-auto px-6 py-8">

        {/* Info bar */}
        <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-3 mb-8">
          {space.description && (
            <p className="text-sm text-ink-2 leading-relaxed max-w-xl">{space.description}</p>
          )}
          <div className="flex items-baseline gap-1 shrink-0">
            <span className="font-serif text-3xl text-ink">
              ${space.price_per_slot.toLocaleString('es-AR')}
            </span>
            <span className="text-sm text-ink-2">/ turno</span>
          </div>
        </div>

        <div className="border-t border-black/[0.06] pt-8">

          {booked ? (
            <motion.div
              initial={{ opacity: 0, scale: 0.96 }}
              animate={{ opacity: 1, scale: 1 }}
              className="max-w-sm mx-auto bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-md p-8 text-center"
            >
              <div className="w-14 h-14 rounded-full bg-status-confirmed/10 flex items-center justify-center mx-auto mb-4">
                <CheckCircle2 size={28} className="text-status-confirmed" />
              </div>
              <h3 className="font-serif text-2xl text-ink mb-1">¡Reserva confirmada!</h3>
              <p className="text-sm text-ink-2 mb-6">Tu turno quedó registrado. Te esperamos.</p>
              <div className="flex flex-col gap-2">
                <button
                  onClick={() => navigate('/mis-reservas')}
                  className="bg-primary text-white font-bold rounded-lg px-6 py-3 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98]"
                >
                  Ver mis reservas
                </button>
                <button
                  onClick={() => { setBooked(false); setDate(today); }}
                  className="text-sm text-ink-2 hover:text-ink transition-colors py-2"
                >
                  Reservar otro turno
                </button>
              </div>
            </motion.div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-[1fr_auto] gap-10 items-start">

              {/* Date + slots */}
              <div>
                {/* Week strip */}
                <label className="text-2xs font-bold tracking-[0.08em] uppercase text-ink-2 mb-3 block">
                  Seleccioná el día
                </label>
                <div className="flex gap-2 overflow-x-auto pb-2 -mx-1 px-1">
                  {DAYS.map((d) => {
                    const iso = d.toISOString().split('T')[0];
                    const isSelected = date === iso;
                    const dayName = d.toLocaleDateString('es-AR', { weekday: 'short' });
                    const dayNum = d.getDate();
                    const monthName = d.toLocaleDateString('es-AR', { month: 'short' });
                    return (
                      <button
                        key={iso}
                        onClick={() => { setDate(iso); setSelectedSlot(null); }}
                        className={cn(
                          'flex flex-col items-center shrink-0 w-[52px] pt-2.5 pb-2 rounded-xl border-[1.5px] transition-all duration-normal active:scale-[0.96]',
                          isSelected
                            ? 'bg-primary border-primary text-white shadow-sm'
                            : 'bg-white border-black/[0.07] text-ink-2 hover:border-primary/40 hover:text-ink',
                        )}
                      >
                        <span className="text-2xs font-bold uppercase tracking-wide capitalize">
                          {dayName.replace('.', '')}
                        </span>
                        <span className="text-xl font-bold leading-none my-1">{dayNum}</span>
                        <span className={cn('text-2xs capitalize', isSelected ? 'text-white/70' : 'text-ink-2/60')}>
                          {monthName.replace('.', '')}
                        </span>
                      </button>
                    );
                  })}
                </div>

                {/* Slots */}
                <div className="mt-8">
                  <label className="text-2xs font-bold tracking-[0.08em] uppercase text-ink-2 mb-3 block">
                    Horarios disponibles
                  </label>

                  {slotsLoading ? (
                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                      {Array.from({ length: 6 }).map((_, i) => (
                        <div key={i} className="h-16 bg-surface rounded-xl animate-pulse" />
                      ))}
                    </div>
                  ) : !slots?.length ? (
                    <div className="flex items-center gap-2 py-6 text-sm text-ink-2">
                      <Clock size={15} className="shrink-0" />
                      No hay turnos disponibles para este día.
                    </div>
                  ) : (
                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                      {slots.map((slot) => (
                        <button
                          key={slot.id}
                          onClick={() => setSelectedSlot(slot.id === selectedSlot ? null : slot.id)}
                          className={cn(
                            'flex items-center justify-center px-3 py-3.5 rounded-xl border-[1.5px] text-sm font-bold transition-all duration-normal active:scale-[0.98]',
                            selectedSlot === slot.id
                              ? 'bg-primary/[0.08] border-primary text-primary'
                              : 'bg-white border-black/[0.07] text-ink hover:border-primary/40',
                          )}
                        >
                          {slot.label}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Confirm panel — sticky */}
              <div className="lg:sticky lg:top-[74px] w-full lg:w-[280px]">
                <div className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-md p-5">
                  <div className="flex flex-col gap-1 mb-5 pb-5 border-b border-black/[0.06]">
                    <span className="text-2xs font-bold uppercase tracking-wide text-ink-2">Resumen</span>
                    <span className="text-sm font-semibold text-ink">{space.name}</span>
                    {selectedSlot && slots && (
                      <span className="text-sm text-primary font-semibold">
                        Turno: {slots.find((s) => s.id === selectedSlot)?.label}
                      </span>
                    )}
                    <span className="text-sm text-ink-2">
                      {new Date(date + 'T12:00:00').toLocaleDateString('es-AR', {
                        weekday: 'long', day: 'numeric', month: 'long',
                      })}
                    </span>
                  </div>

                  <div className="flex items-baseline justify-between mb-5">
                    <span className="text-sm text-ink-2">Total</span>
                    <span className="font-serif text-2xl text-ink">
                      ${space.price_per_slot.toLocaleString('es-AR')}
                    </span>
                  </div>

                  {bookingMutation.error && (
                    <p className="text-xs text-status-cancelled font-medium mb-3">
                      {bookingMutation.error.message}
                    </p>
                  )}

                  <button
                    onClick={handleConfirm}
                    disabled={!selectedSlot || bookingMutation.isPending}
                    className="w-full bg-primary text-white font-bold rounded-lg px-5 py-3 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {bookingMutation.isPending
                      ? 'Confirmando...'
                      : !isAuthenticated
                      ? 'Iniciar sesión para reservar'
                      : 'Confirmar reserva'}
                  </button>

                  {!isAuthenticated && (
                    <p className="text-xs text-center text-ink-2 mt-2.5">
                      Seleccioná un turno y te redirigimos al login.
                    </p>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
