import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Skeleton from '@mui/material/Skeleton';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import TextField from '@mui/material/TextField';
import Divider from '@mui/material/Divider';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import LockIcon from '@mui/icons-material/Lock';
import { motion } from 'framer-motion';
import { getSpace, getSlots } from '../api/spaces';
import { createBooking } from '../api/bookings';
import { useAuth } from '../hooks/useAuth';
import { SPACE_IMAGES, SPACE_LABELS, SPACE_ICONS } from '../components/SpaceCard';
import type { SpaceSlot } from '../types';

const MotionCard = motion(Card);

const today = new Date().toISOString().split('T')[0];

export default function SpaceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  const [date, setDate] = useState('');
  const [selectedSlot, setSelectedSlot] = useState<SpaceSlot | null>(null);
  const [successOpen, setSuccessOpen] = useState(false);
  const [conflictSlotId, setConflictSlotId] = useState<number | null>(null);

  const spaceId = Number(id);

  const { data: space, isLoading: loadingSpace } = useQuery({
    queryKey: ['space', spaceId],
    queryFn: () => getSpace(spaceId),
    enabled: !!spaceId,
  });

  const { data: slots, isLoading: loadingSlots } = useQuery({
    queryKey: ['slots', spaceId],
    queryFn: () => getSlots(spaceId),
    enabled: !!spaceId,
  });

  const bookMutation = useMutation({
    mutationFn: (slotId: number) =>
      createBooking({ space_id: spaceId, slot_id: slotId, booking_date: date }),
    onSuccess: () => {
      setSelectedSlot(null);
      setSuccessOpen(true);
      setTimeout(() => navigate('/mis-reservas'), 1800);
    },
    onError: (err: unknown) => {
      const status = (err as { response?: { status: number } })?.response?.status;
      if (status === 409) {
        setConflictSlotId(selectedSlot?.id ?? null);
      }
    },
  });

  const handleBook = (slot: SpaceSlot) => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    setSelectedSlot(slot);
    setConflictSlotId(null);
    bookMutation.mutate(slot.id);
  };

  if (loadingSpace) return <SpaceDetailSkeleton />;

  if (!space) {
    return (
      <Container maxWidth="md" sx={{ py: 10, textAlign: 'center' }}>
        <Typography variant="h5" color="text.secondary">
          Espacio no encontrado.
        </Typography>
        <Button component={Link} to="/canchas" sx={{ mt: 3 }}>
          Ver todos los espacios
        </Button>
      </Container>
    );
  }

  const Icon = SPACE_ICONS[space.type];

  return (
    <Box>
      {/* Hero image */}
      <Box
        sx={{
          height: { xs: 240, md: 360 },
          backgroundImage: `linear-gradient(to bottom, rgba(0,0,0,0.15), rgba(0,0,0,0.55)),
            url(${SPACE_IMAGES[space.type]})`,
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'flex-end',
          p: { xs: 3, md: 5 },
        }}
      >
        <Button
          component={Link}
          to="/canchas"
          startIcon={<ArrowBackIcon />}
          sx={{ color: '#fff', mb: 2, alignSelf: 'flex-start', opacity: 0.85 }}
        >
          Volver
        </Button>
        <Chip
          icon={<Icon sx={{ color: '#fff !important' }} />}
          label={SPACE_LABELS[space.type]}
          size="small"
          sx={{ bgcolor: 'primary.main', color: '#fff', mb: 1.5, alignSelf: 'flex-start' }}
        />
        <Typography variant="h2" sx={{ color: '#fff', fontSize: { xs: '2rem', md: '3rem' } }}>
          {space.name}
        </Typography>
      </Box>

      <Container maxWidth="lg" sx={{ py: { xs: 4, md: 6 } }}>
        <Grid container spacing={5}>
          {/* Info izquierda */}
          <Grid size={{ xs: 12, md: 4 }}>
            <Box sx={{ position: { md: 'sticky' }, top: 88 }}>
              {space.description && (
                <Typography variant="body1" color="text.secondary" sx={{ mb: 3, lineHeight: 1.7 }}>
                  {space.description}
                </Typography>
              )}
              <Divider sx={{ mb: 3 }} />
              <Box sx={{ display: 'flex', alignItems: 'baseline', gap: 0.5, mb: 1 }}>
                <Typography variant="h4" color="primary.main">
                  ${space.price_per_slot.toLocaleString('es-AR')}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  por turno
                </Typography>
              </Box>

              {!isAuthenticated && (
                <Alert severity="info" icon={<LockIcon />} sx={{ mt: 2 }}>
                  <Typography variant="body2">
                    <Link to="/login" style={{ color: 'inherit', fontWeight: 600 }}>
                      Iniciá sesión
                    </Link>{' '}
                    para poder reservar.
                  </Typography>
                </Alert>
              )}
            </Box>
          </Grid>

          {/* Slots derecha */}
          <Grid size={{ xs: 12, md: 8 }}>
            {/* Date picker */}
            <Box sx={{ mb: 4 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                <CalendarMonthIcon color="primary" />
                <Typography variant="h6">Seleccioná una fecha</Typography>
              </Box>
              <TextField
                type="date"
                value={date}
                onChange={(e) => {
                  setDate(e.target.value);
                  setConflictSlotId(null);
                }}
                inputProps={{ min: today }}
                fullWidth
                sx={{ maxWidth: 280 }}
              />
            </Box>

            {/* Slots */}
            {date && (
              <>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2.5 }}>
                  <AccessTimeIcon color="primary" />
                  <Typography variant="h6">Turnos disponibles</Typography>
                </Box>

                {loadingSlots ? (
                  <Grid container spacing={2}>
                    {Array.from({ length: 4 }).map((_, i) => (
                      <Grid key={i} size={{ xs: 12, sm: 6 }}>
                        <Skeleton variant="rectangular" height={96} sx={{ borderRadius: 2 }} />
                      </Grid>
                    ))}
                  </Grid>
                ) : slots?.length === 0 ? (
                  <Alert severity="info">Este espacio no tiene turnos configurados aún.</Alert>
                ) : (
                  <Grid container spacing={2}>
                    {slots?.map((slot, i) => {
                      const isBooking =
                        bookMutation.isPending && selectedSlot?.id === slot.id;
                      const hasConflict = conflictSlotId === slot.id;

                      return (
                        <Grid key={slot.id} size={{ xs: 12, sm: 6 }}>
                          <MotionCard
                            initial={{ opacity: 0, y: 16 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: i * 0.06 }}
                            sx={{
                              border: hasConflict
                                ? '1.5px solid'
                                : '1.5px solid transparent',
                              borderColor: hasConflict ? 'error.main' : 'transparent',
                            }}
                          >
                            <CardContent
                              sx={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'space-between',
                                gap: 2,
                                p: '16px !important',
                              }}
                            >
                              <Box>
                                <Typography variant="h6" sx={{ fontSize: '1rem' }}>
                                  {slot.label}
                                </Typography>
                                {slot.start_time && slot.end_time && (
                                  <Typography variant="body2" color="text.secondary">
                                    {slot.start_time} – {slot.end_time}
                                  </Typography>
                                )}
                                {hasConflict && (
                                  <Typography variant="caption" color="error">
                                    Ya está reservado para esta fecha
                                  </Typography>
                                )}
                              </Box>
                              <Button
                                variant="contained"
                                size="small"
                                onClick={() => handleBook(slot)}
                                disabled={isBooking}
                                sx={{ flexShrink: 0 }}
                              >
                                {isBooking ? 'Reservando...' : 'Reservar'}
                              </Button>
                            </CardContent>
                          </MotionCard>
                        </Grid>
                      );
                    })}
                  </Grid>
                )}
              </>
            )}

            {!date && (
              <Box
                sx={{
                  border: '2px dashed',
                  borderColor: 'divider',
                  borderRadius: 3,
                  py: 8,
                  textAlign: 'center',
                }}
              >
                <CalendarMonthIcon sx={{ fontSize: 48, color: 'text.disabled', mb: 1 }} />
                <Typography color="text.secondary">
                  Elegí una fecha para ver los turnos disponibles
                </Typography>
              </Box>
            )}
          </Grid>
        </Grid>
      </Container>

      {/* Snackbar de éxito */}
      <Snackbar
        open={successOpen}
        autoHideDuration={3000}
        onClose={() => setSuccessOpen(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity="success" onClose={() => setSuccessOpen(false)}>
          ¡Reserva confirmada! Redirigiendo a tus reservas...
        </Alert>
      </Snackbar>
    </Box>
  );
}

function SpaceDetailSkeleton() {
  return (
    <Box>
      <Skeleton variant="rectangular" height={320} />
      <Container maxWidth="lg" sx={{ py: 5 }}>
        <Grid container spacing={5}>
          <Grid size={{ xs: 12, md: 4 }}>
            <Skeleton height={28} sx={{ mb: 1 }} />
            <Skeleton height={28} width="60%" sx={{ mb: 3 }} />
            <Skeleton height={48} width="50%" />
          </Grid>
          <Grid size={{ xs: 12, md: 8 }}>
            <Skeleton height={56} width={280} sx={{ mb: 4 }} />
          </Grid>
        </Grid>
      </Container>
    </Box>
  );
}
