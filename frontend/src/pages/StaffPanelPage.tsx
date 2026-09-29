import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, ArrowRight, Pencil, Trash2 } from 'lucide-react';
import { motion, AnimatePresence, useReducedMotion } from 'framer-motion';
import {
  getAllBookings, cancelBooking, createManualBooking,
  createRecurringBooking, createMaintenanceBlock, getBatches, cancelBatch,
} from '../api/bookings';
import { getSpaces, getSlots, createSpace, updateSpace, deleteSpace } from '../api/spaces';
import { getUsers, createStaffUser, updateUserRole, deleteUser } from '../api/users';
import type { BookingDetail, BookingStatus, BookingBatch } from '../types';
import { SPACE_LABELS } from '../components/SpaceCard';
import { ModalPortal } from '../components/ModalPortal';
import { useAuth } from '../hooks/useAuth';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/components/ui/button';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { CONTAINER, EASE, MONO } from '@/components/landing/ui';

type Tab = 'bookings' | 'fixed' | 'spaces' | 'users';

const WEEKDAYS_SHORT = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
const WEEKDAYS_LONG = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];

const STATUS_CONFIG: Record<BookingStatus, { label: string; dot: string; text: string }> = {
  pending:   { label: 'Pendiente',  dot: 'bg-light',    text: 'text-light' },
  confirmed: { label: 'Confirmada', dot: 'bg-paper',    text: 'text-paper' },
  cancelled: { label: 'Cancelada',  dot: 'bg-concrete', text: 'text-concrete' },
  completed: { label: 'Completada', dot: 'bg-concrete', text: 'text-concrete' },
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

// ── Shared style tokens ───────────────────────────────────────────────────────
const FIELD =
  'w-full border-0 border-b border-paper/25 bg-transparent py-2.5 text-base text-paper [color-scheme:dark] placeholder:text-paper/40 outline-none transition-colors duration-300 focus:border-paper focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper disabled:opacity-50 [&_option]:bg-night [&_option]:text-paper';
const LABEL = cn(MONO, 'mb-1 block text-concrete');
const BTN_PRIMARY = cn(buttonVariants({ variant: 'court' }), 'h-11 justify-center px-6 text-sm font-semibold');
const BTN_LINE = cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto justify-center py-1 text-paper');
const BTN_LINE_MUTED = cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto justify-center py-1 text-concrete hover:text-paper');
const ERROR_TEXT = cn(MONO, 'normal-case tracking-normal text-red-300');
const TH = cn(MONO, 'whitespace-nowrap px-3 py-3 text-left font-normal text-concrete');
const TD = 'px-3 py-4 align-top';
const ROW = 'border-t border-paper/10 transition-colors last:border-b hover:bg-graphite/50';
const SECTION_TITLE =
  'font-arch font-expanded text-xl font-bold uppercase leading-none tracking-[-0.02em] text-paper md:text-2xl';
const DIALOG_TITLE =
  'font-arch font-expanded text-2xl font-bold uppercase leading-none tracking-[-0.02em] text-paper';

const choiceClass = (selected: boolean) =>
  cn(
    MONO,
    'border px-2 py-2.5 text-center outline-none transition-colors duration-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper',
    selected ? 'border-light text-paper' : 'border-paper/15 text-concrete hover:border-paper/40 hover:text-paper',
  );

/** Centered square night panel. Lives inside ModalPortal, so it re-scopes the landing tokens. */
function DialogPanel({
  onClose,
  labelledBy,
  wide = false,
  children,
}: {
  onClose: () => void;
  labelledBy: string;
  wide?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className="landing on-dark">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-40 bg-night/80"
        onClick={onClose}
      />
      <div className="pointer-events-none fixed inset-0 z-50 flex items-end justify-center p-4 sm:items-center">
        <motion.div
          role="dialog"
          aria-modal="true"
          aria-labelledby={labelledBy}
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 16 }}
          transition={{ duration: 0.22 }}
          className={cn(
            'pointer-events-auto relative max-h-[90svh] w-full overflow-y-auto border border-paper/10 bg-graphite p-6 text-paper sm:p-8',
            wide ? 'sm:w-[480px]' : 'sm:w-[420px]',
          )}
        >
          <LightLine className="absolute inset-x-0 top-0" delay={0.1} />
          {children}
        </motion.div>
      </div>
    </div>
  );
}

function BookingsTable({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[760px] border-collapse text-left">
        <thead>
          <tr>
            <th scope="col" className={TH}>Espacio</th>
            <th scope="col" className={TH}>Turno</th>
            <th scope="col" className={TH}>Cliente</th>
            <th scope="col" className={TH}>Estado</th>
            <th scope="col" className={cn(TH, 'text-right')}>Precio</th>
            <th scope="col" className={TH}><span className="sr-only">Acciones</span></th>
          </tr>
        </thead>
        {children}
      </table>
    </div>
  );
}

