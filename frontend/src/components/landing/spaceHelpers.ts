import { SPACE_IMAGES } from '@/components/SpaceCard';
import type { Space } from '@/types';

export const getSpaceImage = (space: Space) => SPACE_IMAGES[space.type]?.[0];

const priceFormatter = new Intl.NumberFormat('es-AR');
export const formatPrice = (value: number) => priceFormatter.format(value);

export const SPACES_ERROR_MESSAGE = 'No pudimos cargar los espacios. Probá de nuevo en unos minutos.';
export const SPACES_EMPTY_MESSAGE = 'Pronto vas a poder ver acá los espacios disponibles.';
