import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Grid from '@mui/material/Grid';
import SportsTennisIcon from '@mui/icons-material/SportsTennis';
import SportsSoccerIcon from '@mui/icons-material/SportsSoccer';
import CelebrationIcon from '@mui/icons-material/Celebration';
import GridViewIcon from '@mui/icons-material/GridView';
import { motion } from 'framer-motion';
import { getSpaces } from '../api/spaces';
import { SpaceCard, SpaceCardSkeleton } from '../components/SpaceCard';
import type { SpaceType } from '../types';

const MotionBox = motion(Box);

type Filter = SpaceType | 'all';

const FILTERS: { value: Filter; label: string; icon: React.ReactNode }[] = [
  { value: 'all', label: 'Todos', icon: <GridViewIcon fontSize="small" /> },
  { value: 'cancha_padel', label: 'Pádel', icon: <SportsTennisIcon fontSize="small" /> },
  { value: 'cancha_futbol', label: 'Fútbol', icon: <SportsSoccerIcon fontSize="small" /> },
  { value: 'quincho', label: 'Quincho / Salón', icon: <CelebrationIcon fontSize="small" /> },
];

export default function SpacesPage() {
  const [filter, setFilter] = useState<Filter>('all');

  const { data: spaces, isLoading } = useQuery({
    queryKey: ['spaces'],
    queryFn: getSpaces,
  });

  const filtered = filter === 'all' ? spaces : spaces?.filter((s) => s.type === filter);

  return (
    <Box>
      {/* Header */}
      <Box
        sx={{
          background: 'linear-gradient(160deg, #2D5A3D 0%, #3D7A4E 60%, #5A9E6A 100%)',
          py: { xs: 6, md: 8 },
        }}
      >
        <Container maxWidth="lg">
          <MotionBox
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <Typography
              variant="h2"
              sx={{ color: '#fff', fontSize: { xs: '2rem', md: '3rem' }, mb: 1.5 }}
            >
              Nuestros espacios
            </Typography>
            <Typography sx={{ color: 'rgba(255,255,255,0.8)', fontSize: '1.05rem' }}>
              Elegí el espacio que más te convenga y reservá tu turno al instante.
            </Typography>
          </MotionBox>
        </Container>
      </Box>

      <Container maxWidth="lg" sx={{ py: { xs: 5, md: 8 } }}>
        {/* Filtros */}
        <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap', mb: 5 }}>
          {FILTERS.map((f) => (
            <Chip
              key={f.value}
              icon={f.icon as React.ReactElement}
              label={f.label}
              clickable
              onClick={() => setFilter(f.value)}
              variant={filter === f.value ? 'filled' : 'outlined'}
              sx={{
                fontWeight: 600,
                fontSize: '0.9rem',
                px: 0.5,
                ...(filter === f.value
                  ? { bgcolor: 'primary.main', color: '#fff', '& .MuiChip-icon': { color: '#fff' } }
                  : { borderColor: 'primary.main', color: 'primary.main', '& .MuiChip-icon': { color: 'primary.main' } }),
              }}
            />
          ))}
        </Box>

        {/* Grid */}
        <Grid container spacing={3}>
          {isLoading
            ? Array.from({ length: 3 }).map((_, i) => (
                <Grid key={i} size={{ xs: 12, sm: 6, md: 4 }}>
                  <SpaceCardSkeleton />
                </Grid>
              ))
            : filtered?.length === 0
              ? (
                <Grid size={12}>
                  <Box sx={{ textAlign: 'center', py: 10 }}>
                    <Typography variant="h5" color="text.secondary" sx={{ mb: 1 }}>
                      No hay espacios de este tipo disponibles
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Probá con otro filtro o volvé más tarde.
                    </Typography>
                  </Box>
                </Grid>
              )
              : filtered?.map((space, i) => (
                  <Grid key={space.id} size={{ xs: 12, sm: 6, md: 4 }}>
                    <SpaceCard space={space} index={i} />
                  </Grid>
                ))}
        </Grid>
      </Container>
    </Box>
  );
}
