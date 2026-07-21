import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  CalendarDays, Clock, User, Phone, XCircle, Plus, ChevronLeft, ChevronRight,
  Pencil, Trash2, ShieldCheck, Repeat, Wrench, Ban,
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  getAllBookings, cancelBooking, createManualBooking,
  createRecurringBooking, createMaintenanceBlock, getBatches, cancelBatch,
} from '../api/bookings';
import { getSpaces, getSlots, createSpace, updateSpace, deleteSpace } from '../api/spaces';
import { getUsers, createStaffUser, updateUserRole, deleteUser } from '../api/users';
import type { BookingDetail, BookingStatus, BookingBatch } from '../types';
import { SPACE_LABELS, SPACE_ICONS } from '../components/SpaceCard';
import { useAuth } from '../hooks/useAuth';
import { cn } from '@/lib/utils';

type Tab = 'bookings' | 'fixed' | 'spaces' | 'users';

const WEEKDAYS_SHORT = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
const WEEKDAYS_LONG = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];

const STATUS_CONFIG: Record<BookingStatus, { label: string; classes: string }> = {
  pending:   { label: 'Pendiente',  classes: 'bg-status-pending/10 text-status-pending' },
  confirmed: { label: 'Confirmada', classes: 'bg-status-confirmed/10 text-status-confirmed' },
  cancelled: { label: 'Cancelada',  classes: 'bg-status-cancelled/10 text-status-cancelled' },
  completed: { label: 'Completada', classes: 'bg-status-completed/10 text-status-completed' },
};

const TZ = 'America/Argentina/Buenos_Aires';

function toArgISO(d: Date): string {
  return d.toLocaleDateString('en-CA', { timeZone: TZ });
}

function formatDate(iso: string) {
  return new Date(iso + 'T12:00:00').toLocaleDateString('es-AR', {
    weekday: 'long', day: 'numeric', month: 'long', timeZone: TZ,
  });
}

function shiftDate(iso: string, days: number) {
  const d = new Date(iso + 'T12:00:00');
  d.setDate(d.getDate() + days);
  return toArgISO(d);
}

const today = toArgISO(new Date());

// ── Booking card ──────────────────────────────────────────────────────────────
function BookingRow({
  booking,
  onCancel,
}: {
  booking: BookingDetail;
  onCancel: (b: BookingDetail) => void;
}) {
  const status = STATUS_CONFIG[booking.status];
  const canCancel = booking.status === 'pending' || booking.status === 'confirmed';

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.97 }}
      transition={{ duration: 0.25 }}
      className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm p-4 flex flex-col sm:flex-row sm:items-center gap-4"
    >
      {/* Space + slot */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-base font-bold text-ink truncate">{booking.space.name}</span>
          <span className={cn('shrink-0 inline-flex items-center rounded-full px-2 py-0.5 text-2xs font-bold uppercase tracking-wide', status.classes)}>
            {status.label}
          </span>
          {booking.batch && (
            <span className="shrink-0 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-2xs font-bold uppercase tracking-wide bg-primary/10 text-primary">
              {booking.batch.type === 'recurring_teacher' ? <Repeat size={10} /> : <Wrench size={10} />}
              {booking.batch.type === 'recurring_teacher' ? 'Turno fijo' : 'Mantenimiento'}
            </span>
          )}
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-ink-2">
          <span className="flex items-center gap-1.5">
            <Clock size={13} className="text-primary/50" />
            {booking.slot.label}
          </span>
          <span className="text-2xs font-bold uppercase tracking-wide text-ink-2/60">
            {SPACE_LABELS[booking.space.type]}
          </span>
        </div>
        {booking.batch && (
          <p className="text-sm text-ink-2 mt-1 truncate">{booking.batch.reason}</p>
        )}
      </div>

      {/* Customer */}
      <div className="flex flex-col gap-0.5 sm:w-44 shrink-0">
        <span className="flex items-center gap-1.5 text-sm font-semibold text-ink">
          <User size={13} className="text-primary/50 shrink-0" />
          {booking.customer.name}
        </span>
        <span className="flex items-center gap-1.5 text-sm text-ink-2">
          <Phone size={13} className="text-primary/50 shrink-0" />
          {booking.customer.phone}
        </span>
      </div>

      {/* Price + cancel */}
      <div className="flex items-center justify-between sm:justify-end gap-4 shrink-0">
        <span className="font-serif text-xl text-ink">
          ${booking.total_price.toLocaleString('es-AR')}
        </span>
        {canCancel && (
          <button
            onClick={() => onCancel(booking)}
            className="flex items-center gap-1 text-sm font-semibold text-status-cancelled hover:bg-status-cancelled/8 px-2.5 py-1.5 rounded-lg transition-all active:scale-[0.98]"
          >
            <XCircle size={14} /> Cancelar
          </button>
        )}
      </div>
    </motion.div>
  );
}

