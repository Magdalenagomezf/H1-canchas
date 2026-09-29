import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import { getMyBookings, cancelBooking } from '../api/bookings';
import type { BookingDetail, BookingStatus } from '../types';
import { SPACE_LABELS } from '../components/SpaceCard';
import { ModalPortal } from '../components/ModalPortal';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/components/ui/button';
import { LandingFooter } from '@/components/landing/LandingFooter';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { formatPrice } from '@/components/landing/spaceHelpers';
import { CONTAINER, EASE, MONO } from '@/components/landing/ui';

const STATUS_CONFIG: Record<BookingStatus, { label: string; dot: string; text: string }> = {
  pending:   { label: 'Pendiente',  dot: 'bg-light',    text: 'text-light' },
  confirmed: { label: 'Confirmada', dot: 'bg-paper',    text: 'text-paper' },
  cancelled: { label: 'Cancelada',  dot: 'bg-concrete', text: 'text-concrete' },
  completed: { label: 'Completada', dot: 'bg-concrete', text: 'text-concrete' },
};

function formatDate(iso: string) {
  return new Date(iso + 'T12:00:00').toLocaleDateString('es-AR', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
  });
}

function BookingRow({
  booking,
  onCancel,
  muted = false,
}: {
  booking: BookingDetail;
  onCancel: (b: BookingDetail) => void;
  muted?: boolean;
}) {
  const navigate = useNavigate();
  const reduce = useReducedMotion();
  const status = STATUS_CONFIG[booking.status];
  const canCancel = booking.status === 'pending' || booking.status === 'confirmed';

  return (
    <motion.li
      layout
      initial={{ opacity: 0, y: reduce ? 0 : 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.4, ease: EASE }}
      className="flex flex-col gap-5 border-t border-paper/10 py-8 last:border-b md:flex-row md:items-start md:justify-between md:gap-10"
    >
      <div className="min-w-0">
        <p className={cn(MONO, 'text-concrete')}>{SPACE_LABELS[booking.space.type]}</p>
        <h3
          className={cn(
            'mt-2 break-words font-arch font-expanded text-2xl font-extrabold uppercase leading-[1] tracking-[-0.02em] md:text-3xl',
            muted ? 'text-paper/70' : 'text-paper',
          )}
        >
          {booking.space.name}
        </h3>
        <p className={cn(MONO, 'mt-4 text-paper')}>
          <span className="capitalize">{formatDate(booking.booking_date)}</span>
          <span className="text-concrete"> · </span>
          {booking.slot.label}
          <span className="text-concrete"> · </span>${formatPrice(booking.total_price)}
        </p>
        <p className={cn(MONO, 'mt-3 flex items-center gap-2', status.text)}>
          <span aria-hidden="true" className={cn('inline-block size-1.5 shrink-0', status.dot)} />
          {status.label}
        </p>
      </div>

      <div className="flex shrink-0 items-center gap-8">
        <button
          type="button"
          onClick={() => navigate(`/canchas/${booking.space.id}`)}
          className={cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto py-1 text-concrete hover:text-paper')}
        >
          Ver cancha
        </button>
        {canCancel && (
          <button
            type="button"
            onClick={() => onCancel(booking)}
            className={cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto py-1 text-paper')}
          >
            Cancelar
          </button>
        )}
      </div>
    </motion.li>
  );
}

export default function MyBookingsPage() {
  const queryClient = useQueryClient();
  const reduce = useReducedMotion();
  const [cancelTarget, setCancelTarget] = useState<BookingDetail | null>(null);

  const { data: bookings, isLoading } = useQuery({
    queryKey: ['my-bookings'],
    queryFn: getMyBookings,
  });

  const cancelMutation = useMutation({
    mutationFn: (id: number) => cancelBooking(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-bookings'] });
      setCancelTarget(null);
    },
  });

  const active = bookings?.filter((b) => b.status === 'pending' || b.status === 'confirmed') ?? [];
  const past   = bookings?.filter((b) => b.status === 'cancelled' || b.status === 'completed') ?? [];

  const y = reduce ? 0 : 12;

  return (
    <div className="landing">
      <div className="on-dark min-h-[100svh] bg-night text-paper">
        <LandingNav solid />

        <main className="relative pt-16 md:pt-20">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-x-0 top-16 h-[160px] bg-gradient-to-b from-light/10 to-transparent md:top-20 md:h-[200px]"
          />
          <LightLine className="absolute inset-x-0 top-16 md:top-20" delay={0.3} />

          <div className={cn(CONTAINER, 'relative pb-20 pt-14 md:pb-28 md:pt-20 lg:pb-32')}>
            <SectionLabel number="H1" label="Mis reservas" />

            <h1 className="mt-10 break-words font-arch font-expanded text-[clamp(2.25rem,min(9vw,14vh),7.5rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper md:mt-14">
              <motion.span
                className="block"
                initial={{ opacity: 0, y: reduce ? 0 : 28 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, ease: EASE, delay: 0.15 }}
              >
                Mis reservas
              </motion.span>
            </h1>

            <motion.div
              className="mt-14 md:mt-20"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, ease: EASE, delay: 0.35 }}
            >
              {isLoading ? (
                <div aria-busy="true" aria-label="Cargando reservas" className="flex flex-col gap-4">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="h-28 animate-pulse bg-graphite motion-reduce:animate-none" />
                  ))}
                </div>
              ) : !bookings?.length ? (
                <div className="max-w-md">
                  <p className="font-arch font-expanded text-xl font-bold uppercase leading-tight tracking-[-0.02em] text-paper md:text-2xl">
                    Sin reservas todavía
                  </p>
                  <p className="mt-4 text-base leading-relaxed text-paper/85 md:text-lg">
                    Explorá nuestras canchas y reservá tu primer turno.
                  </p>
                  <a
                    href="/canchas"
                    className={cn(buttonVariants({ variant: 'court' }), 'mt-8 h-14 px-8 text-base font-semibold')}
                  >
                    Ver canchas
                  </a>
                </div>
              ) : (
                <div className="flex flex-col gap-16 md:gap-20">
                  {/* Active bookings */}
                  {active.length > 0 && (
                    <section aria-labelledby="active-bookings-title">
                      <h2 id="active-bookings-title" className={cn(MONO, 'mb-6 text-paper')}>
                        Próximas
                        <span className="ml-3 text-concrete">{String(active.length).padStart(2, '0')}</span>
                      </h2>
                      <ul>
                        <AnimatePresence mode="popLayout">
                          {active.map((b) => (
                            <BookingRow key={b.id} booking={b} onCancel={setCancelTarget} />
                          ))}
                        </AnimatePresence>
                      </ul>
                    </section>
                  )}

                  {/* Past bookings */}
                  {past.length > 0 && (
                    <section aria-labelledby="past-bookings-title">
                      <h2 id="past-bookings-title" className={cn(MONO, 'mb-6 text-concrete')}>
                        Historial
                        <span className="ml-3">{String(past.length).padStart(2, '0')}</span>
                      </h2>
                      <ul>
                        {past.map((b) => (
                          <BookingRow key={b.id} booking={b} onCancel={setCancelTarget} muted />
                        ))}
                      </ul>
                    </section>
                  )}
                </div>
              )}
            </motion.div>
          </div>
        </main>

        <LandingFooter />
      </div>

      {/* Cancel dialog */}
      <AnimatePresence>
        {cancelTarget && (
          <ModalPortal>
            {/* Portal renders outside .landing, so re-scope the landing tokens (zero-size, no transform). */}
            <div className="landing on-dark">
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="fixed inset-0 z-40 bg-night/80"
                onClick={() => setCancelTarget(null)}
              />
              <div className="pointer-events-none fixed inset-0 z-50 flex items-end justify-center p-4 sm:items-center">
                <motion.div
                  role="dialog"
                  aria-modal="true"
                  aria-labelledby="cancel-dialog-title"
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: 16 }}
                  transition={{ duration: 0.22 }}
                  className="pointer-events-auto relative w-full border border-paper/10 bg-graphite p-6 text-paper sm:w-[420px] sm:p-8"
                >
                  <LightLine className="absolute inset-x-0 top-0" delay={0.1} />
                  <h3
                    id="cancel-dialog-title"
                    className="font-arch font-expanded text-2xl font-bold uppercase leading-none tracking-[-0.02em] text-paper"
                  >
                    ¿Cancelar reserva?
                  </h3>
                  <p className={cn(MONO, 'mt-6 text-concrete')}>
                    <span className="text-paper">{cancelTarget.space.name}</span>
                    {' · '}
                    <span className="capitalize">{formatDate(cancelTarget.booking_date)}</span>
                  </p>
                  {cancelMutation.error && (
                    <p role="alert" className={cn(MONO, 'mt-4 normal-case tracking-normal text-red-300')}>
                      {cancelMutation.error.message}
                    </p>
                  )}
                  <div className="mt-8 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <button
                      type="button"
                      onClick={() => setCancelTarget(null)}
                      className={cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto justify-center py-2 text-paper')}
                    >
                      Volver
                    </button>
                    <button
                      type="button"
                      onClick={() => cancelMutation.mutate(cancelTarget.id)}
                      disabled={cancelMutation.isPending}
                      className={cn(buttonVariants({ variant: 'court' }), 'h-12 justify-center px-6 text-sm font-semibold')}
                    >
                      {cancelMutation.isPending ? 'Cancelando...' : 'Sí, cancelar'}
                    </button>
                  </div>
                </motion.div>
              </div>
            </div>
          </ModalPortal>
        )}
      </AnimatePresence>
    </div>
  );
}
