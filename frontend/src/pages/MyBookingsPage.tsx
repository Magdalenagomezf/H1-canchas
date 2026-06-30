import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Skeleton from '@mui/material/Skeleton';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Divider from '@mui/material/Divider';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import EventBusyIcon from '@mui/icons-material/EventBusy';
import { motion } from 'framer-motion';
import { getMyBookings, cancelBooking } from '../api/bookings';
import type { BookingStatus } from '../types';

const MotionCard = motion(Card);

const STATUS_CONFIG: Record<BookingStatus, { label: string; color: 'warning' | 'success' | 'default' | 'info' }> = {
  pending: { label: 'Pendiente', color: 'warning' },
  confirmed: { label: 'Confirmada', color: 'success' },
  cancelled: { label: 'Cancelada', color: 'default' },
  completed: { label: 'Completada', color: 'info' },
};

const CANCELLABLE: BookingStatus[] = ['pending', 'confirmed'];

function formatDate(dateStr: string): string {
  const [year, month, day] = dateStr.split('-').map(Number);
  return new Date(year, month - 1, day).toLocaleDateString('es-AR', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export default function MyBookingsPage() {
  const queryClient = useQueryClient();
  const [confirmingId, setConfirmingId] = useState<number | null>(null);
  const [successOpen, setSuccessOpen] = useState(false);

  const { data: myBookings, isLoading, isError } = useQuery({
    queryKey: ['my-bookings'],
    queryFn: getMyBookings,
  });

  const cancelMutation = useMutation({
    mutationFn: cancelBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-bookings'] });
      setConfirmingId(null);
      setSuccessOpen(true);
    },
  });

  if (isLoading) return <MyBookingsSkeleton />;

  if (isError) {
    return (
      <Container maxWidth="md" sx={{ py: 8 }}>
        <Alert severity="error">No pudimos cargar tus reservas. Intentá de nuevo más tarde.</Alert>
      </Container>
    );
  }

  return (
    <Box>
      <Box
        sx={{
          background: 'linear-gradient(160deg, #2D5A3D 0%, #3D7A4E 60%, #5A9E6A 100%)',
          py: { xs: 6, md: 8 },
        }}
      >
        <Container maxWidth="lg">
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>
            <Typography variant="h2" sx={{ color: '#fff', fontSize: { xs: '2rem', md: '3rem' }, mb: 1 }}>
              Mis reservas
            </Typography>
            <Typography sx={{ color: 'rgba(255,255,255,0.8)' }}>
              Revisá y gestioná tus turnos reservados.
            </Typography>
          </motion.div>
        </Container>
      </Box>

      <Container maxWidth="md" sx={{ py: { xs: 5, md: 8 } }}>
        {myBookings?.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 10 }}>
            <EventBusyIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 2 }} />
            <Typography variant="h5" color="text.secondary" sx={{ mb: 1 }}>
              Todavía no tenés reservas
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Cuando reserves una cancha, la vas a ver acá.
            </Typography>
          </Box>
        ) : (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
            {myBookings?.map((booking, i) => {
              const status = STATUS_CONFIG[booking.status];
              const canCancel = CANCELLABLE.includes(booking.status);
              const isConfirming = confirmingId === booking.id;
              const isCancelling = cancelMutation.isPending && confirmingId === booking.id;

              return (
                <MotionCard
                  key={booking.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.06 }}
                >
                  <CardContent sx={{ p: { xs: 2.5, sm: 3 } }}>
                    <Box
                      sx={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'flex-start',
                        mb: 2,
                        gap: 2,
                        flexWrap: 'wrap',
                      }}
                    >
                      <Typography variant="h6">{booking.space.name}</Typography>
                      <Chip label={status.label} color={status.color} size="small" sx={{ fontWeight: 600 }} />
                    </Box>

                    <Divider sx={{ mb: 2 }} />

                    <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', mb: 2 }}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                        <CalendarMonthIcon fontSize="small" color="action" />
                        <Typography variant="body2" color="text.secondary">
                          {formatDate(booking.booking_date)}
                        </Typography>
                      </Box>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                        <AccessTimeIcon fontSize="small" color="action" />
                        <Typography variant="body2" color="text.secondary">
                          {booking.slot.label}
                          {booking.slot.start_time && booking.slot.end_time
                            ? ` (${booking.slot.start_time} – ${booking.slot.end_time})`
                            : ''}
                        </Typography>
                      </Box>
                    </Box>

                    <Box
                      sx={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        flexWrap: 'wrap',
                        gap: 2,
                      }}
                    >
                      <Typography variant="h6" color="primary.main">
                        ${booking.total_price.toLocaleString('es-AR')}
                      </Typography>

                      {canCancel &&
                        (isConfirming ? (
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Typography variant="body2" color="text.secondary">
                              ¿Cancelar este turno?
                            </Typography>
                            <Button
                              size="small"
                              color="error"
                              variant="contained"
                              disabled={isCancelling}
                              onClick={() => cancelMutation.mutate(booking.id)}
                            >
                              {isCancelling ? 'Cancelando...' : 'Sí, cancelar'}
                            </Button>
                            <Button size="small" onClick={() => setConfirmingId(null)}>
                              No
                            </Button>
                          </Box>
                        ) : (
                          <Button
                            size="small"
                            variant="outlined"
                            color="error"
                            onClick={() => setConfirmingId(booking.id)}
                          >
                            Cancelar turno
                          </Button>
                        ))}
                    </Box>
                  </CardContent>
                </MotionCard>
              );
            })}
          </Box>
        )}
      </Container>

      <Snackbar
        open={successOpen}
        autoHideDuration={3000}
        onClose={() => setSuccessOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity="success" onClose={() => setSuccessOpen(false)}>
          Reserva cancelada correctamente.
        </Alert>
      </Snackbar>
    </Box>
  );
}

function MyBookingsSkeleton() {
  return (
    <Box>
      <Box
        sx={{
          background: 'linear-gradient(160deg, #2D5A3D 0%, #3D7A4E 60%, #5A9E6A 100%)',
          py: { xs: 6, md: 8 },
        }}
      >
        <Container maxWidth="lg">
          <Skeleton variant="text" width={240} height={56} sx={{ bgcolor: 'rgba(255,255,255,0.15)' }} />
          <Skeleton variant="text" width={300} height={28} sx={{ bgcolor: 'rgba(255,255,255,0.1)' }} />
        </Container>
      </Box>
      <Container maxWidth="md" sx={{ py: { xs: 5, md: 8 } }}>
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i} sx={{ mb: 3 }}>
            <CardContent sx={{ p: 3 }}>
              <Skeleton height={32} width="50%" sx={{ mb: 2 }} />
              <Skeleton height={1} sx={{ mb: 2 }} />
              <Skeleton height={20} width="70%" sx={{ mb: 1 }} />
              <Skeleton height={20} width="50%" sx={{ mb: 2 }} />
              <Skeleton height={32} width="30%" />
            </CardContent>
          </Card>
        ))}
      </Container>
    </Box>
  );
}
