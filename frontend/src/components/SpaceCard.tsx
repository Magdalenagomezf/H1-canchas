import { useNavigate } from 'react-router-dom';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Skeleton from '@mui/material/Skeleton';
import SportsTennisIcon from '@mui/icons-material/SportsTennis';
import SportsSoccerIcon from '@mui/icons-material/SportsSoccer';
import CelebrationIcon from '@mui/icons-material/Celebration';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import { motion } from 'framer-motion';
import type { Space, SpaceType } from '../types';

const MotionCard = motion(Card);
const MotionButton = motion(Button);

export const SPACE_IMAGES: Record<SpaceType, string> = {
  cancha_padel: 'https://images.unsplash.com/photo-1554068865-24cecd4e34b8?w=600&q=80',
  cancha_futbol: 'https://images.unsplash.com/photo-1508098682722-e99c43a406b2?w=600&q=80',
  quincho: 'https://images.unsplash.com/photo-1537640538966-79f369143f8f?w=600&q=80',
};

export const SPACE_ICONS: Record<SpaceType, typeof SportsTennisIcon> = {
  cancha_padel: SportsTennisIcon,
  cancha_futbol: SportsSoccerIcon,
  quincho: CelebrationIcon,
};

export const SPACE_LABELS: Record<SpaceType, string> = {
  cancha_padel: 'Pádel',
  cancha_futbol: 'Fútbol',
  quincho: 'Quincho / Salón',
};

export function SpaceCard({ space, index }: { space: Space; index: number }) {
  const navigate = useNavigate();
  const Icon = SPACE_ICONS[space.type];

  return (
    <MotionCard
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ delay: index * 0.08, duration: 0.45 }}
      whileHover={{ y: -6, boxShadow: '0 12px 32px rgba(61,122,78,0.18)' }}
      sx={{ height: '100%', display: 'flex', flexDirection: 'column', cursor: 'pointer' }}
      onClick={() => navigate(`/canchas/${space.id}`)}
    >
      <CardMedia
        component="img"
        height="200"
        image={SPACE_IMAGES[space.type]}
        alt={space.name}
        sx={{ objectFit: 'cover' }}
      />
      <CardContent sx={{ flexGrow: 1, p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Chip
            icon={<Icon />}
            label={SPACE_LABELS[space.type]}
            size="small"
            sx={{ bgcolor: 'primary.main', color: '#fff', '& .MuiChip-icon': { color: '#fff' } }}
          />
        </Box>
        <Typography variant="h5" sx={{ fontFamily: '"DM Serif Display", serif', mb: 1 }}>
          {space.name}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2, minHeight: 40 }}>
          {space.description}
        </Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6" color="primary.main">
            ${space.price_per_slot.toLocaleString('es-AR')}
            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 0.5 }}>
              / turno
            </Typography>
          </Typography>
          <MotionButton
            variant="contained"
            size="small"
            endIcon={<ArrowForwardIcon />}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.96 }}
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/canchas/${space.id}`);
            }}
          >
            Reservar
          </MotionButton>
        </Box>
      </CardContent>
    </MotionCard>
  );
}

export function SpaceCardSkeleton() {
  return (
    <Card>
      <Skeleton variant="rectangular" height={200} />
      <CardContent>
        <Skeleton width="40%" height={28} sx={{ mb: 1 }} />
        <Skeleton width="70%" height={32} sx={{ mb: 1 }} />
        <Skeleton width="100%" height={20} />
        <Skeleton width="80%" height={20} sx={{ mb: 2 }} />
        <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
          <Skeleton width={80} height={36} />
          <Skeleton width={100} height={36} />
        </Box>
      </CardContent>
    </Card>
  );
}
