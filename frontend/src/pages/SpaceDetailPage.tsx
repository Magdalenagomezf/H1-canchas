import { useState, useEffect, useRef, type ReactNode } from 'react';
import { useParams, useNavigate, useLocation, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion, AnimatePresence, useInView, useReducedMotion } from 'framer-motion';
import { ArrowRight, Clock } from 'lucide-react';
import { getSpace, getSlots } from '../api/spaces';
import { createBooking } from '../api/bookings';
import { generatePreference } from '../api/payments';
import { SPACE_IMAGES } from '../components/SpaceCard';
import { ModalPortal } from '../components/ModalPortal';
import { useAuth } from '../hooks/useAuth';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/components/ui/button';
import { LandingFooter } from '@/components/landing/LandingFooter';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { SPACE_TYPE_LABEL } from '@/components/landing/content';
import { formatPrice } from '@/components/landing/spaceHelpers';
import { CONTAINER, EASE, GRID, MONO, VIEWPORT } from '@/components/landing/ui';
import type { BookingDetail } from '../types';

const TZ = 'America/Argentina/Buenos_Aires';

function toArgISO(d: Date): string {
  return d.toLocaleDateString('en-CA', { timeZone: TZ });
}

function getNextDays(n: number) {
  const now = new Date();
  return Array.from({ length: n }, (_, i) => {
    const d = new Date(now);
    d.setDate(d.getDate() + i);
    return d;
  });
}

const DAYS = getNextDays(14);
const today = toArgISO(DAYS[0]);

function getArgNowMinutes(): number {
  const [h, m] = new Date().toLocaleTimeString('en-GB', { timeZone: TZ, hour12: false }).split(':');
  return Number(h) * 60 + Number(m);
}