// ── Manual booking dialog ─────────────────────────────────────────────────────
function ManualBookingDialog({
  defaultDate,
  onClose,
}: {
  defaultDate: string;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const [spaceId, setSpaceId] = useState<number | ''>('');
  const [date, setDate] = useState(defaultDate);
  const [slotId, setSlotId] = useState<number | ''>('');
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');

  const { data: spaces } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });

  const { data: slots } = useQuery({
    queryKey: ['slots', spaceId, date],
    queryFn: () => getSlots(Number(spaceId), date),
    enabled: !!spaceId && !!date,
  });

  const mutation = useMutation({
    mutationFn: () =>
      createManualBooking({
        space_id: Number(spaceId),
        slot_id: Number(slotId),
        booking_date: date,
        customer_name: name,
        customer_phone: phone,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['all-bookings'] });
      onClose();
    },
  });

  const inputClass =
    'w-full px-3.5 py-2.5 bg-surface border-[1.5px] border-black/10 rounded-lg text-sm font-medium text-ink placeholder:text-ink-2/60 outline-none transition-all duration-normal focus:border-primary focus:ring-2 focus:ring-primary/[0.13]';

  return (
    <>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/40 z-40 backdrop-blur-sm"
        onClick={onClose}
      />
      <motion.div
        initial={{ opacity: 0, scale: 0.94, y: 16 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.94, y: 16 }}
        transition={{ duration: 0.22 }}
        className="fixed inset-x-4 bottom-4 sm:inset-auto sm:left-1/2 sm:-translate-x-1/2 sm:top-1/2 sm:-translate-y-1/2 sm:w-[460px] z-50 bg-white rounded-2xl shadow-xl p-6 max-h-[90vh] overflow-y-auto"
      >
        <h3 className="font-serif text-xl text-ink mb-5">Nueva reserva manual</h3>

        <form
          onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }}
          className="flex flex-col gap-4"
        >
          {/* Space */}
          <div>
            <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Espacio</label>
            <select
              value={spaceId}
              onChange={(e) => { setSpaceId(Number(e.target.value)); setSlotId(''); }}
              required
              className={inputClass}
            >
              <option value="">Seleccioná un espacio</option>
              {spaces?.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </div>

          {/* Date */}
          <div>
            <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Fecha</label>
            <input
              type="date"
              value={date}
              onChange={(e) => { setDate(e.target.value); setSlotId(''); }}
              required
              className={inputClass}
            />
          </div>

          {/* Slot */}
          <div>
            <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Turno</label>
            {!spaceId ? (
              <p className="text-sm text-ink-2/60 py-2">Seleccioná un espacio primero.</p>
            ) : !slots?.length ? (
              <p className="text-sm text-ink-2/60 py-2">Sin turnos disponibles para esta fecha.</p>
            ) : (
              <div className="grid grid-cols-2 gap-2">
                {slots.map((s) => (
                  <button
                    key={s.id}
                    type="button"
                    onClick={() => setSlotId(s.id)}
                    className={cn(
                      'py-2.5 rounded-lg border-[1.5px] text-sm font-semibold transition-all active:scale-[0.98]',
                      slotId === s.id
                        ? 'bg-primary/[0.08] border-primary text-primary'
                        : 'bg-surface border-black/[0.07] text-ink hover:border-primary/40',
                    )}
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Customer */}
          <div>
            <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Cliente</label>
            <div className="flex flex-col gap-2">
              <input
                type="text"
                placeholder="Nombre completo"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                className={inputClass}
              />
              <input
                type="tel"
                placeholder="Teléfono"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                required
                className={inputClass}
              />
            </div>
          </div>

          {mutation.error && (
            <p className="text-xs text-status-cancelled font-medium">{mutation.error.message}</p>
          )}

          <div className="flex gap-3 mt-1">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 bg-surface text-ink font-semibold rounded-lg py-3 text-sm transition-all hover:bg-surface-2 active:scale-[0.98]"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={!spaceId || !slotId || !name || !phone || mutation.isPending}
              className="flex-1 bg-primary text-white font-bold rounded-lg py-3 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {mutation.isPending ? 'Guardando...' : 'Confirmar reserva'}
            </button>
          </div>
        </form>
      </motion.div>
    </>
  );
}

// ── Recurring booking dialog (turno fijo semanal) ─────────────────────────────
function RecurringBookingDialog({ defaultDate, onClose }: { defaultDate: string; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [spaceId, setSpaceId] = useState<number | ''>('');
  const [weekday, setWeekday] = useState<number | null>(null);
  const [slotId, setSlotId] = useState<number | ''>('');
  const [startDate, setStartDate] = useState(defaultDate);
  const [endDate, setEndDate] = useState('');
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [note, setNote] = useState('');
  const [result, setResult] = useState<{ created: number; skipped: number } | null>(null);

  const { data: spaces } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });
  const { data: slots } = useQuery({
    queryKey: ['slots', spaceId, startDate],
    queryFn: () => getSlots(Number(spaceId), startDate),
    enabled: !!spaceId && !!startDate,
  });

  const mutation = useMutation({
    mutationFn: () =>
      createRecurringBooking({
        space_id: Number(spaceId),
        slot_id: Number(slotId),
        weekday: weekday!,
        customer_name: name,
        customer_phone: phone,
        start_date: startDate,
        end_date: endDate,
        note: note || undefined,
      }),
    onSuccess: (r) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] });
      setResult({ created: r.created_dates.length, skipped: r.skipped_dates.length });
    },
  });

  const setOneYear = () => {
    if (!startDate) return;
    const d = new Date(startDate + 'T12:00:00');
    d.setFullYear(d.getFullYear() + 1);
    setEndDate(d.toISOString().slice(0, 10));
  };

  const inputClass =
    'w-full px-3.5 py-2.5 bg-surface border-[1.5px] border-black/10 rounded-lg text-sm font-medium text-ink placeholder:text-ink-2/60 outline-none transition-all duration-normal focus:border-primary focus:ring-2 focus:ring-primary/[0.13]';

  return (
    <>
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/40 z-40 backdrop-blur-sm" onClick={onClose} />
      <motion.div initial={{ opacity: 0, scale: 0.94, y: 16 }} animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.94, y: 16 }} transition={{ duration: 0.22 }}
        className="fixed inset-x-4 bottom-4 sm:inset-auto sm:left-1/2 sm:-translate-x-1/2 sm:top-1/2 sm:-translate-y-1/2 sm:w-[460px] z-50 bg-white rounded-2xl shadow-xl p-6 max-h-[90vh] overflow-y-auto"
      >
        {result ? (
          <>
            <h3 className="font-serif text-xl text-ink mb-2">Turno fijo creado</h3>
            <p className="text-sm text-ink-2 mb-1">
              Se reservaron <strong className="text-ink">{result.created}</strong> fechas.
            </p>
            {result.skipped > 0 && (
              <p className="text-sm text-status-pending mb-4">
                {result.skipped} fechas ya estaban ocupadas y se saltearon — revisá la lista para ver cuáles.
              </p>
            )}
            <button onClick={onClose} className="w-full bg-primary text-white font-bold rounded-lg py-3 text-sm mt-3 hover:bg-primary-dark active:scale-[0.98] transition-all">
              Listo
            </button>
          </>
        ) : (
          <>
            <h3 className="font-serif text-xl text-ink mb-1">Nuevo turno fijo</h3>
            <p className="text-sm text-ink-2 mb-5">Repite el mismo turno todas las semanas, para un profesor u otro cliente fijo.</p>

            <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }} className="flex flex-col gap-4">
              <div>
                <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Espacio</label>
                <select value={spaceId} onChange={(e) => { setSpaceId(Number(e.target.value)); setSlotId(''); }} required className={inputClass}>
                  <option value="">Seleccioná un espacio</option>
                  {spaces?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>

              <div>
                <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Día de la semana</label>
                <div className="grid grid-cols-7 gap-1.5">
                  {WEEKDAYS_SHORT.map((label, i) => (
                    <button key={i} type="button" onClick={() => setWeekday(i)}
                      className={cn(
                        'py-2 rounded-lg border-[1.5px] text-2xs font-bold transition-all active:scale-[0.96]',
                        weekday === i ? 'bg-primary/[0.08] border-primary text-primary' : 'bg-surface border-black/[0.07] text-ink-2 hover:border-primary/40',
                      )}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Turno</label>
                {!spaceId ? (
                  <p className="text-sm text-ink-2/60 py-2">Seleccioná un espacio primero.</p>
                ) : !slots?.length ? (
                  <p className="text-sm text-ink-2/60 py-2">Sin turnos para este espacio.</p>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    {slots.map((s) => (
                      <button key={s.id} type="button" onClick={() => setSlotId(s.id)}
                        className={cn(
                          'py-2.5 rounded-lg border-[1.5px] text-sm font-semibold transition-all active:scale-[0.98]',
                          slotId === s.id ? 'bg-primary/[0.08] border-primary text-primary' : 'bg-surface border-black/[0.07] text-ink hover:border-primary/40',
                        )}
                      >
                        {s.label}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Desde</label>
                  <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className={inputClass} />
                </div>
                <div>
                  <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Hasta</label>
                  <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className={inputClass} />
                </div>
              </div>
              <button type="button" onClick={setOneYear} className="text-xs font-semibold text-primary hover:underline self-start -mt-2">
                Poner 1 año desde la fecha de inicio
              </button>

              <div>
                <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Cliente / profesor</label>
                <div className="flex flex-col gap-2">
                  <input type="text" placeholder="Nombre completo" value={name} onChange={(e) => setName(e.target.value)} required className={inputClass} />
                  <input type="tel" placeholder="Teléfono" value={phone} onChange={(e) => setPhone(e.target.value)} required className={inputClass} />
                  <input type="text" placeholder="Nota (opcional, ej. 'Clases de pádel')" value={note} onChange={(e) => setNote(e.target.value)} className={inputClass} />
                </div>
              </div>

              {mutation.error && <p className="text-xs text-status-cancelled font-medium">{mutation.error.message}</p>}

              <div className="flex gap-3 mt-1">
                <button type="button" onClick={onClose} className="flex-1 bg-surface text-ink font-semibold rounded-lg py-3 text-sm transition-all hover:bg-surface-2 active:scale-[0.98]">
                  Cancelar
                </button>
                <button type="submit" disabled={!spaceId || weekday === null || !slotId || !startDate || !endDate || !name || !phone || mutation.isPending}
                  className="flex-1 bg-primary text-white font-bold rounded-lg py-3 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {mutation.isPending ? 'Guardando...' : 'Crear turno fijo'}
                </button>
              </div>
            </form>
          </>
        )}
      </motion.div>
    </>
  );
}

// ── Maintenance block dialog ───────────────────────────────────────────────────
function MaintenanceBlockDialog({ defaultDate, onClose }: { defaultDate: string; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [spaceId, setSpaceId] = useState<number | ''>('');
  const [wholeSpace, setWholeSpace] = useState(true);
  const [slotId, setSlotId] = useState<number | ''>('');
  const [startDate, setStartDate] = useState(defaultDate);
  const [endDate, setEndDate] = useState(defaultDate);
  const [reason, setReason] = useState('');
  const [result, setResult] = useState<{ created: number; skipped: number } | null>(null);

  const { data: spaces } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });
  const { data: slots } = useQuery({
    queryKey: ['slots', spaceId, startDate],
    queryFn: () => getSlots(Number(spaceId), startDate),
    enabled: !!spaceId && !!startDate && !wholeSpace,
  });

  const mutation = useMutation({
    mutationFn: () =>
      createMaintenanceBlock({
        space_id: Number(spaceId),
        slot_id: wholeSpace ? null : Number(slotId),
        start_date: startDate,
        end_date: endDate,
        reason,
      }),
    onSuccess: (r) => {
      queryClient.invalidateQueries({ queryKey: ['batches'] });
      setResult({ created: r.created_dates.length, skipped: r.skipped_dates.length });
    },
  });

  const inputClass =
    'w-full px-3.5 py-2.5 bg-surface border-[1.5px] border-black/10 rounded-lg text-sm font-medium text-ink placeholder:text-ink-2/60 outline-none transition-all duration-normal focus:border-primary focus:ring-2 focus:ring-primary/[0.13]';

  return (
    <>
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/40 z-40 backdrop-blur-sm" onClick={onClose} />
      <motion.div initial={{ opacity: 0, scale: 0.94, y: 16 }} animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.94, y: 16 }} transition={{ duration: 0.22 }}
        className="fixed inset-x-4 bottom-4 sm:inset-auto sm:left-1/2 sm:-translate-x-1/2 sm:top-1/2 sm:-translate-y-1/2 sm:w-[460px] z-50 bg-white rounded-2xl shadow-xl p-6 max-h-[90vh] overflow-y-auto"
      >
        {result ? (
          <>
            <h3 className="font-serif text-xl text-ink mb-2">Bloqueo creado</h3>
            <p className="text-sm text-ink-2 mb-1">
              Se bloquearon <strong className="text-ink">{result.created}</strong> turnos.
            </p>
            {result.skipped > 0 && (
              <p className="text-sm text-status-pending mb-4">
                {result.skipped} ya tenían una reserva y se saltearon — esa reserva sigue en pie, contactá al cliente si hace falta reprogramarla.
              </p>
            )}
            <button onClick={onClose} className="w-full bg-primary text-white font-bold rounded-lg py-3 text-sm mt-3 hover:bg-primary-dark active:scale-[0.98] transition-all">
              Listo
            </button>
          </>
        ) : (
          <>
            <h3 className="font-serif text-xl text-ink mb-1">Nuevo bloqueo por mantenimiento</h3>
            <p className="text-sm text-ink-2 mb-5">Bloquea uno o todos los turnos de una cancha durante un rango de días.</p>

            <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }} className="flex flex-col gap-4">
              <div>
                <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Espacio</label>
                <select value={spaceId} onChange={(e) => { setSpaceId(Number(e.target.value)); setSlotId(''); }} required className={inputClass}>
                  <option value="">Seleccioná un espacio</option>
                  {spaces?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>

              <div className="flex gap-2">
                <button type="button" onClick={() => setWholeSpace(true)}
                  className={cn('flex-1 py-2.5 rounded-lg border-[1.5px] text-sm font-semibold transition-all active:scale-[0.98]',
                    wholeSpace ? 'bg-primary/[0.08] border-primary text-primary' : 'bg-surface border-black/[0.07] text-ink-2 hover:border-primary/40')}
                >
                  Todos los turnos
                </button>
                <button type="button" onClick={() => setWholeSpace(false)}
                  className={cn('flex-1 py-2.5 rounded-lg border-[1.5px] text-sm font-semibold transition-all active:scale-[0.98]',
                    !wholeSpace ? 'bg-primary/[0.08] border-primary text-primary' : 'bg-surface border-black/[0.07] text-ink-2 hover:border-primary/40')}
                >
                  Un turno puntual
                </button>
              </div>

              {!wholeSpace && (
                <div>
                  <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Turno</label>
                  {!spaceId ? (
                    <p className="text-sm text-ink-2/60 py-2">Seleccioná un espacio primero.</p>
                  ) : !slots?.length ? (
                    <p className="text-sm text-ink-2/60 py-2">Sin turnos para este espacio.</p>
                  ) : (
                    <div className="grid grid-cols-2 gap-2">
                      {slots.map((s) => (
                        <button key={s.id} type="button" onClick={() => setSlotId(s.id)}
                          className={cn(
                            'py-2.5 rounded-lg border-[1.5px] text-sm font-semibold transition-all active:scale-[0.98]',
                            slotId === s.id ? 'bg-primary/[0.08] border-primary text-primary' : 'bg-surface border-black/[0.07] text-ink hover:border-primary/40',
                          )}
                        >
                          {s.label}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              )}

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Desde</label>
                  <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className={inputClass} />
                </div>
                <div>
                  <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Hasta</label>
                  <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className={inputClass} />
                </div>
              </div>

              <div>
                <label className="text-2xs font-bold uppercase tracking-wide text-ink-2 mb-1.5 block">Motivo</label>
                <input type="text" placeholder="Ej. Cambio de red, poda, pintura..." value={reason} onChange={(e) => setReason(e.target.value)} required className={inputClass} />
              </div>

              {mutation.error && <p className="text-xs text-status-cancelled font-medium">{mutation.error.message}</p>}

              <div className="flex gap-3 mt-1">
                <button type="button" onClick={onClose} className="flex-1 bg-surface text-ink font-semibold rounded-lg py-3 text-sm transition-all hover:bg-surface-2 active:scale-[0.98]">
                  Cancelar
                </button>
                <button type="submit" disabled={!spaceId || (!wholeSpace && !slotId) || !startDate || !endDate || !reason || mutation.isPending}
                  className="flex-1 bg-primary text-white font-bold rounded-lg py-3 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {mutation.isPending ? 'Guardando...' : 'Crear bloqueo'}
                </button>
              </div>
            </form>
          </>
        )}
      </motion.div>
    </>
  );
}

// ── Fixed slots panel (turnos fijos + bloqueos) ────────────────────────────────
function BatchRow({ batch, onCancel }: { batch: BookingBatch; onCancel: (b: BookingBatch) => void }) {
  const isRecurring = batch.type === 'recurring_teacher';
  return (
    <motion.div layout initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, scale: 0.97 }}
      transition={{ duration: 0.25 }}
      className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm p-4 flex flex-col sm:flex-row sm:items-center gap-4"
    >
      <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
        {isRecurring ? <Repeat size={16} className="text-primary" /> : <Wrench size={16} className="text-primary" />}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-base font-bold text-ink truncate">{batch.space.name}</span>
          <span className={cn('shrink-0 inline-flex items-center rounded-full px-2 py-0.5 text-2xs font-bold uppercase tracking-wide',
            batch.status === 'active' ? 'bg-status-confirmed/10 text-status-confirmed' : 'bg-status-cancelled/10 text-status-cancelled')}>
            {batch.status === 'active' ? 'Activo' : 'Cancelado'}
          </span>
        </div>
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-ink-2">
          <span className="flex items-center gap-1.5">
            <Clock size={13} className="text-primary/50" />
            {batch.slot?.label ?? 'Todos los turnos'}
          </span>
          <span className="text-2xs font-bold uppercase tracking-wide text-ink-2/60">
            {isRecurring && batch.weekday !== undefined
              ? `Todos los ${WEEKDAYS_LONG[batch.weekday]}s`
              : `${batch.start_date} → ${batch.end_date}`}
          </span>
        </div>
        <p className="text-sm text-ink mt-1 truncate">{batch.reason}</p>
      </div>

      {batch.status === 'active' && (
        <button onClick={() => onCancel(batch)}
          className="flex items-center gap-1 text-sm font-semibold text-status-cancelled hover:bg-status-cancelled/8 px-2.5 py-1.5 rounded-lg transition-all active:scale-[0.98] shrink-0 self-start sm:self-center"
        >
          <Ban size={14} /> Cancelar serie
        </button>
      )}
    </motion.div>
  );
}

function FixedSlotsPanel({ defaultDate }: { defaultDate: string }) {
  const queryClient = useQueryClient();
  const [showRecurring, setShowRecurring] = useState(false);
  const [showBlock, setShowBlock] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<BookingBatch | null>(null);

  const { data: batches, isLoading } = useQuery({ queryKey: ['batches'], queryFn: () => getBatches() });

  const cancelMutation = useMutation({
    mutationFn: (id: number) => cancelBatch(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['batches'] });
      queryClient.invalidateQueries({ queryKey: ['all-bookings'] });
      setCancelTarget(null);
    },
  });

  const active = batches?.filter((b) => b.status === 'active') ?? [];
  const cancelled = batches?.filter((b) => b.status === 'cancelled') ?? [];

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-primary">Turnos fijos y bloqueos</h2>
        <div className="flex gap-2">
          <button onClick={() => setShowRecurring(true)} className="flex items-center gap-1.5 bg-primary text-white font-bold rounded-lg px-4 py-2 text-sm transition-all hover:bg-primary-dark active:scale-[0.98]">
            <Repeat size={14} /> Turno fijo
          </button>
          <button onClick={() => setShowBlock(true)} className="flex items-center gap-1.5 bg-surface text-ink font-bold rounded-lg px-4 py-2 text-sm border-[1.5px] border-black/[0.07] transition-all hover:border-primary/40 active:scale-[0.98]">
            <Wrench size={14} /> Bloqueo
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex flex-col gap-3">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-20 bg-white rounded-xl border-[1.5px] border-black/[0.07] animate-pulse" />)}</div>
      ) : !batches?.length ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-14 h-14 rounded-xl bg-surface flex items-center justify-center mb-4">
            <Repeat size={24} className="text-ink-2/50" />
          </div>
          <p className="text-base font-semibold text-ink mb-1">Sin turnos fijos ni bloqueos</p>
          <p className="text-sm text-ink-2">Creá uno con los botones de arriba.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-8">
          {active.length > 0 && (
            <section>
              <h3 className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-3">Activos ({active.length})</h3>
              <AnimatePresence mode="popLayout">
                <div className="flex flex-col gap-3">
                  {active.map((b) => <BatchRow key={b.id} batch={b} onCancel={setCancelTarget} />)}
                </div>
              </AnimatePresence>
            </section>
          )}
          {cancelled.length > 0 && (
            <section>
              <h3 className="text-2xs font-bold tracking-[0.1em] uppercase text-ink-2 mb-3">Cancelados ({cancelled.length})</h3>
              <div className="flex flex-col gap-3">
                {cancelled.map((b) => <BatchRow key={b.id} batch={b} onCancel={setCancelTarget} />)}
              </div>
            </section>
          )}
        </div>
      )}

      <AnimatePresence>
        {cancelTarget && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/40 z-40 backdrop-blur-sm" onClick={() => setCancelTarget(null)} />
            <motion.div initial={{ opacity: 0, scale: 0.94, y: 16 }} animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.94, y: 16 }} transition={{ duration: 0.22 }}
              className="fixed inset-x-4 bottom-6 sm:inset-auto sm:left-1/2 sm:-translate-x-1/2 sm:top-1/2 sm:-translate-y-1/2 sm:w-[400px] z-50 bg-white rounded-2xl shadow-xl p-6"
            >
              <h3 className="font-serif text-xl text-ink mb-1">¿Cancelar toda la serie?</h3>
              <p className="text-sm text-ink-2 mb-6">
                Se van a cancelar todas las ocurrencias futuras de <strong className="text-ink">{cancelTarget.space.name}</strong>
                {' — '}{cancelTarget.reason}. Las que ya pasaron quedan como historial.
              </p>
              {cancelMutation.error && <p className="text-xs text-status-cancelled mb-3">{cancelMutation.error.message}</p>}
              <div className="flex gap-3">
                <button onClick={() => setCancelTarget(null)} className="flex-1 bg-surface text-ink font-semibold rounded-lg py-3 text-sm transition-all hover:bg-surface-2 active:scale-[0.98]">
                  Volver
                </button>
                <button onClick={() => cancelMutation.mutate(cancelTarget.id)} disabled={cancelMutation.isPending}
                  className="flex-1 bg-status-cancelled/10 border-[1.5px] border-status-cancelled/20 text-status-cancelled font-bold rounded-lg py-3 text-sm transition-all hover:bg-status-cancelled/18 active:scale-[0.98] disabled:opacity-60"
                >
                  {cancelMutation.isPending ? 'Cancelando...' : 'Sí, cancelar todo'}
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {showRecurring && <RecurringBookingDialog defaultDate={defaultDate} onClose={() => setShowRecurring(false)} />}
        {showBlock && <MaintenanceBlockDialog defaultDate={defaultDate} onClose={() => setShowBlock(false)} />}
      </AnimatePresence>
    </div>
  );
}