// ── Booking row ───────────────────────────────────────────────────────────────
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
    <motion.tr
      layout
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.3, ease: EASE }}
      className={ROW}
    >
      {/* Space + slot */}
      <td className={cn(TD, 'min-w-0')}>
        <p className="font-arch text-base font-semibold text-paper">{booking.space.name}</p>
        <p className={cn(MONO, 'mt-1 text-concrete')}>{SPACE_LABELS[booking.space.type]}</p>
        {booking.batch && (
          <p className={cn(MONO, 'mt-1 text-paper')}>
            {booking.batch.type === 'recurring_teacher' ? 'Turno fijo' : 'Mantenimiento'}
            <span className="text-concrete"> · {booking.batch.reason}</span>
          </p>
        )}
      </td>

      <td className={cn(TD, MONO, 'whitespace-nowrap text-paper')}>{booking.slot.label}</td>

      {/* Customer */}
      <td className={TD}>
        <p className="text-sm font-medium text-paper">{booking.customer.name}</p>
        <p className={cn(MONO, 'mt-1 text-concrete')}>{booking.customer.phone}</p>
      </td>

      <td className={cn(TD, MONO, 'whitespace-nowrap', status.text)}>
        <span aria-hidden="true" className={cn('mr-2 inline-block size-1.5 align-middle', status.dot)} />
        {status.label}
      </td>

      {/* Price + cancel */}
      <td className={cn(TD, MONO, 'whitespace-nowrap text-right text-paper')}>
        ${booking.total_price.toLocaleString('es-AR')}
      </td>
      <td className={cn(TD, 'whitespace-nowrap text-right')}>
        {canCancel && (
          <button type="button" onClick={() => onCancel(booking)} className={BTN_LINE}>
            Cancelar
          </button>
        )}
      </td>
    </motion.tr>
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

  return (
    <ModalPortal>
      <DialogPanel onClose={onClose} labelledBy="manual-dialog-title" wide>
        <h3 id="manual-dialog-title" className={cn(DIALOG_TITLE, 'mb-8')}>Nueva reserva manual</h3>

        <form
          onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }}
          className="flex flex-col gap-6"
        >
          {/* Space */}
          <div>
            <label htmlFor="manual-space" className={LABEL}>Espacio</label>
            <select
              id="manual-space"
              value={spaceId}
              onChange={(e) => { setSpaceId(Number(e.target.value)); setSlotId(''); }}
              required
              className={FIELD}
            >
              <option value="">Seleccioná un espacio</option>
              {spaces?.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </div>

          {/* Date */}
          <div>
            <label htmlFor="manual-date" className={LABEL}>Fecha</label>
            <input
              id="manual-date"
              type="date"
              value={date}
              onChange={(e) => { setDate(e.target.value); setSlotId(''); }}
              required
              className={FIELD}
            />
          </div>

          {/* Slot */}
          <div>
            <span className={LABEL}>Turno</span>
            {!spaceId ? (
              <p className={cn(MONO, 'py-2 text-concrete')}>Seleccioná un espacio primero.</p>
            ) : !slots?.length ? (
              <p className={cn(MONO, 'py-2 text-concrete')}>Sin turnos disponibles para esta fecha.</p>
            ) : (
              <div className="grid grid-cols-2 gap-2">
                {slots.map((s) => (
                  <button
                    key={s.id}
                    type="button"
                    aria-pressed={slotId === s.id}
                    onClick={() => setSlotId(s.id)}
                    className={choiceClass(slotId === s.id)}
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Customer */}
          <div>
            <span className={LABEL}>Cliente</span>
            <div className="flex flex-col gap-3">
              <input
                type="text"
                aria-label="Nombre completo"
                placeholder="Nombre completo"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                className={FIELD}
              />
              <input
                type="tel"
                aria-label="Teléfono"
                placeholder="Teléfono"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                required
                className={FIELD}
              />
            </div>
          </div>

          {mutation.error && (
            <p role="alert" className={ERROR_TEXT}>{mutation.error.message}</p>
          )}

          <div className="mt-2 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
            <button type="button" onClick={onClose} className={BTN_LINE}>
              Cancelar
            </button>
            <button
              type="submit"
              disabled={!spaceId || !slotId || !name || !phone || mutation.isPending}
              className={BTN_PRIMARY}
            >
              {mutation.isPending ? 'Guardando...' : 'Confirmar reserva'}
            </button>
          </div>
        </form>
      </DialogPanel>
    </ModalPortal>
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

  return (
    <ModalPortal>
      <DialogPanel onClose={onClose} labelledBy="recurring-dialog-title" wide>
        {result ? (
          <>
            <h3 id="recurring-dialog-title" className={cn(DIALOG_TITLE, 'mb-6')}>Turno fijo creado</h3>
            <p className="text-base leading-relaxed text-paper/85">
              Se reservaron <strong className="font-semibold text-paper">{result.created}</strong> fechas.
            </p>
            {result.skipped > 0 && (
              <p className="mt-3 text-base leading-relaxed text-light">
                {result.skipped} fechas ya estaban ocupadas y se saltearon — revisá la lista para ver cuáles.
              </p>
            )}
            <button type="button" onClick={onClose} className={cn(BTN_PRIMARY, 'mt-8 w-full')}>
              Listo
            </button>
          </>
        ) : (
          <>
            <h3 id="recurring-dialog-title" className={cn(DIALOG_TITLE, 'mb-4')}>Nuevo turno fijo</h3>
            <p className="mb-8 text-base leading-relaxed text-paper/85">Repite el mismo turno todas las semanas, para un profesor u otro cliente fijo.</p>

            <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }} className="flex flex-col gap-6">
              <div>
                <label htmlFor="recurring-space" className={LABEL}>Espacio</label>
                <select id="recurring-space" value={spaceId} onChange={(e) => { setSpaceId(Number(e.target.value)); setSlotId(''); }} required className={FIELD}>
                  <option value="">Seleccioná un espacio</option>
                  {spaces?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>

              <div>
                <span className={LABEL}>Día de la semana</span>
                <div className="grid grid-cols-7 gap-1.5">
                  {WEEKDAYS_SHORT.map((label, i) => (
                    <button key={i} type="button" aria-pressed={weekday === i} onClick={() => setWeekday(i)}
                      className={cn(choiceClass(weekday === i), 'px-0')}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <span className={LABEL}>Turno</span>
                {!spaceId ? (
                  <p className={cn(MONO, 'py-2 text-concrete')}>Seleccioná un espacio primero.</p>
                ) : !slots?.length ? (
                  <p className={cn(MONO, 'py-2 text-concrete')}>Sin turnos para este espacio.</p>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    {slots.map((s) => (
                      <button key={s.id} type="button" aria-pressed={slotId === s.id} onClick={() => setSlotId(s.id)}
                        className={choiceClass(slotId === s.id)}
                      >
                        {s.label}
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label htmlFor="recurring-start" className={LABEL}>Desde</label>
                  <input id="recurring-start" type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className={FIELD} />
                </div>
                <div>
                  <label htmlFor="recurring-end" className={LABEL}>Hasta</label>
                  <input id="recurring-end" type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className={FIELD} />
                </div>
              </div>
              <button type="button" onClick={setOneYear} className={cn(BTN_LINE_MUTED, 'self-start -mt-3')}>
                Poner 1 año desde la fecha de inicio
              </button>

              <div>
                <span className={LABEL}>Cliente / profesor</span>
                <div className="flex flex-col gap-3">
                  <input type="text" aria-label="Nombre completo" placeholder="Nombre completo" value={name} onChange={(e) => setName(e.target.value)} required className={FIELD} />
                  <input type="tel" aria-label="Teléfono" placeholder="Teléfono" value={phone} onChange={(e) => setPhone(e.target.value)} required className={FIELD} />
                  <input type="text" aria-label="Nota (opcional)" placeholder="Nota (opcional, ej. 'Clases de pádel')" value={note} onChange={(e) => setNote(e.target.value)} className={FIELD} />
                </div>
              </div>

              {mutation.error && <p role="alert" className={ERROR_TEXT}>{mutation.error.message}</p>}

              <div className="mt-2 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
                <button type="button" onClick={onClose} className={BTN_LINE}>
                  Cancelar
                </button>
                <button type="submit" disabled={!spaceId || weekday === null || !slotId || !startDate || !endDate || !name || !phone || mutation.isPending}
                  className={BTN_PRIMARY}
                >
                  {mutation.isPending ? 'Guardando...' : 'Crear turno fijo'}
                </button>
              </div>
            </form>
          </>
        )}
      </DialogPanel>
    </ModalPortal>
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

  return (
    <ModalPortal>
      <DialogPanel onClose={onClose} labelledBy="block-dialog-title" wide>
        {result ? (
          <>
            <h3 id="block-dialog-title" className={cn(DIALOG_TITLE, 'mb-6')}>Bloqueo creado</h3>
            <p className="text-base leading-relaxed text-paper/85">
              Se bloquearon <strong className="font-semibold text-paper">{result.created}</strong> turnos.
            </p>
            {result.skipped > 0 && (
              <p className="mt-3 text-base leading-relaxed text-light">
                {result.skipped} ya tenían una reserva y se saltearon — esa reserva sigue en pie, contactá al cliente si hace falta reprogramarla.
              </p>
            )}
            <button type="button" onClick={onClose} className={cn(BTN_PRIMARY, 'mt-8 w-full')}>
              Listo
            </button>
          </>
        ) : (
          <>
            <h3 id="block-dialog-title" className={cn(DIALOG_TITLE, 'mb-4')}>Nuevo bloqueo por mantenimiento</h3>
            <p className="mb-8 text-base leading-relaxed text-paper/85">Bloquea uno o todos los turnos de una cancha durante un rango de días.</p>

            <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(); }} className="flex flex-col gap-6">
              <div>
                <label htmlFor="block-space" className={LABEL}>Espacio</label>
                <select id="block-space" value={spaceId} onChange={(e) => { setSpaceId(Number(e.target.value)); setSlotId(''); }} required className={FIELD}>
                  <option value="">Seleccioná un espacio</option>
                  {spaces?.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </select>
              </div>

              <div className="flex gap-2">
                <button type="button" aria-pressed={wholeSpace} onClick={() => setWholeSpace(true)}
                  className={cn(choiceClass(wholeSpace), 'flex-1')}
                >
                  Todos los turnos
                </button>
                <button type="button" aria-pressed={!wholeSpace} onClick={() => setWholeSpace(false)}
                  className={cn(choiceClass(!wholeSpace), 'flex-1')}
                >
                  Un turno puntual
                </button>
              </div>

              {!wholeSpace && (
                <div>
                  <span className={LABEL}>Turno</span>
                  {!spaceId ? (
                    <p className={cn(MONO, 'py-2 text-concrete')}>Seleccioná un espacio primero.</p>
                  ) : !slots?.length ? (
                    <p className={cn(MONO, 'py-2 text-concrete')}>Sin turnos para este espacio.</p>
                  ) : (
                    <div className="grid grid-cols-2 gap-2">
                      {slots.map((s) => (
                        <button key={s.id} type="button" aria-pressed={slotId === s.id} onClick={() => setSlotId(s.id)}
                          className={choiceClass(slotId === s.id)}
                        >
                          {s.label}
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label htmlFor="block-start" className={LABEL}>Desde</label>
                  <input id="block-start" type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className={FIELD} />
                </div>
                <div>
                  <label htmlFor="block-end" className={LABEL}>Hasta</label>
                  <input id="block-end" type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className={FIELD} />
                </div>
              </div>

              <div>
                <label htmlFor="block-reason" className={LABEL}>Motivo</label>
                <input id="block-reason" type="text" placeholder="Ej. Cambio de red, poda, pintura..." value={reason} onChange={(e) => setReason(e.target.value)} required className={FIELD} />
              </div>

              {mutation.error && <p role="alert" className={ERROR_TEXT}>{mutation.error.message}</p>}

              <div className="mt-2 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
                <button type="button" onClick={onClose} className={BTN_LINE}>
                  Cancelar
                </button>
                <button type="submit" disabled={!spaceId || (!wholeSpace && !slotId) || !startDate || !endDate || !reason || mutation.isPending}
                  className={BTN_PRIMARY}
                >
                  {mutation.isPending ? 'Guardando...' : 'Crear bloqueo'}
                </button>
              </div>
            </form>
          </>
        )}
      </DialogPanel>
    </ModalPortal>
  );
}

// ── Fixed slots panel (turnos fijos + bloqueos) ────────────────────────────────
function BatchesTable({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[760px] border-collapse text-left">
        <thead>
          <tr>
            <th scope="col" className={TH}>Espacio</th>
            <th scope="col" className={TH}>Turno</th>
            <th scope="col" className={TH}>Detalle</th>
            <th scope="col" className={TH}>Motivo</th>
            <th scope="col" className={TH}>Estado</th>
            <th scope="col" className={TH}><span className="sr-only">Acciones</span></th>
          </tr>
        </thead>
        {children}
      </table>
    </div>
  );
}

function BatchRow({ batch, onCancel }: { batch: BookingBatch; onCancel: (b: BookingBatch) => void }) {
  const isRecurring = batch.type === 'recurring_teacher';
  return (
    <motion.tr layout initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
      transition={{ duration: 0.3, ease: EASE }}
      className={ROW}
    >
      <td className={TD}>
        <p className="font-arch text-base font-semibold text-paper">{batch.space.name}</p>
        <p className={cn(MONO, 'mt-1 text-concrete')}>{isRecurring ? 'Turno fijo' : 'Mantenimiento'}</p>
      </td>
      <td className={cn(TD, MONO, 'whitespace-nowrap text-paper')}>
        {batch.slot?.label ?? 'Todos los turnos'}
      </td>
      <td className={cn(TD, MONO, 'whitespace-nowrap text-concrete')}>
        {isRecurring && batch.weekday !== undefined
          ? `Todos los ${WEEKDAYS_LONG[batch.weekday]}s`
          : `${batch.start_date} → ${batch.end_date}`}
      </td>
      <td className={cn(TD, 'text-sm text-paper')}>{batch.reason}</td>
      <td className={cn(TD, MONO, 'whitespace-nowrap', batch.status === 'active' ? 'text-paper' : 'text-concrete')}>
        <span
          aria-hidden="true"
          className={cn('mr-2 inline-block size-1.5 align-middle', batch.status === 'active' ? 'bg-paper' : 'bg-concrete')}
        />
        {batch.status === 'active' ? 'Activo' : 'Cancelado'}
      </td>
      <td className={cn(TD, 'whitespace-nowrap text-right')}>
        {batch.status === 'active' && (
          <button type="button" onClick={() => onCancel(batch)} className={BTN_LINE}>
            Cancelar serie
          </button>
        )}
      </td>
    </motion.tr>
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
      <div className="mb-8 flex flex-wrap items-center justify-end gap-x-8 gap-y-4">
        <button type="button" onClick={() => setShowBlock(true)} className={BTN_LINE}>
          Bloqueo
        </button>
        <button type="button" onClick={() => setShowRecurring(true)} className={BTN_PRIMARY}>
          Turno fijo
        </button>
      </div>

      {isLoading ? (
        <div aria-busy="true" aria-label="Cargando turnos fijos" className="flex flex-col gap-4">
          {Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-16 animate-pulse bg-graphite motion-reduce:animate-none" />)}
        </div>
      ) : !batches?.length ? (
        <div className="max-w-md py-10">
          <p className="font-arch font-expanded text-xl font-bold uppercase leading-tight tracking-[-0.02em] text-paper">Sin turnos fijos ni bloqueos</p>
          <p className="mt-4 text-base leading-relaxed text-paper/85">Creá uno con los botones de arriba.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-16">
          {active.length > 0 && (
            <section aria-labelledby="batches-active-title">
              <h3 id="batches-active-title" className={cn(MONO, 'mb-4 text-paper')}>Activos ({active.length})</h3>
              <BatchesTable>
                <AnimatePresence mode="popLayout">
                  <tbody>
                    {active.map((b) => <BatchRow key={b.id} batch={b} onCancel={setCancelTarget} />)}
                  </tbody>
                </AnimatePresence>
              </BatchesTable>
            </section>
          )}
          {cancelled.length > 0 && (
            <section aria-labelledby="batches-cancelled-title">
              <h3 id="batches-cancelled-title" className={cn(MONO, 'mb-4 text-concrete')}>Cancelados ({cancelled.length})</h3>
              <BatchesTable>
                <tbody>
                  {cancelled.map((b) => <BatchRow key={b.id} batch={b} onCancel={setCancelTarget} />)}
                </tbody>
              </BatchesTable>
            </section>
          )}
        </div>
      )}

      <AnimatePresence>
        {cancelTarget && (
          <ModalPortal>
            <DialogPanel onClose={() => setCancelTarget(null)} labelledBy="cancel-batch-title">
              <h3 id="cancel-batch-title" className={DIALOG_TITLE}>¿Cancelar toda la serie?</h3>
              <p className="mt-6 text-base leading-relaxed text-paper/85">
                Se van a cancelar todas las ocurrencias futuras de <strong className="font-semibold text-paper">{cancelTarget.space.name}</strong>
                {' — '}{cancelTarget.reason}. Las que ya pasaron quedan como historial.
              </p>
              {cancelMutation.error && <p role="alert" className={cn(ERROR_TEXT, 'mt-4')}>{cancelMutation.error.message}</p>}
              <div className="mt-8 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
                <button type="button" onClick={() => setCancelTarget(null)} className={BTN_LINE}>
                  Volver
                </button>
                <button type="button" onClick={() => cancelMutation.mutate(cancelTarget.id)} disabled={cancelMutation.isPending}
                  className={BTN_PRIMARY}
                >
                  {cancelMutation.isPending ? 'Cancelando...' : 'Sí, cancelar todo'}
                </button>
              </div>
            </DialogPanel>
          </ModalPortal>
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
    <div className="mt-6 flex flex-col gap-5 border-l border-paper/15 pl-5">
      <input aria-label="Nombre" placeholder="Nombre" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className={FIELD} />
      {showNew && (
        <select aria-label="Tipo de espacio" value={form.type} onChange={e => setForm(f => ({ ...f, type: e.target.value }))} className={FIELD}>
          <option value="cancha_padel">Pádel</option>
          <option value="cancha_futbol">Fútbol</option>
          <option value="cancha_padbol">Padbol</option>
          <option value="cancha_beach_voley">Beach vóley</option>
          <option value="quincho">Quincho / Salón</option>
        </select>
      )}
      <input aria-label="Descripción (opcional)" placeholder="Descripción (opcional)" value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} className={FIELD} />
      <input aria-label="Precio por turno" type="number" placeholder="Precio por turno" value={form.price_per_slot} onChange={e => setForm(f => ({ ...f, price_per_slot: e.target.value }))} className={FIELD} />
      <div className="mt-2 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
        <button type="button" onClick={onCancel} className={BTN_LINE}>Cancelar</button>
        <button type="button" onClick={onSubmit} disabled={isPending || !form.name || !form.price_per_slot} className={BTN_PRIMARY}>{isPending ? 'Guardando...' : submitLabel}</button>
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
      <div className="mb-8 flex items-center justify-end">
        {!showNew && <button type="button" onClick={() => { setShowNew(true); setEditing(null); }} className={BTN_PRIMARY}>Nuevo espacio</button>}
      </div>
      {showNew && <div className="mb-10"><SpaceForm form={form} setForm={setForm} showNew={showNew} onCancel={() => { setShowNew(false); setEditing(null); }} onSubmit={() => createMutation.mutate()} isPending={createMutation.isPending} submitLabel="Crear espacio" /></div>}
      {isLoading ? (
        <div aria-busy="true" aria-label="Cargando espacios" className="flex flex-col gap-4">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-16 animate-pulse bg-graphite motion-reduce:animate-none" />)}</div>
      ) : (
        <ul>
          {spaces?.map(space => {
            return (
              <li key={space.id} className="border-t border-paper/10 py-5 last:border-b">
                <div className="flex items-center justify-between gap-4">
                  <div className="min-w-0">
                    <p className="truncate font-arch text-base font-semibold text-paper">{space.name}</p>
                    <p className={cn(MONO, 'mt-1 text-concrete')}>{SPACE_LABELS[space.type]} · <span className="text-paper">${space.price_per_slot.toLocaleString('es-AR')}</span> / turno</p>
                  </div>
                  <div className="flex shrink-0 gap-1">
                    <button type="button" aria-label={`Editar ${space.name}`} onClick={() => { setEditing(space.id); setShowNew(false); setForm({ name: space.name, type: space.type, description: space.description ?? '', price_per_slot: String(space.price_per_slot) }); }} className="p-2.5 text-concrete outline-none transition-colors hover:text-paper focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper"><Pencil size={16} /></button>
                    <button type="button" aria-label={`Eliminar ${space.name}`} onClick={() => deleteMutation.mutate(space.id)} className="p-2.5 text-concrete outline-none transition-colors hover:text-red-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper"><Trash2 size={16} /></button>
                  </div>
                </div>
                {editing === space.id && <SpaceForm form={form} setForm={setForm} showNew={false} onCancel={() => setEditing(null)} onSubmit={() => updateMutation.mutate(space.id)} isPending={updateMutation.isPending} submitLabel="Guardar cambios" />}
              </li>
            );
          })}
        </ul>
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

  return (
    <div>
      <div className="mb-8 flex items-center justify-end">
        {!showNew && <button type="button" onClick={() => setShowNew(true)} className={BTN_PRIMARY}>Nuevo usuario</button>}
      </div>

      {showNew && (
        <div className="mb-10 flex flex-col gap-5 border-l border-paper/15 pl-5">
          <input aria-label="Nombre" placeholder="Nombre" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} className={FIELD} />
          <input aria-label="Teléfono" type="tel" placeholder="Teléfono" value={form.phone} onChange={e => setForm(f => ({ ...f, phone: e.target.value }))} className={FIELD} />
          <input aria-label="Contraseña" type="password" placeholder="Contraseña" value={form.password} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} className={FIELD} />
          <select aria-label="Rol" value={form.role} onChange={e => setForm(f => ({ ...f, role: e.target.value as any }))} className={FIELD}>
            <option value="receptionist">Recepcionista</option>
            <option value="admin">Administrador</option>
          </select>
          {createMutation.error && <p role="alert" className={ERROR_TEXT}>{createMutation.error.message}</p>}
          <div className="mt-2 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
            <button type="button" onClick={() => setShowNew(false)} className={BTN_LINE}>Cancelar</button>
            <button type="button" onClick={() => createMutation.mutate()} disabled={createMutation.isPending || !form.name || !form.phone || !form.password} className={BTN_PRIMARY}>{createMutation.isPending ? 'Creando...' : 'Crear usuario'}</button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div aria-busy="true" aria-label="Cargando usuarios" className="flex flex-col gap-4">{Array.from({ length: 3 }).map((_, i) => <div key={i} className="h-14 animate-pulse bg-graphite motion-reduce:animate-none" />)}</div>
      ) : (
        <ul>
          {users?.map(user => (
            <li key={user.id} className="flex items-center justify-between gap-4 border-t border-paper/10 py-5 last:border-b">
              <div className="min-w-0">
                <p className="truncate font-arch text-base font-semibold text-paper">{user.name}</p>
                <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1">
                  <p className={cn(MONO, 'text-concrete')}>{user.phone}</p>
                  <select
                    aria-label={`Rol de ${user.name}`}
                    value={user.role}
                    onChange={(e) => roleMutation.mutate({ id: user.id, role: e.target.value })}
                    disabled={roleMutation.isPending}
                    className={cn(
                      MONO,
                      'cursor-pointer border-0 border-b border-paper/25 bg-transparent py-0.5 outline-none transition-colors [color-scheme:dark] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper disabled:opacity-50 [&_option]:bg-night [&_option]:text-paper',
                      user.role === 'admin' ? 'text-light' : 'text-paper',
                    )}
                  >
                    <option value="receptionist">Recepcionista</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
              </div>
              <button type="button" aria-label={`Eliminar ${user.name}`} onClick={() => deleteMutation.mutate(user.id)} className="shrink-0 p-2.5 text-concrete outline-none transition-colors hover:text-red-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper"><Trash2 size={16} /></button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────
export default function StaffPanelPage() {
  const queryClient = useQueryClient();
  const reduce = useReducedMotion();
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

  const y = reduce ? 0 : 12;
  const tabTitle = tab === 'bookings' ? 'Reservas del día' : tab === 'fixed' ? 'Turnos fijos y bloqueos' : tab === 'spaces' ? 'Canchas' : 'Usuarios';
  const iconButtonClass =
    'inline-flex size-10 shrink-0 items-center justify-center border border-paper/15 text-paper outline-none transition-colors hover:border-paper/40 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper';

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
            <SectionLabel number="H1" label="Panel" />

            <h1 className="mt-10 break-words font-arch font-expanded text-[clamp(2.25rem,min(9vw,14vh),6rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper md:mt-14">
              <motion.span
                className="block"
                initial={{ opacity: 0, y: reduce ? 0 : 28 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, ease: EASE, delay: 0.15 }}
              >
                Panel
              </motion.span>
            </h1>

            <motion.p
              className={cn(MONO, 'mt-6 text-concrete')}
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, ease: EASE, delay: 0.3 }}
            >
              <span className="text-paper">{isAdmin ? 'Panel de administración' : 'Panel de recepción'}</span>
              {tab === 'bookings' && (
                <>
                  <span> · </span>
                  <span className="capitalize">{formatDate(date)}</span>
                </>
              )}
            </motion.p>

            <motion.div
              className="mt-14 md:mt-20"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, ease: EASE, delay: 0.4 }}
            >
              {/* Tabs — reservas + turnos fijos para todo el staff; canchas + usuarios solo admin */}
              <div role="group" aria-label="Secciones del panel" className="flex flex-wrap gap-x-8 gap-y-4">
                {([
                  { key: 'bookings', label: 'Reservas' },
                  { key: 'fixed',    label: 'Turnos fijos' },
                  ...(isAdmin ? [
                    { key: 'spaces', label: 'Canchas' },
                    { key: 'users',  label: 'Usuarios' },
                  ] : []),
                ] as { key: Tab; label: string }[]).map(({ key, label }) => {
                  const selected = tab === key;
                  return (
                    <button
                      key={key}
                      type="button"
                      aria-pressed={selected}
                      onClick={() => setTab(key)}
                      className={cn(
                        MONO,
                        'relative pb-3 outline-none transition-colors duration-300 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper',
                        selected ? 'text-paper' : 'text-concrete hover:text-paper',
                      )}
                    >
                      {label}
                      {selected && (
                        <motion.span
                          layoutId="panel-tab-led"
                          aria-hidden="true"
                          className="absolute inset-x-0 bottom-0 h-[2px] bg-light"
                          transition={{ duration: reduce ? 0 : 0.35, ease: EASE }}
                        />
                      )}
                    </button>
                  );
                })}
              </div>

              <div className="mt-12 flex flex-wrap items-end justify-between gap-x-8 gap-y-6 border-t border-paper/10 pt-8">
                <h2 className={SECTION_TITLE}>{tabTitle}</h2>
                {tab === 'bookings' && (
                  <button
                    type="button"
                    onClick={() => setShowManual(true)}
                    className={BTN_PRIMARY}
                  >
                    Nueva reserva
                  </button>
                )}
              </div>

              {/* Date nav — solo en tab reservas */}
              {tab === 'bookings' && <div className="mt-8 flex flex-wrap items-center gap-3">
                <button
                  type="button"
                  aria-label="Día anterior"
                  onClick={() => setDate((d) => shiftDate(d, -1))}
                  className={iconButtonClass}
                >
                  <ArrowLeft size={16} />
                </button>

                <input
                  type="date"
                  aria-label="Fecha"
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                  className={cn(MONO, 'h-10 cursor-pointer border-0 border-b border-paper/25 bg-transparent px-1 text-paper outline-none transition-colors [color-scheme:dark] focus:border-paper focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper')}
                />

                <button
                  type="button"
                  aria-label="Día siguiente"
                  onClick={() => setDate((d) => shiftDate(d, 1))}
                  className={iconButtonClass}
                >
                  <ArrowRight size={16} />
                </button>

                {date !== today && (
                  <button
                    type="button"
                    onClick={() => setDate(today)}
                    className={cn(BTN_LINE, 'ml-2')}
                  >
                    Hoy
                  </button>
                )}
              </div>}

              {/* Content */}
              <div className="mt-10">

                {/* Admin tabs content */}
                {tab === 'spaces' && <SpacesPanel />}
                {tab === 'users'  && <UsersPanel />}
                {tab === 'fixed'  && <FixedSlotsPanel defaultDate={date} />}

                {/* Bookings content */}
                {tab === 'bookings' && <>

                {/* Summary line */}
                {!isLoading && bookings && bookings.length > 0 && (
                  <p className={cn(MONO, 'mb-8 flex flex-wrap gap-x-6 gap-y-1 text-concrete')}>
                    <span><span className="text-paper">{String(bookings.length).padStart(2, '0')}</span> total</span>
                    {active.length > 0 && (
                      <span><span className="text-paper">{String(active.length).padStart(2, '0')}</span> activas</span>
                    )}
                    {past.filter(b => b.status === 'cancelled').length > 0 && (
                      <span><span className="text-paper">{String(past.filter(b => b.status === 'cancelled').length).padStart(2, '0')}</span> canceladas</span>
                    )}
                  </p>
                )}

                {isLoading ? (
                  <div aria-busy="true" aria-label="Cargando reservas" className="flex flex-col gap-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="h-16 animate-pulse bg-graphite motion-reduce:animate-none" />
                    ))}
                  </div>
                ) : !bookings?.length ? (
                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    className="max-w-md py-10"
                  >
                    <p className="font-arch font-expanded text-xl font-bold uppercase leading-tight tracking-[-0.02em] text-paper md:text-2xl">Sin reservas</p>
                    <p className="mt-4 text-base leading-relaxed text-paper/85">
                      No hay reservas para el{' '}
                      <span className="capitalize">{formatDate(date)}</span>.
                    </p>
                  </motion.div>
                ) : (
                  <div className="flex flex-col gap-16">
                    {active.length > 0 && (
                      <section aria-labelledby="staff-active-title">
                        <h2 id="staff-active-title" className={cn(MONO, 'mb-4 text-paper')}>
                          Activas ({active.length})
                        </h2>
                        <BookingsTable>
                          <AnimatePresence mode="popLayout">
                            <tbody>
                              {active.map((b) => (
                                <BookingRow key={b.id} booking={b} onCancel={setCancelTarget} />
                              ))}
                            </tbody>
                          </AnimatePresence>
                        </BookingsTable>
                      </section>
                    )}
                    {past.length > 0 && (
                      <section aria-labelledby="staff-past-title">
                        <h2 id="staff-past-title" className={cn(MONO, 'mb-4 text-concrete')}>
                          Historial ({past.length})
                        </h2>
                        <BookingsTable>
                          <tbody>
                            {past.map((b) => (
                              <BookingRow key={b.id} booking={b} onCancel={setCancelTarget} />
                            ))}
                          </tbody>
                        </BookingsTable>
                      </section>
                    )}
                  </div>
                )}
                </>}
              </div>
            </motion.div>
          </div>
        </main>
      </div>

      {/* Cancel dialog */}
      <AnimatePresence>
        {cancelTarget && (
          <ModalPortal>
            <DialogPanel onClose={() => setCancelTarget(null)} labelledBy="cancel-booking-title">
              <h3 id="cancel-booking-title" className={DIALOG_TITLE}>¿Cancelar reserva?</h3>
              <p className={cn(MONO, 'mt-6 text-concrete')}>
                <span className="text-paper">{cancelTarget.space.name}</span>
                {' · '}
                {cancelTarget.customer.name}
                {' · '}
                <span className="capitalize">{formatDate(cancelTarget.booking_date)}</span>
              </p>
              {cancelMutation.error && (
                <p role="alert" className={cn(ERROR_TEXT, 'mt-4')}>{cancelMutation.error.message}</p>
              )}
              <div className="mt-8 flex flex-col-reverse gap-4 sm:flex-row sm:items-center sm:justify-between">
                <button
                  type="button"
                  onClick={() => setCancelTarget(null)}
                  className={BTN_LINE}
                >
                  Volver
                </button>
                <button
                  type="button"
                  onClick={() => cancelMutation.mutate(cancelTarget.id)}
                  disabled={cancelMutation.isPending}
                  className={BTN_PRIMARY}
                >
                  {cancelMutation.isPending ? 'Cancelando...' : 'Sí, cancelar'}
                </button>
              </div>
            </DialogPanel>
          </ModalPortal>
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
