import { useRef } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { motion, useReducedMotion } from 'framer-motion';
import { getBooking } from '../api/bookings';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/components/ui/button';
import { LandingFooter } from '@/components/landing/LandingFooter';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { formatPrice } from '@/components/landing/spaceHelpers';
import { CONTAINER, EASE, MONO } from '@/components/landing/ui';

const MAX_WAIT_MS = 30_000;
const POLL_INTERVAL_MS = 3000;

export default function PaymentResultPage() {
  const { bookingId } = useParams<{ bookingId: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const reduce = useReducedMotion();
  const mpStatus = searchParams.get('status');
  const startedAtRef = useRef(Date.now());

  const { data: booking } = useQuery({
    queryKey: ['booking', bookingId],
    queryFn: () => getBooking(Number(bookingId)),
    enabled: !!bookingId,
    refetchInterval: (query) => {
      const paid = query.state.data?.deposit_status === 'paid';
      const expired = Date.now() - startedAtRef.current > MAX_WAIT_MS;
      if (paid || mpStatus === 'rejected' || expired) return false;
      return POLL_INTERVAL_MS;
    },
  });

  const isPaid = booking?.deposit_status === 'paid';
  const isRejected = mpStatus === 'rejected' && !isPaid;
  const timedOut = !isPaid && !isRejected && Date.now() - startedAtRef.current > MAX_WAIT_MS;

  const title = isPaid ? 'Pago aprobado' : isRejected ? 'Pago rechazado' : 'Pago pendiente';
  const heading = isPaid
    ? '¡Pago confirmado!'
    : isRejected
    ? 'El pago no se pudo procesar'
    : 'Confirmando tu pago...';
  const body = isPaid
    ? 'Tu seña quedó registrada y la reserva está confirmada.'
    : isRejected
    ? 'Podés reintentar el pago desde "Mis reservas".'
    : timedOut
    ? 'Está tardando más de lo normal. Revisá el estado en "Mis reservas" en unos minutos.'
    : 'Mercado Pago nos avisa apenas se acredite. Puede tardar unos segundos.';

  const y = reduce ? 0 : 12;
  const row = 'flex items-baseline justify-between gap-6 border-t border-paper/10 py-3';

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
            <SectionLabel number="H1" label="Pago" />

            <h1 className="mt-10 break-words font-arch font-expanded text-[clamp(2.25rem,min(8vw,13vh),6.5rem)] font-extrabold uppercase leading-[0.92] tracking-[-0.02em] text-paper md:mt-14">
              <motion.span
                className="block"
                initial={{ opacity: 0, y: reduce ? 0 : 28 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, ease: EASE, delay: 0.15 }}
              >
                {title}
              </motion.span>
            </h1>

            <motion.div
              className="mt-8 max-w-md"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: EASE, delay: 0.35 }}
              role="status"
              aria-live="polite"
            >
              <p className="flex items-center gap-3 font-arch font-expanded text-lg font-bold uppercase leading-tight tracking-[-0.01em] text-paper md:text-xl">
                {!isPaid && !isRejected && !timedOut && (
                  <span aria-hidden="true" className="size-2 shrink-0 animate-pulse bg-light motion-reduce:animate-none" />
                )}
                {heading}
              </p>
              <p className="mt-4 text-base leading-relaxed text-paper/85 md:text-lg">{body}</p>

              {booking && (
                <dl className={cn(MONO, 'mt-10 flex flex-col')}>
                  <div className={row}>
                    <dt className="text-concrete">Espacio</dt>
                    <dd className="text-right text-paper">{booking.space.name}</dd>
                  </div>
                  <div className={row}>
                    <dt className="text-concrete">Turno</dt>
                    <dd className="text-right text-paper">{booking.slot.label}</dd>
                  </div>
                  <div className={cn(row, 'border-y')}>
                    <dt className="text-concrete">Seña</dt>
                    <dd className="text-right text-paper">${formatPrice(booking.deposit_amount)}</dd>
                  </div>
                </dl>
              )}

              <div className="mt-10 flex flex-col gap-6 sm:flex-row sm:items-center sm:gap-10">
                <button
                  type="button"
                  onClick={() => navigate('/mis-reservas')}
                  className={cn(buttonVariants({ variant: 'court' }), 'h-14 justify-center px-8 text-base font-semibold')}
                >
                  Ver mis reservas
                </button>
                <button
                  type="button"
                  onClick={() => navigate('/')}
                  className={cn(buttonVariants({ variant: 'line' }), MONO, 'h-auto justify-center py-2 text-paper')}
                >
                  Volver al inicio
                </button>
              </div>

              {!isPaid && !isRejected && !timedOut && (
                <p className={cn(MONO, 'mt-6 normal-case tracking-normal text-concrete')}>
                  Esta página se actualiza sola.
                </p>
              )}
            </motion.div>
          </div>
        </main>

        <LandingFooter />
      </div>
    </div>
  );
}
