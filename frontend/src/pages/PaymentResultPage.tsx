import { useRef } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import { motion } from 'framer-motion';
import { getBooking } from '../api/bookings';

const MAX_WAIT_MS = 30_000;
const POLL_INTERVAL_MS = 3000;

export default function PaymentResultPage() {
  const { bookingId } = useParams<{ bookingId: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
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

  return (
    <div className="min-h-[calc(100vh-58px)] bg-bg flex items-center justify-center px-6">
      <motion.div
        initial={{ opacity: 0, scale: 0.96 }}
        animate={{ opacity: 1, scale: 1 }}
        className="max-w-sm w-full bg-white rounded-xl border-[1.5px] border-black/[0.07] shadow-md p-8 text-center"
      >
        {isPaid ? (
          <>
            <div className="w-14 h-14 rounded-full bg-status-confirmed/10 flex items-center justify-center mx-auto mb-4">
              <CheckCircle2 size={28} className="text-status-confirmed" />
            </div>
            <h3 className="font-serif text-2xl text-ink mb-1">¡Pago confirmado!</h3>
            <p className="text-sm text-ink-2 mb-6">Tu seña quedó registrada y la reserva está confirmada.</p>
          </>
        ) : isRejected ? (
          <>
            <div className="w-14 h-14 rounded-full bg-status-cancelled/10 flex items-center justify-center mx-auto mb-4">
              <XCircle size={28} className="text-status-cancelled" />
            </div>
            <h3 className="font-serif text-2xl text-ink mb-1">El pago no se pudo procesar</h3>
            <p className="text-sm text-ink-2 mb-6">Podés reintentar el pago desde "Mis reservas".</p>
          </>
        ) : (
          <>
            <div className="w-14 h-14 rounded-full bg-status-pending/10 flex items-center justify-center mx-auto mb-4">
              <Loader2 size={28} className="text-status-pending animate-spin" />
            </div>
            <h3 className="font-serif text-2xl text-ink mb-1">Confirmando tu pago...</h3>
            <p className="text-sm text-ink-2 mb-6">
              {timedOut
                ? 'Está tardando más de lo normal. Revisá el estado en "Mis reservas" en unos minutos.'
                : 'Mercado Pago nos avisa apenas se acredite. Puede tardar unos segundos.'}
            </p>
          </>
        )}

        <button
          onClick={() => navigate('/mis-reservas')}
          className="w-full bg-primary text-white font-bold rounded-lg px-6 py-3 text-sm transition-all hover:bg-primary-dark hover:shadow-glow active:scale-[0.98]"
        >
          Ver mis reservas
        </button>
        {!isPaid && !isRejected && !timedOut && (
          <p className="text-2xs text-ink-2/60 mt-3">Esta página se actualiza sola.</p>
        )}
      </motion.div>
    </div>
  );
}