// ── Space form (must be outside SpacesPanel to avoid focus loss on re-render) ──
const spaceInputClass = 'w-full px-3.5 py-2.5 bg-surface border-[1.5px] border-black/10 rounded-lg text-sm font-medium text-ink placeholder:text-ink-2/60 outline-none transition-all duration-normal focus:border-primary focus:ring-2 focus:ring-primary/[0.13]';

function SpaceForm({ form, setForm, showNew, onCancel, onSubmit, isPending, submitLabel }: {
  form: { name: string; type: string; description: string; price_per_slot: string };
  setForm: React.Dispatch<React.SetStateAction<{ name: string; type: string; description: string; price_per_slot: string }>>;
  showNew: boolean;
  onCancel: () => void;
  onSubmit: () => void;
  isPending: boolean;
  submitLabel: string;
}) {
  return (
    <div className="flex flex-col gap-3 mt-3 p-4 bg-bg rounded-xl border border-black/[0.06]">
      <input placeholder="Nombre" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className={spaceInputClass} />
      {showNew && (
        <select value={form.type} onChange={e => setForm(f => ({ ...f, type: e.target.value }))} className={spaceInputClass}>
          <option value="cancha_padel">Pádel</option>
          <option value="cancha_futbol">Fútbol</option>
          <option value="quincho">Quincho / Salón</option>
        </select>
      )}
      <input placeholder="Descripción (opcional)" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className={spaceInputClass} />
      <input type="number" placeholder="Precio por turno" value={form.price_per_slot} onChange={e => setForm(f => ({ ...f, price_per_slot: e.target.value }))} className={spaceInputClass} />
      <div className="flex gap-2">
        <button onClick={onCancel} className="flex-1 bg-surface text-ink font-semibold rounded-lg py-2.5 text-sm hover:bg-surface-2 active:scale-[0.98] transition-all">Cancelar</button>
        <button onClick={onSubmit} disabled={isPending || !form.name || !form.price_per_slot} className="flex-1 bg-primary text-white font-bold rounded-lg py-2.5 text-sm hover:bg-primary-dark active:scale-[0.98] transition-all disabled:opacity-50">{isPending ? 'Guardando...' : submitLabel}</button>
      </div>
    </div>
  );
}

