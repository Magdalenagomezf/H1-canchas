import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { motion, useReducedMotion } from 'framer-motion';
import { ArrowUpRight } from 'lucide-react';
import { getSpaces } from '@/api/spaces';
import { SPACE_IMAGES } from '@/components/SpaceCard';
import type { Space } from '@/types';
import { cn } from '@/lib/utils';
import { SPACE_TYPE_LABEL } from './content';
import { FadeIn, ImageReveal, LineTitle } from './Reveal';
import { SectionLabel } from './SectionLabel';
import { CONTAINER, EASE, GRID, MONO, SECTION_Y } from './ui';

const getSpaceImage = (space: Space) => SPACE_IMAGES[space.type]?.[0];
const priceFormatter = new Intl.NumberFormat('es-AR');

const NAME_SIZE = 'text-[clamp(2.5rem,5vw,6rem)]';

function SpaceRow({
  space,
  index,
  active,
  onActivate,
}: {
  space: Space;
  index: number;
  active: boolean;
  onActivate: () => void;
}) {
  return (
    <li className="relative border-t border-paper/10 last:border-b">
      <FadeIn>
        <Link
          to={`/canchas/${space.id}`}
          onMouseEnter={onActivate}
          onFocus={onActivate}
          className="group relative block py-7 md:py-9"
        >
          {/* Mobile / tablet: inline image */}
          <div className="mb-6 lg:hidden">
            <ImageReveal src={getSpaceImage(space)} alt={`Foto de ${space.name}`} className="aspect-[16/10] w-full" />
          </div>

          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between md:gap-8">
            <h3
              className={cn(
                'min-w-0 break-words font-arch font-expanded font-bold uppercase leading-[0.95] tracking-[-0.02em] transition-colors duration-500 ease-arch motion-reduce:transition-none',
                NAME_SIZE,
                active ? 'text-paper' : 'text-paper lg:text-paper/55',
              )}
            >
              {space.name}
            </h3>

            <div className="flex shrink-0 items-start gap-4 md:pt-2">
              <div className={cn(MONO, 'flex flex-col gap-1 text-concrete md:text-right')}>
                <span>
                  <span className="text-paper">{String(index + 1).padStart(2, '0')}</span> — {SPACE_TYPE_LABEL[space.type]}
                </span>
                <span className="text-paper">${priceFormatter.format(space.price_per_slot)} / turno</span>
              </div>
              <ArrowUpRight
                aria-hidden="true"
                size={28}
                className={cn(
                  'shrink-0 transition-[transform,color] duration-500 ease-arch motion-reduce:transition-none',
                  active ? 'text-light lg:-translate-y-0.5 lg:translate-x-0.5' : 'text-paper lg:text-paper/55',
                )}
              />
            </div>
          </div>

          {/* Always mounted so rows never shift; only visible for the active row on desktop */}
          {space.description && (
            <p
              className={cn(
                'mt-4 line-clamp-1 max-w-xl text-sm leading-relaxed text-paper/70 transition-opacity duration-500 motion-reduce:transition-none',
                active ? 'lg:opacity-100' : 'lg:opacity-0',
              )}
            >
              {space.description}
            </p>
          )}

          {/* LED underline on the active row */}
          <span
            aria-hidden="true"
            className={cn(
              'pointer-events-none absolute inset-x-0 -bottom-px hidden h-[2px] origin-left bg-light shadow-[0_0_8px_color-mix(in_srgb,var(--color-light)_50%,transparent)] transition-[transform,opacity] duration-700 ease-arch motion-reduce:transform-none lg:block',
              active ? 'scale-x-100 opacity-100' : 'scale-x-0 opacity-0',
            )}
          />
        </Link>
      </FadeIn>
    </li>
  );
}

type PanelState = 'active' | 'previous' | 'idle';

