import { Link } from 'react-router-dom';
import { motion, useReducedMotion } from 'framer-motion';
import { ArrowUpRight } from 'lucide-react';
import type { Space } from '@/types';
import { cn } from '@/lib/utils';
import { SPACE_TYPE_LABEL } from './content';
import { SPACES_EMPTY_MESSAGE, SPACES_ERROR_MESSAGE, formatPrice, getSpaceImage } from './spaceHelpers';
import { EASE, MONO } from './ui';

const GRID_CLASS = 'grid grid-cols-1 gap-x-6 gap-y-14 md:grid-cols-2 lg:grid-cols-3';

function SpaceCardItem({ space, index }: { space: Space; index: number }) {
  const reduce = useReducedMotion();
  const src = getSpaceImage(space);
  return (
    <motion.li
      initial={{ opacity: 0, y: reduce ? 0 : 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: EASE, delay: reduce ? 0 : Math.min(index, 8) * 0.04 }}
    >
      <Link to={`/canchas/${space.id}`} className="group block">
        <div className="aspect-[4/3] w-full overflow-hidden bg-graphite">
          {src && (
            <img
              src={src}
              alt={`Foto de ${space.name}`}
              loading="lazy"
              decoding="async"
              className="h-full w-full object-cover transition-transform duration-700 ease-arch group-hover:scale-[1.03] motion-reduce:transition-none motion-reduce:group-hover:scale-100"
            />
          )}
        </div>

        <p className={cn(MONO, 'mt-5 text-concrete')}>
          <span className="text-paper">{String(index + 1).padStart(2, '0')}</span> — {SPACE_TYPE_LABEL[space.type]}
        </p>

        <h3 className="mt-3 break-words font-arch font-expanded text-[clamp(1.25rem,2vw,1.75rem)] font-bold uppercase leading-[1.05] tracking-[-0.02em] text-paper">
          {space.name}
        </h3>
        <span
          aria-hidden="true"
          className="mt-3 block h-[2px] w-0 bg-light transition-[width] duration-500 ease-arch group-hover:w-12 group-focus-visible:w-12 motion-reduce:transition-none"
        />

        {space.description && <p className="mt-3 line-clamp-1 text-sm leading-relaxed text-paper/70">{space.description}</p>}

        <div className="mt-4 flex items-center justify-between gap-4">
          <span className={cn(MONO, 'text-paper')}>${formatPrice(space.price_per_slot)} / turno</span>
          <ArrowUpRight
            aria-hidden="true"
            size={22}
            className="shrink-0 text-paper transition-[transform,color] duration-500 ease-arch group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-light motion-reduce:transition-none motion-reduce:group-hover:transform-none"
          />
        </div>
      </Link>
    </motion.li>
  );
}

function SpaceCardSkeleton() {
  const reduce = useReducedMotion();
  return (
    <li aria-hidden="true">
      <motion.div
        animate={reduce ? undefined : { opacity: [0.45, 1, 0.45] }}
        transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }}
      >
        <div className="aspect-[4/3] w-full bg-graphite" />
        <div className="mt-5 h-3 w-28 bg-graphite" />
        <div className="mt-3 h-6 w-3/4 bg-graphite" />
        <div className="mt-4 h-3 w-24 bg-graphite" />
      </motion.div>
    </li>
  );
}

interface SpacesGridProps {
  spaces: Space[];
  isLoading?: boolean;
  isError?: boolean;
  emptyMessage?: string;
}

/** Compact grid of spaces: small images, 3 per row on desktop. */
export function SpacesGrid({ spaces, isLoading = false, isError = false, emptyMessage = SPACES_EMPTY_MESSAGE }: SpacesGridProps) {
  if (isLoading) {
    return (
      <ul className={GRID_CLASS}>
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <SpaceCardSkeleton key={i} />
        ))}
      </ul>
    );
  }

  if (isError) {
    return <p className="max-w-md text-base leading-relaxed text-paper/70">{SPACES_ERROR_MESSAGE}</p>;
  }

  if (spaces.length === 0) {
    return <p className="max-w-md text-base leading-relaxed text-paper/70">{emptyMessage}</p>;
  }

  return (
    <ul className={GRID_CLASS}>
      {spaces.map((space, i) => (
        <SpaceCardItem key={space.id} space={space} index={i} />
      ))}
    </ul>
  );
}