// ── Spaces panel (admin only) ─────────────────────────────────────────────────
function SpacesPanel() {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState<number | null>(null);
  const [showNew, setShowNew] = useState(false);

  const { data: spaces, isLoading } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });

  const [form, setForm] = useState({ name: '', type: 'cancha_padel', description: '', price_per_slot: '' });

  const createMutation = useMutation({
    mutationFn: () => createSpace({ name: form.name, type: form.type as any, description: form.description, price_per_slot: Number(form.price_per_slot) }),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['spaces'] }); setShowNew(false); setForm({ name: '', type: 'cancha_padel', description: '', price_per_slot: '' }); },
  });

  const updateMutation = useMutation({
    mutationFn: (id: number) => updateSpace(id, { name: form.name, description: form.description, price_per_slot: Number(form.price_per_slot) }),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['spaces'] }); setEditing(null); },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteSpace(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['spaces'] }),
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-primary">Espacios</h2>
        {!showNew && <button onClick={() => { setShowNew(true); setEditing(null); }} className="flex items-center gap-1.5 text-sm font-semibold text-primary hover:underline"><Plus size={14} /> Nuevo espacio</button>}
      </div>
      {showNew && <SpaceForm form={form} setForm={setForm} showNew={showNew} onCancel={() => { setShowNew(false); setEditing(null); }} onSubmit={() => createMutation.mutate()} isPending={createMutation.isPending} submitLabel="Crear espacio" />}
      {isLoading ? (
        <div className="flex flex-col gap-3">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-16 bg-white rounded-xl animate-pulse" />)}</div>
      ) : (
        <div className="flex flex-col gap-3">
          {spaces?.map(space => {
            const Icon = SPACE_ICONS[space.type];
            return (
              <div key={space.id} className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm p-4">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"><Icon size={16} className="text-primary" /></div>
                    <div className="min-w-0">
                      <p className="text-sm font-bold text-ink truncate">{space.name}</p>
                      <p className="text-xs text-ink-2">{SPACE_LABELS[space.type]} · <span className="font-serif">${space.price_per_slot.toLocaleString('es-AR')}</span> / turno</p>
                    </div>
                  </div>
                  <div className="flex gap-1 shrink-0">
                    <button onClick={() => { setEditing(space.id); setShowNew(false); setForm({ name: space.name, type: space.type, description: space.description ?? '', price_per_slot: String(space.price_per_slot) }); }} className="p-2 rounded-lg text-ink-2 hover:text-primary hover:bg-primary/8 transition-all active:scale-[0.98]"><Pencil size={15} /></button>
                    <button onClick={() => deleteMutation.mutate(space.id)} className="p-2 rounded-lg text-ink-2 hover:text-status-cancelled hover:bg-status-cancelled/8 transition-all active:scale-[0.98]"><Trash2 size={15} /></button>
                  </div>
                </div>
                {editing === space.id && <SpaceForm form={form} setForm={setForm} showNew={false} onCancel={() => setEditing(null)} onSubmit={() => updateMutation.mutate(space.id)} isPending={updateMutation.isPending} submitLabel="Guardar cambios" />}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ── Users panel (admin only) ──────────────────────────────────────────────────
function UsersPanel() {
  const queryClient = useQueryClient();
  const [showNew, setShowNew] = useState(false);
  const [form, setForm] = useState({ name: '', phone: '', password: '', role: 'receptionist' as 'receptionist' | 'admin' });

  const { data: users, isLoading } = useQuery({ queryKey: ['admin-users'], queryFn: getUsers });

  const createMutation = useMutation({
    mutationFn: () => createStaffUser({ ...form }),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['admin-users'] }); setShowNew(false); setForm({ name: '', phone: '', password: '', role: 'receptionist' }); },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteUser(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const roleMutation = useMutation({
    mutationFn: ({ id, role }: { id: number; role: string }) => updateUserRole(id, role),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const inputClass = 'w-full px-3.5 py-2.5 bg-surface border-[1.5px] border-black/10 rounded-lg text-sm font-medium text-ink placeholder:text-ink-2/60 outline-none transition-all duration-normal focus:border-primary focus:ring-2 focus:ring-primary/[0.13]';

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-primary">Usuarios staff</h2>
        {!showNew && <button onClick={() => setShowNew(true)} className="flex items-center gap-1.5 text-sm font-semibold text-primary hover:underline"><Plus size={14} /> Nuevo usuario</button>}
      </div>

      {showNew && (
        <div className="flex flex-col gap-3 mb-5 p-4 bg-bg rounded-xl border border-black/[0.06]">
          <input placeholder="Nombre" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className={inputClass} />
          <input type="tel" placeholder="Teléfono" value={form.phone} onChange={e => setForm(f => ({ ...f, phone: e.target.value }))} className={inputClass} />
          <input type="password" placeholder="Contraseña" value={form.password} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} className={inputClass} />
          <select value={form.role} onChange={e => setForm(f => ({ ...f, role: e.target.value as any }))} className={inputClass}>
            <option value="receptionist">Recepcionista</option>
            <option value="admin">Administrador</option>
          </select>
          {createMutation.error && <p className="text-xs text-status-cancelled">{createMutation.error.message}</p>}
          <div className="flex gap-2">
            <button onClick={() => setShowNew(false)} className="flex-1 bg-surface text-ink font-semibold rounded-lg py-2.5 text-sm hover:bg-surface-2 active:scale-[0.98] transition-all">Cancelar</button>
            <button onClick={() => createMutation.mutate()} disabled={createMutation.isPending || !form.name || !form.phone || !form.password} className="flex-1 bg-primary text-white font-bold rounded-lg py-2.5 text-sm hover:bg-primary-dark active:scale-[0.98] transition-all disabled:opacity-50">{createMutation.isPending ? 'Creando...' : 'Crear usuario'}</button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="flex flex-col gap-3">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-14 bg-white rounded-xl animate-pulse" />)}</div>
      ) : (
        <div className="flex flex-col gap-3">
          {users?.map(user => (
            <div key={user.id} className="bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm p-4 flex items-center justify-between gap-3">
              <div className="flex items-center gap-3 min-w-0">
                <div className="w-9 h-9 rounded-full bg-primary/10 flex items-center justify-center shrink-0"><User size={15} className="text-primary" /></div>
                <div className="min-w-0">
                  <p className="text-sm font-bold text-ink truncate">{user.name}</p>
                  <div className="flex items-center gap-2">
                    <p className="text-xs text-ink-2">{user.phone}</p>
                    <select
                      value={user.role}
                      onChange={(e) => roleMutation.mutate({ id: user.id, role: e.target.value })}
                      disabled={roleMutation.isPending}
                      className={cn(
                        'text-2xs font-bold uppercase tracking-wide rounded-full px-2 py-0.5 border-0 outline-none cursor-pointer transition-all',
                        user.role === 'admin'
                          ? 'bg-status-pending/10 text-status-pending'
                          : 'bg-primary/10 text-primary',
                      )}
                    >
                      <option value="receptionist">Recepcionista</option>
                      <option value="admin">Admin</option>
                    </select>
                  </div>
                </div>
              </div>
              <button onClick={() => deleteMutation.mutate(user.id)} className="p-2 rounded-lg text-ink-2 hover:text-status-cancelled hover:bg-status-cancelled/8 transition-all active:scale-[0.98] shrink-0"><Trash2 size={15} /></button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────
export default function StaffPanelPage() {
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [tab, setTab] = useState<Tab>('bookings');
  const [date, setDate] = useState(today);
  const [cancelTarget, setCancelTarget] = useState<BookingDetail | null>(null);
  const [showManual, setShowManual] = useState(false);

  const { data: bookings, isLoading } = useQuery({
    queryKey: ['all-bookings', date],
    queryFn: () => getAllBookings(date),
  });

  const cancelMutation = useMutation({
    mutationFn: (id: number) => cancelBooking(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['all-bookings'] });
      setCancelTarget(null);
    },
  });

  const active = bookings?.filter((b) => b.status === 'pending' || b.status === 'confirmed') ?? [];
  const past   = bookings?.filter((b) => b.status === 'cancelled' || b.status === 'completed') ?? [];

  return (
    <div className="animate-fade-up min-h-[calc(100vh-58px)] bg-bg">

      {/* Header */}
      <div className="bg-surface border-b border-black/[0.06]">
        <div className="max-w-[1140px] mx-auto px-6 py-8">
          <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
            <div>
              <span className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-1.5 block">
                {isAdmin ? 'Panel de administración' : 'Panel de recepción'}
              </span>
              <h1 className="font-serif text-4xl text-ink tracking-tight">
                {tab === 'bookings' ? 'Reservas del día' : tab === 'fixed' ? 'Turnos fijos y bloqueos' : tab === 'spaces' ? 'Canchas' : 'Usuarios'}
              </h1>
            </div>
            {tab === 'bookings' && (
              <button
                onClick={() => setShowManual(true)}
                className="flex items-center gap-2 bg-primary text-white font-bold rounded-lg px-5 py-2.5 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] shrink-0"
              >
                <Plus size={16} /> Nueva reserva
              </button>
            )}
          </div>

          {/* Tabs — reservas + turnos fijos para todo el staff; canchas + usuarios solo admin */}
          <div className="flex gap-1 mt-6 bg-surface rounded-lg p-1 w-fit">
            {([
              { key: 'bookings', label: 'Reservas',    icon: CalendarDays },
              { key: 'fixed',    label: 'Turnos fijos', icon: Repeat },
              ...(isAdmin ? [
                { key: 'spaces', label: 'Canchas',  icon: ShieldCheck },
                { key: 'users',  label: 'Usuarios', icon: User },
              ] as { key: Tab; label: string; icon: any }[] : []),
            ] as { key: Tab; label: string; icon: any }[]).map(({ key, label, icon: Icon }) => (
              <button
                key={key}
                onClick={() => setTab(key)}
                className={cn(
                  'flex items-center gap-1.5 px-4 py-2 rounded-md text-sm font-semibold transition-all duration-normal',
                  tab === key ? 'bg-white text-ink shadow-sm' : 'text-ink-2 hover:text-ink',
                )}
              >
                <Icon size={14} /> {label}
              </button>
            ))}
          </div>

          {/* Date nav — solo en tab reservas */}
          {tab === 'bookings' && <div className="flex items-center gap-3 mt-6">
            <button
              onClick={() => setDate((d) => shiftDate(d, -1))}
              className="w-9 h-9 rounded-lg bg-white border-[1.5px] border-black/[0.07] flex items-center justify-center text-ink-2 hover:border-primary/40 hover:text-primary transition-all active:scale-[0.98]"
            >
              <ChevronLeft size={17} />
            </button>

            <div className="flex items-center gap-2">
              <CalendarDays size={15} className="text-primary/60" />
              <input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className="text-sm font-semibold text-ink bg-transparent outline-none cursor-pointer"
              />
            </div>

            <button
              onClick={() => setDate((d) => shiftDate(d, 1))}
              className="w-9 h-9 rounded-lg bg-white border-[1.5px] border-black/[0.07] flex items-center justify-center text-ink-2 hover:border-primary/40 hover:text-primary transition-all active:scale-[0.98]"
            >
              <ChevronRight size={17} />
            </button>

            {date !== today && (
              <button
                onClick={() => setDate(today)}
                className="text-sm font-semibold text-primary hover:underline ml-1"
              >
                Hoy
              </button>
            )}
          </div>}
        </div>
      </div>

      {/* Content */}
      <div className="max-w-[1140px] mx-auto px-6 py-8">

        {/* Admin tabs content */}
        {tab === 'spaces' && <SpacesPanel />}
        {tab === 'users'  && <UsersPanel />}
        {tab === 'fixed'  && <FixedSlotsPanel defaultDate={date} />}

        {/* Bookings content */}
        {tab === 'bookings' && <>

        {/* Summary chips */}
        {!isLoading && bookings && bookings.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-6">
            <span className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 bg-white border-[1.5px] border-black/[0.07] text-sm font-semibold text-ink">
              {bookings.length} total
            </span>
            {active.length > 0 && (
              <span className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 bg-status-confirmed/10 text-status-confirmed text-sm font-semibold">
                {active.length} activas
              </span>
            )}
            {past.filter(b => b.status === 'cancelled').length > 0 && (
              <span className="inline-flex items-center gap-1.5 rounded-full px-3 py-1 bg-status-cancelled/10 text-status-cancelled text-sm font-semibold">
                {past.filter(b => b.status === 'cancelled').length} canceladas
              </span>
            )}
          </div>
        )}

        {isLoading ? (
          <div className="flex flex-col gap-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-20 bg-white rounded-xl border-[1.5px] border-black/[0.07] animate-pulse" />
            ))}
          </div>
        ) : !bookings?.length ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex flex-col items-center justify-center py-20 text-center"
          >
            <div className="w-14 h-14 rounded-xl bg-surface flex items-center justify-center mb-4">
              <CalendarDays size={24} className="text-ink-2/50" />
            </div>
            <p className="text-base font-semibold text-ink mb-1">Sin reservas</p>
            <p className="text-sm text-ink-2">
              No hay reservas para el{' '}
              <span className="capitalize">{formatDate(date)}</span>.
            </p>
          </motion.div>
        ) : (
          <div className="flex flex-col gap-8">
            {active.length > 0 && (
              <section>
                <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-3">
                  Activas ({active.length})
                </h2>
                <AnimatePresence mode="popLayout">
                  <div className="flex flex-col gap-3">
                    {active.map((b) => (
                      <BookingRow key={b.id} booking={b} onCancel={setCancelTarget} />
                    ))}
                  </div>
                </AnimatePresence>
              </section>
            )}
            {past.length > 0 && (
              <section>
                <h2 className="text-2xs font-bold tracking-[0.1em] uppercase text-ink-2 mb-3">
                  Historial ({past.length})
                </h2>
                <div className="flex flex-col gap-3">
                  {past.map((b) => (
                    <BookingRow key={b.id} booking={b} onCancel={setCancelTarget} />
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
        </>}
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
                {cancelTarget.customer.name}
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

      {/* Manual booking dialog */}
      <AnimatePresence>
        {showManual && (
          <ManualBookingDialog
            defaultDate={date}
            onClose={() => setShowManual(false)}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