/** Sticky image panel: stacked images, the active one is revealed over the previous one. */
function SpacePanel({ spaces, active, previous }: { spaces: Space[]; active: number; previous: number }) {
  const reduce = useReducedMotion();
  const variants = reduce
    ? {
        idle: { opacity: 0, zIndex: 0, transition: { duration: 0 } },
        previous: { opacity: 1, zIndex: 1, transition: { duration: 0 } },
        active: { opacity: 1, zIndex: 2, transition: { duration: 0.3 } },
      }
    : {
        idle: { clipPath: 'inset(0 0 100% 0)', zIndex: 0, transition: { duration: 0 } },
        previous: { clipPath: 'inset(0 0 0% 0)', zIndex: 1, transition: { duration: 0 } },
        active: { clipPath: 'inset(0 0 0% 0)', zIndex: 2, transition: { duration: 0.6, ease: EASE } },
      };
  return (
    <div className="relative aspect-[4/5] w-full overflow-hidden bg-graphite">
      {spaces.map((space, i) => {
        const state: PanelState = i === active ? 'active' : i === previous ? 'previous' : 'idle';
        const src = getSpaceImage(space);
        return (
          <motion.div
            key={space.id}
            className="absolute inset-0"
            variants={variants}
            initial={i === active ? 'active' : 'idle'}
            animate={state}
            aria-hidden="true"
          >
            {src && <img src={src} alt="" loading="lazy" decoding="async" className="h-full w-full object-cover" />}
          </motion.div>
        );
      })}
    </div>
  );
}

function SpaceRowSkeleton() {
  const reduce = useReducedMotion();
  return (
    <li className="border-t border-paper/10 py-7 last:border-b md:py-9" aria-hidden="true">
      <motion.div
        animate={reduce ? undefined : { opacity: [0.45, 1, 0.45] }}
        transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }}
        className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between md:gap-8"
      >
        <div className="h-[clamp(2.5rem,5vw,4.5rem)] w-3/4 bg-graphite" />
        <div className="flex flex-col gap-2">
          <div className="h-3 w-28 bg-graphite" />
          <div className="h-3 w-24 bg-graphite" />
        </div>
      </motion.div>
    </li>
  );
}

export function Spaces() {
  const { data: spaces, isLoading, isError } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });
  const [{ active, previous }, setFocus] = useState({ active: 0, previous: -1 });

  const list = spaces ?? [];
  const ready = !isLoading && !isError && list.length > 0;
  const current = Math.min(active, Math.max(list.length - 1, 0));

  const activate = (i: number) => setFocus((s) => (s.active === i ? s : { active: i, previous: s.active }));

  return (
    <section id="espacios" aria-label="Espacios" className={cn('on-dark bg-night text-paper', SECTION_Y)}>
      <div className={CONTAINER}>
        <SectionLabel number="02" label="Espacios" />
        <LineTitle
          lines={['Elegí dónde', 'jugar']}
          className="mt-10 font-arch font-expanded text-[clamp(2.25rem,6vw,5.5rem)] font-extrabold uppercase leading-[0.92] tracking-[-0.02em] text-paper"
        />

        <div className="mt-16 md:mt-24">
          {isLoading && (
            <div className={GRID}>
              <ul className="col-span-12 lg:col-span-7">
                {[0, 1, 2].map((i) => (
                  <SpaceRowSkeleton key={i} />
                ))}
              </ul>
              <div className="col-span-5 hidden lg:block" aria-hidden="true">
                <div className="aspect-[4/5] w-full bg-graphite" />
              </div>
            </div>
          )}

          {ready && (
            <div className={GRID}>
              <ul className="col-span-12 lg:col-span-7">
                {list.map((space, i) => (
                  <SpaceRow
                    key={space.id}
                    space={space}
                    index={i}
                    active={i === current}
                    onActivate={() => activate(i)}
                  />
                ))}
              </ul>
              <div className="col-span-5 hidden lg:block">
                <div className="sticky top-[120px]">
                  <SpacePanel spaces={list} active={current} previous={previous === current ? -1 : previous} />
                </div>
              </div>
            </div>
          )}

          {!isLoading && !isError && list.length === 0 && (
            <p className="max-w-md text-base leading-relaxed text-paper/70">
              Pronto vas a poder ver acá los espacios disponibles.
            </p>
          )}

          {!isLoading && isError && (
            <p className="max-w-md text-base leading-relaxed text-paper/70">
              No pudimos cargar los espacios. Probá de nuevo en unos minutos.
            </p>
          )}
        </div>
      </div>
    </section>
  );
}
