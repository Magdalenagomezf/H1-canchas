import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutlined';
import SportsSoccerIcon from '@mui/icons-material/SportsSoccer';
import SportsTennisIcon from '@mui/icons-material/SportsTennis';
import CelebrationIcon from '@mui/icons-material/Celebration';
import { motion } from 'framer-motion';
import { getSpaces } from '../api/spaces';
import { SpaceCard, SpaceCardSkeleton } from '../components/SpaceCard';

const MotionBox = motion(Box);
const MotionButton = motion(Button);

const STEPS = [
  {
    icon: <SportsSoccerIcon sx={{ fontSize: 36 }} />,
    title: 'Elegí tu espacio',
    desc: 'Explorá nuestras canchas de pádel, fútbol y quinchos disponibles.',
  },
  {
    icon: <CalendarMonthIcon sx={{ fontSize: 36 }} />,
    title: 'Seleccioná fecha y turno',
    desc: 'Chequeá la disponibilidad en tiempo real y elegí el horario que más te convenga.',
  },
  {
    icon: <CheckCircleOutlineIcon sx={{ fontSize: 36 }} />,
    title: 'Confirmá tu reserva',
    desc: 'Recibís la confirmación al instante. ¡Listo para jugar!',
  },
];

export default function HomePage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const { data: spaces, isLoading } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });

  return (
    <Box>
      {/* Hero */}
      <Box
        sx={{
          minHeight: { xs: '75vh', md: '85vh' },
          background: 'linear-gradient(160deg, #2D5A3D 0%, #3D7A4E 50%, #5A9E6A 100%)',
          display: 'flex',
          alignItems: 'center',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* Decorative circles */}
        <Box
          sx={{
            position: 'absolute',
            width: 500,
            height: 500,
            borderRadius: '50%',
            background: 'rgba(255,255,255,0.04)',
            right: -100,
            top: -100,
          }}
        />
        <Box
          sx={{
            position: 'absolute',
            width: 300,
            height: 300,
            borderRadius: '50%',
            background: 'rgba(255,255,255,0.04)',
            left: -80,
            bottom: -80,
          }}
        />

        <Container maxWidth="lg" sx={{ position: 'relative', zIndex: 1 }}>
          <MotionBox
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, ease: 'easeOut' }}
            sx={{ maxWidth: 680 }}
          >
            <Typography
              component="span"
              sx={{
                display: 'inline-block',
                bgcolor: 'rgba(255,255,255,0.15)',
                color: '#fff',
                px: 2,
                py: 0.5,
                borderRadius: 4,
                fontSize: '0.85rem',
                fontWeight: 600,
                letterSpacing: '0.08em',
                textTransform: 'uppercase',
                mb: 3,
              }}
            >
              Reservas online
            </Typography>

            <Typography
              variant="h1"
              sx={{
                color: '#fff',
                fontSize: { xs: '2.8rem', md: '4rem' },
                lineHeight: 1.1,
                mb: 3,
              }}
            >
              Tu próxima jugada empieza aquí
            </Typography>

            <Typography
              variant="body1"
              sx={{ color: 'rgba(255,255,255,0.85)', fontSize: '1.1rem', mb: 5, maxWidth: 520 }}
            >
              Reservá canchas de pádel, fútbol y quinchos de forma rápida y sencilla. Disponibilidad
              en tiempo real, sin llamadas ni esperas.
            </Typography>

            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              <MotionButton
                variant="contained"
                size="large"
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.97 }}
                onClick={() => navigate('/canchas')}
                sx={{
                  bgcolor: '#fff',
                  color: 'primary.dark',
                  px: 4,
                  py: 1.5,
                  fontSize: '1rem',
                  '&:hover': { bgcolor: '#f5f5f0' },
                }}
              >
                Ver canchas
              </MotionButton>
              {!isAuthenticated && (
                <MotionButton
                  variant="outlined"
                  size="large"
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.97 }}
                  onClick={() => navigate('/login', { state: { tab: 'register' } })}
                  sx={{
                    borderColor: 'rgba(255,255,255,0.6)',
                    color: '#fff',
                    px: 4,
                    py: 1.5,
                    fontSize: '1rem',
                    '&:hover': { borderColor: '#fff', bgcolor: 'rgba(255,255,255,0.08)' },
                  }}
                >
                  Registrarme
                </MotionButton>
              )}
            </Box>
          </MotionBox>
        </Container>
      </Box>

      {/* Spaces grid */}
      <Container maxWidth="lg" sx={{ py: { xs: 6, md: 10 } }}>
        <MotionBox
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          sx={{ mb: 6, textAlign: 'center' }}
        >
          <Typography variant="h2" sx={{ fontSize: { xs: '2rem', md: '2.6rem' }, mb: 1.5 }}>
            Nuestros espacios
          </Typography>
          <Typography variant="body1" color="text.secondary" sx={{ maxWidth: 480, mx: 'auto' }}>
            Canchas techadas e iluminadas, listos para usarlos en cualquier momento.
          </Typography>
        </MotionBox>

        <Grid container spacing={3}>
          {isLoading
            ? Array.from({ length: 3 }).map((_, i) => (
                <Grid key={i} size={{ xs: 12, sm: 6, md: 4 }}>
                  <SpaceCardSkeleton />
                </Grid>
              ))
            : spaces?.map((space, i) => (
                <Grid key={space.id} size={{ xs: 12, sm: 6, md: 4 }}>
                  <SpaceCard space={space} index={i} />
                </Grid>
              ))}
        </Grid>
      </Container>

      {/* How it works */}
      <Box sx={{ bgcolor: 'background.paper', py: { xs: 6, md: 10 } }}>
        <Container maxWidth="lg">
          <MotionBox
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            sx={{ textAlign: 'center', mb: 6 }}
          >
            <Typography variant="h2" sx={{ fontSize: { xs: '2rem', md: '2.6rem' }, mb: 1.5 }}>
              ¿Cómo funciona?
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Reservar es fácil y toma menos de 2 minutos.
            </Typography>
          </MotionBox>

          <Grid container spacing={4}>
            {STEPS.map((step, i) => (
              <Grid key={i} size={{ xs: 12, md: 4 }}>
                <MotionBox
                  initial={{ opacity: 0, y: 30 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.15, duration: 0.5 }}
                  sx={{ textAlign: 'center', p: 3 }}
                >
                  <Box
                    sx={{
                      width: 72,
                      height: 72,
                      borderRadius: '50%',
                      bgcolor: 'primary.main',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mx: 'auto',
                      mb: 2.5,
                      color: '#fff',
                    }}
                  >
                    {step.icon}
                  </Box>
                  <Typography
                    variant="h6"
                    sx={{ mb: 1.5, fontFamily: '"DM Serif Display", serif', fontSize: '1.3rem' }}
                  >
                    {step.title}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.7 }}>
                    {step.desc}
                  </Typography>
                </MotionBox>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>

      {/* CTA banner — only for unauthenticated users */}
      {!isAuthenticated && (
        <Box
          sx={{
            background: 'linear-gradient(135deg, #7B5E3A, #5C4429)',
            py: { xs: 7, md: 10 },
            textAlign: 'center',
          }}
        >
          <Container maxWidth="sm">
            <MotionBox
              initial={{ opacity: 0, scale: 0.96 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
            >
              <Typography
                variant="h2"
                sx={{ color: '#fff', fontSize: { xs: '1.8rem', md: '2.4rem' }, mb: 2 }}
              >
                ¿Listo para jugar?
              </Typography>
              <Typography sx={{ color: 'rgba(255,255,255,0.8)', mb: 4, fontSize: '1rem' }}>
                Creá tu cuenta gratis y empezá a reservar hoy mismo.
              </Typography>
              <MotionButton
                variant="contained"
                size="large"
                whileHover={{ scale: 1.06 }}
                whileTap={{ scale: 0.97 }}
                onClick={() => navigate('/login')}
                sx={{
                  bgcolor: '#fff',
                  color: 'secondary.dark',
                  px: 5,
                  py: 1.5,
                  fontSize: '1rem',
                  fontWeight: 700,
                  '&:hover': { bgcolor: '#f5f5f0' },
                }}
              >
                Crear cuenta
              </MotionButton>
            </MotionBox>
          </Container>
        </Box>
      )}

      {/* Footer */}
      <Box
        component="footer"
        sx={{
          bgcolor: '#2A2A28',
          color: 'rgba(255,255,255,0.6)',
          py: 4,
          textAlign: 'center',
        }}
      >
        <Typography variant="body2">
          © {new Date().getFullYear()} H1 Canchas. Todos los derechos reservados.
        </Typography>
      </Box>
    </Box>
  );
}
