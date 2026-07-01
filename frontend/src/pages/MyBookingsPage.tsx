import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { CalendarDays, Clock, MapPin, XCircle, LayoutGrid } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { getMyBookings, cancelBooking } from '../api/bookings';
import type { BookingDetail, BookingStatus } from '../types';
import { SPACE_LABELS } from '../components/SpaceCard';
import { cn } from '@/lib/utils';

const STATUS_CONFIG: Record<BookingStatus, { label: string; classes: string }> = {
  pending:   { label: 'Pendiente',  classes: 'bg-status-pending/10 text-status-pending' },
  confirmed: { label: 'Confirmada', classes: 'bg-status-confirmed/10 text-status-confirmed' },
  cancelled: { label: 'Cancelada',  classes: 'bg-status-cancelled/10 text-status-cancelled' },
  completed: { label: 'Completada', classes: 'bg-status-completed/10 text-status-completed' },
};

function formatDate(iso: string) {
  return new Date(iso + 'T12:00:00').toLocaleDateString('es-AR', {
    weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
  });
}

function BookingCard({
  booking,
  onCancel,
}: {
  booking: BookingDetail;
  onCancel: (b: BookingDetail) => void;
}) {
  const navigate = useNavigate();
  const status = STATUS_CONFIG[booking.status];
  const canCancel = booking.status === 'pending' || booking.status === 'confirmed';

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.97 }}
      transition={{ duration: 0.3 }}
      className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm p-5 flex flex-col gap-4"
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-base font-bold text-ink">{booking.space.name}</h3>
          <span className="text-2xs font-bold uppercase tracking-wide text-ink-2">
            {SPACE_LABELS[booking.space.type]}
          </span>
        </div>
        <span className={cn('shrink-0 inline-flex items-center rounded-full px-2.5 py-1 text-2xs font-bold uppercase tracking-wide', status.classes)}>
          {status.label}
        </span>
      </div>

      {/* Details */}
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center gap-2 text-sm text-ink-2">
          <CalendarDays size={14} className="shrink-0 text-primary/60" />
          <span className="capitalize">{formatDate(booking.booking_date)}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-ink-2">
          <Clock size={14} className="shrink-0 text-primary/60" />
          <span>{booking.slot.label}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-ink-2">
          <MapPin size={14} className="shrink-0 text-primary/60" />
          <button
            onClick={() => navigate(`/canchas/${booking.space.id}`)}
            className="hover:text-primary hover:underline transition-colors"
          >
            Ver cancha
          </button>
        </div>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between pt-3 border-t border-black/[0.05]">
        <span className="font-serif text-2xl text-ink">
          ${booking.total_price.toLocaleString('es-AR')}
        </span>
        {canCancel && (
          <button
            onClick={() => onCancel(booking)}
            className="flex items-center gap-1.5 text-sm font-semibold text-status-cancelled hover:bg-status-cancelled/8 px-3 py-1.5 rounded-lg transition-all active:scale-[0.98]"
          >
            <XCircle size={14} /> Cancelar
          </button>
        )}
      </div>
    </motion.div>
  );
}

export default function MyBookingsPage() {
  const queryClient = useQueryClient();
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

  return (
    <div className="animate-fade-up min-h-[calc(100vh-58px)] bg-bg">

      {/* Header */}
      <div className="bg-surface border-b border-black/[0.06]">
        <div className="max-w-[1140px] mx-auto px-6 py-12">
          <span className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-1.5 block">
            Mi cuenta
          </span>
          <h1 className="font-serif text-4xl text-ink tracking-tight">Mis reservas</h1>
        </div>
      </div>

      <div className="max-w-[1140px] mx-auto px-6 py-10">

        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-52 bg-white rounded-xl border-[1.5px] border-black/[0.07] animate-pulse" />
            ))}
          </div>
        ) : !bookings?.length ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex flex-col items-center justify-center py-24 text-center"
          >
            <div className="w-14 h-14 rounded-xl bg-surface flex items-center justify-center mb-4">
              <LayoutGrid size={24} className="text-ink-2/50" />
            </div>
            <p className="text-base font-semibold text-ink mb-1">Sin reservas todavía</p>
            <p className="text-sm text-ink-2 mb-6">Explorá nuestras canchas y reservá tu primer turno.</p>
            <a
              href="/canchas"
              className="bg-primary text-white font-bold rounded-lg px-6 py-3 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98]"
            >
              Ver canchas
            </a>
          </motion.div>
        ) : (
          <div className="flex flex-col gap-10">

            {/* Active bookings */}
            {active.length > 0 && (
              <section>
                <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-4">
                  Próximas
                </h2>
                <AnimatePresence mode="popLayout">
                  <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
                    {active.map((b) => (
                      <BookingCard key={b.id} booking={b} onCancel={setCancelTarget} />
                    ))}
                  </div>
                </AnimatePresence>
              </section>
            )}

            {/* Past bookings */}
            {past.length > 0 && (
              <section>
                <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-ink-2 mb-4">
                  Historial
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
                  {past.map((b) => (
                    <BookingCard key={b.id} booking={b} onCancel={setCancelTarget} />
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </div>

      {/* Cancel dialog */}
      <AnimatePresence>
        {cancelTarget && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/40 z-40 backdrop-blur-sm"
              onClick={() => setCancelTarget(null)}
            />
            <motion.div
              initial={{ opacity: 0, scale: 0.94, y: 16 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.94, y: 16 }}
              transition={{ duration: 0.22 }}
              className="fixed inset-x-4 bottom-6 sm:inset-auto sm:left-1/2 sm:-translate-x-1/2 sm:top-1/2 sm:-translate-y-1/2 sm:w-[400px] z-50 bg-white rounded-2xl shadow-xl p-6"
            >
              <h3 className="font-serif text-xl text-ink mb-1">¿Cancelar reserva?</h3>
              <p className="text-sm text-ink-2 mb-6">
                <strong className="text-ink">{cancelTarget.space.name}</strong>
                {' · '}
                <span className="capitalize">{formatDate(cancelTarget.booking_date)}</span>
              </p>
              {cancelMutation.error && (
                <p className="text-xs text-status-cancelled mb-3">{cancelMutation.error.message}</p>
              )}
              <div className="flex gap-3">
                <button
                  onClick={() => setCancelTarget(null)}
                  className="flex-1 bg-surface text-ink font-semibold rounded-lg py-3 text-sm transition-all hover:bg-surface-2 active:scale-[0.98]"
                >
                  Volver
                </button>
                <button
                  onClick={() => cancelMutation.mutate(cancelTarget.id)}
                  disabled={cancelMutation.isPending}
                  className="flex-1 bg-status-cancelled/10 border-[1.5px] border-status-cancelled/20 text-status-cancelled font-bold rounded-lg py-3 text-sm transition-all hover:bg-status-cancelled/18 active:scale-[0.98] disabled:opacity-60"
                >
                  {cancelMutation.isPending ? 'Cancelando...' : 'Sí, cancelar'}
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}