function slotStartMinutes(time: string): number {
  // start_time viene como "HH:MM:SS" o, si la API serializa un TIME sin zona
  // como datetime completo, como "0000-01-01THH:MM:SSZ" — en ambos casos
  // la hora es local (Argentina), la "Z" no implica una conversión real a UTC.
  const clock = time.includes('T') ? time.split('T')[1] : time;
  const [h, m] = clock.split(':');
  return Number(h) * 60 + Number(m);
}

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m} min`;
  if (m === 0) return h === 1 ? '1 hora' : `${h} horas`;
  return `${h} h ${m} min`;
}

export default function SpaceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated } = useAuth();

  const locationState = location.state as { pendingBooking?: { date: string; slotId: number } } | null;
  const [date, setDate] = useState(locationState?.pendingBooking?.date ?? today);
  const [selectedSlot, setSelectedSlot] = useState<number | null>(locationState?.pendingBooking?.slotId ?? null);
  const [imgIndex, setImgIndex] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const [nowMinutes, setNowMinutes] = useState(() => getArgNowMinutes());
  const [confirmedBooking, setConfirmedBooking] = useState<BookingDetail | null>(null);

  const queryClient = useQueryClient();
  const reduce = useReducedMotion();

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
    onSuccess: (booking) => {
      queryClient.invalidateQueries({ queryKey: ['slots', id, date] });
      setConfirmedBooking(booking);
    },
  });

  const handleConfirm = () => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: location, pendingBooking: { date, slotId: selectedSlot } } });
      return;
    }
    bookingMutation.mutate();
  };

  const images = space ? (SPACE_IMAGES[space.type] ?? []) : [];

  useEffect(() => {
    if (images.length <= 1) return;
    intervalRef.current = setInterval(() => {
      setImgIndex((prev) => (prev + 1) % images.length);
    }, 3000);
    return () => { if (intervalRef.current) clearInterval(intervalRef.current); };
  }, [images.length]);

  useEffect(() => {
    const id = setInterval(() => setNowMinutes(getArgNowMinutes()), 30_000);
    return () => clearInterval(id);
  }, []);

  const handleDotClick = (i: number) => {
    setImgIndex(i);
    if (intervalRef.current) clearInterval(intervalRef.current);
    intervalRef.current = setInterval(() => {
      setImgIndex((prev) => (prev + 1) % images.length);
    }, 3000);
  };

  const y = reduce ? 0 : 12;

  if (spaceLoading) {
    return (
      <SpaceShell>
        <div className={cn(CONTAINER, 'relative pb-20 pt-10 md:pt-14')} aria-busy="true" aria-label="Cargando espacio">
          <div className="h-4 w-40 animate-pulse bg-graphite motion-reduce:animate-none" />
          <div className={cn(GRID, 'mt-10 gap-y-10 lg:items-end')}>
            <div className="col-span-12 flex flex-col gap-5 lg:col-span-5">
              <div className="h-6 w-40 animate-pulse bg-graphite motion-reduce:animate-none" />
              <div className="h-16 w-3/4 animate-pulse bg-graphite motion-reduce:animate-none" />
              <div className="h-4 w-2/3 animate-pulse bg-graphite motion-reduce:animate-none" />
              <div className="h-4 w-32 animate-pulse bg-graphite motion-reduce:animate-none" />
            </div>
            <div className="col-span-12 lg:col-span-7">
              <div className="aspect-[4/3] w-full animate-pulse bg-graphite motion-reduce:animate-none" />
            </div>
          </div>
        </div>
      </SpaceShell>
    );
  }

  if (!space) {
    return (
      <SpaceShell>
        <div className={cn(CONTAINER, 'relative pb-28 pt-10 md:pt-14')}>
          <BackLink />
          <p className="mt-14 max-w-md text-base leading-relaxed text-paper/85 md:text-lg">
            No encontramos este espacio. Puede que ya no esté disponible.
          </p>
        </div>
      </SpaceShell>
    );
  }

  const selectedLabel = selectedSlot && slots ? slots.find((s) => s.id === selectedSlot)?.label : undefined;
  const dateLabel = new Date(date + 'T12:00:00').toLocaleDateString('es-AR', {
    weekday: 'long', day: 'numeric', month: 'long',
  });
  const ctaLabel = bookingMutation.isPending
    ? 'Confirmando...'
    : !isAuthenticated
    ? 'Iniciar sesión para reservar'
    : 'Reservar turno';
  const ctaDisabled = !selectedSlot || bookingMutation.isPending;
  const ctaClass = cn(
    buttonVariants({ variant: 'court' }),
    'h-14 w-full justify-center gap-3 text-base font-semibold disabled:cursor-not-allowed disabled:opacity-40',
  );

  return (
    <SpaceShell padBottom>
      <div className={cn(CONTAINER, 'relative pb-14 pt-10 md:pb-20 md:pt-14')}>
        <motion.div
          initial={{ opacity: 0, y }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: EASE, delay: 0.1 }}
        >
          <BackLink />
        </motion.div>

        <div className={cn(GRID, 'mt-10 gap-y-10 md:mt-14 lg:items-end')}>
          <div className="col-span-12 lg:col-span-5">
            <SectionLabel number="H1" label={SPACE_TYPE_LABEL[space.type]} />

            <h1 className="mt-8 break-words font-arch font-expanded text-[clamp(2.25rem,min(5vw,10vh),5rem)] font-extrabold uppercase leading-[0.92] tracking-[-0.02em] text-paper md:mt-10">
              <motion.span
                className="block"
                initial={{ opacity: 0, y: reduce ? 0 : 28 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, ease: EASE, delay: 0.15 }}
              >
                {space.name}
              </motion.span>
            </h1>

            <motion.div
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: EASE, delay: 0.35 }}
            >
              {space.description && (
                <p className="mt-6 max-w-md text-base leading-relaxed text-paper/85 md:text-lg">{space.description}</p>
              )}
              <p className={cn(MONO, 'mt-8 text-concrete')}>
                <span className="text-lg text-paper md:text-xl">${formatPrice(space.price_per_slot)}</span> / turno
              </p>
            </motion.div>
          </div>

          <div className="col-span-12 lg:col-span-7">
            <SpaceGallery images={images} alt={space.name} index={imgIndex} onSelect={handleDotClick} />
          </div>
        </div>
      </div>

      {/* ── Booking ── */}
      <section aria-label="Reservar turno" className="border-t border-paper/10">
        <div className={cn(CONTAINER, GRID, 'py-14 md:py-20 lg:py-24')}>
          <div className="col-span-12 lg:col-span-7">
            {/* Date */}
            <SectionLabel number="01" label="Fecha" />
            <div
              role="group"
              aria-label="Seleccioná el día"
              className="-mx-6 mt-8 flex gap-1 overflow-x-auto px-6 pb-2 md:mx-0 md:grid md:grid-cols-7 md:gap-x-2 md:gap-y-4 md:overflow-visible md:px-0"
            >
              {DAYS.map((d) => {
                const iso = toArgISO(d);
                const isSelected = date === iso;
                const dayName = d.toLocaleDateString('es-AR', { weekday: 'short', timeZone: TZ });
                const dayNum = Number(d.toLocaleDateString('en-CA', { day: 'numeric', timeZone: TZ }));
                const monthName = d.toLocaleDateString('es-AR', { month: 'short', timeZone: TZ });
                return (
                  <button
                    key={iso}
                    type="button"
                    aria-pressed={isSelected}
                    onClick={() => { setDate(iso); setSelectedSlot(null); }}
                    className={cn(
                      MONO,
                      'relative flex w-16 shrink-0 flex-col items-start gap-1 pb-3 pt-2 text-left outline-none transition-colors duration-300 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper motion-reduce:transition-none md:w-auto',
                      isSelected ? 'text-paper' : 'text-concrete hover:text-paper',
                    )}
                  >
                    <span>{dayName.replace('.', '')}</span>
                    <span className="font-jb text-2xl leading-none tracking-normal">{String(dayNum).padStart(2, '0')}</span>
                    <span>{monthName.replace('.', '')}</span>
                    <span aria-hidden="true" className="absolute inset-x-0 bottom-0 h-px bg-paper/10" />
                    {isSelected && (
                      <span
                        aria-hidden="true"
                        className="absolute inset-x-0 bottom-0 h-[2px] bg-light shadow-[0_0_8px_color-mix(in_srgb,var(--color-light)_50%,transparent)]"
                      />
                    )}
                  </button>
                );
              })}
            </div>

            {/* Slots */}
            <SectionLabel number="02" label="Turno" className="mt-16 md:mt-20" />
            <div className="mt-8">
              {slotsLoading ? (
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4" aria-busy="true">
                  {Array.from({ length: 6 }).map((_, i) => (
                    <div key={i} className="h-14 animate-pulse bg-graphite motion-reduce:animate-none" />
                  ))}
                </div>
              ) : !slots?.length ? (
                <div className={cn(MONO, 'flex items-center gap-3 py-6 normal-case tracking-normal text-concrete')}>
                  <Clock size={15} className="shrink-0" aria-hidden="true" />
                  No hay turnos disponibles para este día.
                </div>
              ) : (
                <div role="group" aria-label="Turnos" className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4">
                  {slots.map((slot) => {
                    const isBooked = slot.available === false;
                    const isPast = date === today && slot.start_time !== null
                      && slotStartMinutes(slot.start_time) <= nowMinutes;
                    const isDisabled = isBooked || isPast;
                    const isSelected = selectedSlot === slot.id;
                    return (
                      <button
                        key={slot.id}
                        type="button"
                        disabled={isDisabled}
                        aria-pressed={isDisabled ? undefined : isSelected}
                        onClick={() => !isDisabled && setSelectedSlot(slot.id === selectedSlot ? null : slot.id)}
                        className={cn(
                          'relative flex min-h-14 flex-col items-center justify-center px-3 py-3 text-center font-jb text-sm tracking-[0.08em] outline-none transition-colors duration-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-paper motion-reduce:transition-none',
                          isDisabled
                            ? 'cursor-not-allowed border-b border-paper/10 text-paper/25 line-through'
                            : isSelected
                            ? 'bg-dusk text-white'
                            : 'border-b border-paper/15 bg-graphite text-paper hover:bg-paper/10',
                        )}
                      >
                        <span>{slot.label}</span>
                        {isBooked && (
                          <span className="mt-0.5 text-[0.625rem] uppercase tracking-[0.12em] text-paper/25 no-underline">
                            Reservado
                          </span>
                        )}
                        {isSelected && (
                          <span aria-hidden="true" className="absolute inset-x-0 bottom-0 h-[2px] bg-light" />
                        )}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Summary */}
          <aside className="col-span-12 mt-16 lg:col-span-4 lg:col-start-9 lg:mt-0">
            <div className="lg:sticky lg:top-28">
              <p className={cn(MONO, 'text-concrete')}>Resumen</p>
              <dl className={cn(MONO, 'mt-6 flex flex-col')}>
                <div className="flex items-baseline justify-between gap-6 border-t border-paper/10 py-4">
                  <dt className="text-concrete">Espacio</dt>
                  <dd className="text-right text-paper">{space.name}</dd>
                </div>
                <div className="flex items-baseline justify-between gap-6 border-t border-paper/10 py-4">
                  <dt className="text-concrete">Fecha</dt>
                  <dd className="text-right text-paper">{dateLabel}</dd>
                </div>
                <div className="flex items-baseline justify-between gap-6 border-t border-paper/10 py-4">
                  <dt className="text-concrete">Turno</dt>
                  <dd className={cn('text-right', selectedLabel ? 'text-paper' : 'text-concrete')}>
                    {selectedLabel ?? '—'}
                  </dd>
                </div>
                <div className="flex items-baseline justify-between gap-6 border-y border-paper/10 py-4">
                  <dt className="text-concrete">Total</dt>
                  <dd className="text-xl text-paper">${formatPrice(space.price_per_slot)}</dd>
                </div>
              </dl>

              {bookingMutation.error && (
                <p role="alert" className={cn(MONO, 'mt-6 hidden normal-case tracking-normal text-red-300 lg:block')}>
                  {bookingMutation.error.message}
                </p>
              )}

              <button
                type="button"
                onClick={handleConfirm}
                disabled={ctaDisabled}
                className={cn(ctaClass, 'mt-8 hidden lg:inline-flex')}
              >
                {ctaLabel}
                <ArrowRight size={18} className="transition-transform duration-300 group-hover:translate-x-1" />
              </button>

              {!isAuthenticated && (
                <p className={cn(MONO, 'mt-4 normal-case tracking-normal text-concrete')}>
                  Seleccioná un turno y te redirigimos al login.
                </p>
              )}
            </div>
          </aside>
        </div>
      </section>

      {/* Mobile sticky CTA. Fixed but transform-free ancestors keep it viewport-anchored. */}
      <div className="fixed inset-x-0 bottom-0 z-30 border-t border-paper/10 bg-night px-6 py-3 lg:hidden">
        {bookingMutation.error && (
          <p className={cn(MONO, 'mb-2 normal-case tracking-normal text-red-300')}>{bookingMutation.error.message}</p>
        )}
        <div className="flex items-center gap-4">
          <p className={cn(MONO, 'min-w-0 flex-1 truncate text-concrete')}>
            {selectedLabel ? <span className="text-paper">{selectedLabel}</span> : 'Elegí un turno'}
            <br />
            ${formatPrice(space.price_per_slot)}
          </p>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={ctaDisabled}
            className={cn(ctaClass, 'w-auto shrink-0 px-6 text-sm')}
          >
            {ctaLabel}
            <ArrowRight size={16} />
          </button>
        </div>
      </div>

      <AnimatePresence>
        {confirmedBooking && (
          <PaymentSummaryModal
            booking={confirmedBooking}
            onClose={() => setConfirmedBooking(null)}
          />
        )}
      </AnimatePresence>
    </SpaceShell>
  );
}

// ── Layout pieces ─────────────────────────────────────────────────────────────

/** Landing chrome for a non-home route: solid nav, top LED slab, footer. No transforms on ancestors. */
function SpaceShell({ children, padBottom = false }: { children: ReactNode; padBottom?: boolean }) {
  return (
    <div className="landing">
      <div className={cn('on-dark min-h-[100svh] bg-night text-paper', padBottom && 'pb-24 lg:pb-0')}>
        <LandingNav solid />
        <main className="relative pt-16 md:pt-20">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-x-0 top-16 h-[160px] bg-gradient-to-b from-light/10 to-transparent md:top-20 md:h-[200px]"
          />
          <LightLine className="absolute inset-x-0 top-16 md:top-20" delay={0.3} />
          {children}
        </main>
        <LandingFooter />
      </div>
    </div>
  );
}

function BackLink() {
  return (
    <Link to="/canchas" className={cn(MONO, 'inline-block text-concrete transition-colors hover:text-paper')}>
      ← Volver a espacios
    </Link>
  );
}

/** Crossfading gallery inside a clip-path reveal. Rotation state lives in the page. */
function SpaceGallery({
  images,
  alt,
  index,
  onSelect,
}: {
  images: string[];
  alt: string;
  index: number;
  onSelect: (i: number) => void;
}) {
  const reduce = useReducedMotion();
  const frameRef = useRef<HTMLDivElement>(null);
  // Observe the unclipped frame: a fully clipped target never reports as intersecting.
  const inView = useInView(frameRef, VIEWPORT);
  const hidden = reduce ? { opacity: 0 } : { clipPath: 'inset(100% 0 0 0)' };
  const shown = reduce ? { opacity: 1 } : { clipPath: 'inset(0% 0 0 0)' };

  return (
    <div ref={frameRef} className="relative aspect-[4/3] w-full overflow-hidden bg-graphite">
      <motion.div
        className="absolute inset-0"
        initial={hidden}
        animate={inView ? shown : hidden}
        transition={{ duration: 0.7, ease: EASE }}
      >
        {images.map((src, i) => (
          <img
            key={src}
            src={src}
            alt={i === index ? alt : ''}
            aria-hidden={i === index ? undefined : true}
            className={cn(
              'absolute inset-0 h-full w-full object-cover transition-opacity duration-700 motion-reduce:transition-none',
              i === index ? 'opacity-100' : 'opacity-0',
            )}
          />
        ))}
        {images.length > 1 && (
          <div className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-4 bg-gradient-to-t from-night/70 to-transparent px-4 pb-4 pt-10 md:px-6">
            <div className="flex">
              {images.map((_, i) => (
                <button
                  key={i}
                  type="button"
                  onClick={() => onSelect(i)}
                  aria-label={`Ver foto ${i + 1}`}
                  aria-current={i === index}
                  className="group/dot flex h-8 w-10 items-end pb-2 outline-none focus-visible:outline-2 focus-visible:outline-offset-0 focus-visible:outline-paper"
                >
                  <span
                    className={cn(
                      'block h-[2px] w-full transition-colors duration-300',
                      i === index ? 'bg-light' : 'bg-paper/40 group-hover/dot:bg-paper',
                    )}
                  />
                </button>
              ))}
            </div>
            <span aria-hidden="true" className={cn(MONO, 'text-paper')}>
              {String(index + 1).padStart(2, '0')} / {String(images.length).padStart(2, '0')}
            </span>
          </div>
        )}
      </motion.div>
    </div>
  );
}

// ── Payment summary modal ─────────────────────────────────────────────────────
function PaymentSummaryModal({ booking, onClose }: { booking: BookingDetail; onClose: () => void }) {
  const { data: preference, isLoading, isError } = useQuery({
    queryKey: ['preference', booking.id],
    queryFn: () => generatePreference(booking.id, 'deposit'),
  });

  const durationMinutes = booking.slot.start_time && booking.slot.end_time
    ? slotStartMinutes(booking.slot.end_time) - slotStartMinutes(booking.slot.start_time)
    : null;

  const payClass = cn(buttonVariants({ variant: 'court' }), 'h-14 flex-1 justify-center gap-3 text-base font-semibold');

  const row = 'flex items-baseline justify-between gap-6 border-t border-paper/10 py-3';

  return (
    <ModalPortal>
      {/* Portal renders outside .landing, so re-scope the landing tokens (zero-size, no transform). */}
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
            aria-labelledby="payment-modal-title"
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 16 }}
            transition={{ duration: 0.22 }}
            className="pointer-events-auto relative max-h-[90vh] w-full overflow-y-auto border border-paper/10 bg-graphite p-6 text-paper sm:w-[480px] sm:p-8"
          >
            <LightLine className="absolute inset-x-0 top-0" delay={0.1} />

            <h3
              id="payment-modal-title"
              className="font-arch font-expanded text-2xl font-bold uppercase leading-none tracking-[-0.02em] text-paper"
            >
              Confirmá tu reserva
            </h3>
            <p className="mt-4 text-sm leading-relaxed text-paper/80">
              Para reservar el turno necesitás pagar la seña ahora. El resto se abona en el lugar.
            </p>

            <dl className={cn(MONO, 'mt-8 flex flex-col')}>
              <div className={row}>
                <dt className="text-concrete">Espacio</dt>
                <dd className="text-right text-paper">{booking.space.name}</dd>
              </div>
              <div className={row}>
                <dt className="text-concrete">Turno</dt>
                <dd className="text-right text-paper">{booking.slot.label}</dd>
              </div>
              <div className={row}>
                <dt className="text-concrete">Fecha</dt>
                <dd className="text-right text-paper">
                  {new Date(booking.booking_date + 'T12:00:00').toLocaleDateString('es-AR', {
                    weekday: 'long', day: 'numeric', month: 'long',
                  })}
                </dd>
              </div>
              {durationMinutes !== null && (
                <div className={row}>
                  <dt className="text-concrete">Duración</dt>
                  <dd className="text-right text-paper">{formatDuration(durationMinutes)}</dd>
                </div>
              )}
              <div className={row}>
                <dt className="text-concrete">Precio total</dt>
                <dd className="text-paper">${formatPrice(booking.total_price)}</dd>
              </div>
              <div className={cn(row, 'border-y')}>
                <dt className="text-concrete">Seña a pagar ahora</dt>
                <dd className="text-2xl text-paper">${formatPrice(booking.deposit_amount)}</dd>
              </div>
            </dl>

            {isError && (
              <p role="alert" className={cn(MONO, 'mt-6 normal-case tracking-normal text-red-300')}>
                No pudimos generar el link de pago. Cerrá este resumen e intentá de nuevo.
              </p>
            )}

            <div className="mt-8 flex flex-col-reverse gap-4 sm:flex-row sm:items-center">
              <button
                type="button"
                onClick={onClose}
                className={cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto justify-center py-2 text-paper sm:px-2')}
              >
                Cancelar
              </button>
              {preference ? (
                <a href={preference.sandbox_init_point} className={payClass}>
                  Pagar con Mercado Pago
                  <ArrowRight size={18} className="transition-transform duration-300 group-hover:translate-x-1" />
                </a>
              ) : (
                <button type="button" disabled className={cn(payClass, 'cursor-not-allowed opacity-50')}>
                  {isLoading ? 'Preparando el pago...' : 'Cargando...'}
                </button>
              )}
            </div>
          </motion.div>
        </div>
      </div>
    </ModalPortal>
  );
}
