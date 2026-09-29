import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion';
import { getSpaces } from '../api/spaces';
import type { SpaceType } from '../types';
import { LandingFooter } from '@/components/landing/LandingFooter';
import { LandingNav } from '@/components/landing/LandingNav';
import { LightLine } from '@/components/landing/LightLine';
import { SectionLabel } from '@/components/landing/SectionLabel';
import { SpacesGrid } from '@/components/landing/SpacesGrid';
import { SpacesIndex } from '@/components/landing/SpacesIndex';
import { SPACE_TYPE_LABEL } from '@/components/landing/content';
import { CONTAINER, EASE, MONO } from '@/components/landing/ui';
import { cn } from '@/lib/utils';

type Filter = SpaceType | null;

const FILTERS: { label: string; value: Filter }[] = [
  { label: 'Todos', value: null },
  ...(Object.keys(SPACE_TYPE_LABEL) as SpaceType[]).map((type) => ({
    label: SPACE_TYPE_LABEL[type],
    value: type as Filter,
  })),
];

export default function SpacesPage() {
  const reduce = useReducedMotion();
  const [filter, setFilter] = useState<Filter>(null);

  const { data: spaces, isLoading, isError } = useQuery({
    queryKey: ['spaces'],
    queryFn: getSpaces,
  });

  const filtered = filter ? (spaces ?? []).filter((s) => s.type === filter) : (spaces ?? []);
  const y = reduce ? 0 : 12;

  return (
    <div className="landing">
      <div className="on-dark min-h-[100svh] bg-night text-paper">
        <LandingNav solid />

        <main className="relative pt-16 md:pt-20">
          {/* Slab LED with its soft warm wash, same as the hero */}
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-x-0 top-16 h-[160px] bg-gradient-to-b from-light/10 to-transparent md:top-20 md:h-[200px]"
          />
          <LightLine className="absolute inset-x-0 top-16 md:top-20" delay={0.3} />

          <div className={cn(CONTAINER, 'relative pb-20 pt-14 md:pb-28 md:pt-20 lg:pb-32')}>
            <SectionLabel number="H1" label="Espacios" />

            <h1 className="mt-10 font-arch font-expanded text-[clamp(2.25rem,min(9vw,14vh),7.5rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper md:mt-14">
              <motion.span
                className="block"
                initial={{ opacity: 0, y: reduce ? 0 : 28 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.7, ease: EASE, delay: 0.15 }}
              >
                Espacios
              </motion.span>
            </h1>
            <motion.p
              className="mt-6 max-w-md text-base leading-relaxed text-paper/85 md:text-lg"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, ease: EASE, delay: 0.35 }}
            >
              Elegí dónde jugar y reservá tu turno online.
            </motion.p>

            <motion.div
              className="mt-14 flex flex-col gap-6 md:mt-20 md:flex-row md:items-end md:justify-between"
              initial={{ opacity: 0, y }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, ease: EASE, delay: 0.45 }}
            >
              <div role="group" aria-label="Filtrar por tipo" className="flex flex-wrap gap-x-8 gap-y-4">
                {FILTERS.map((f) => {
                  const selected = filter === f.value;
                  return (
                    <button
                      key={f.label}
                      type="button"
                      aria-pressed={selected}
                      onClick={() => setFilter(f.value)}
                      className={cn(
                        MONO,
                        'relative pb-3 outline-none transition-colors duration-300 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-paper',
                        selected ? 'text-paper' : 'text-concrete hover:text-paper',
                      )}
                    >
                      {f.label}
                      {selected && (
                        <motion.span
                          layoutId="spaces-filter-led"
                          aria-hidden="true"
                          className="absolute inset-x-0 bottom-0 h-[2px] bg-light"
                          transition={{ duration: reduce ? 0 : 0.35, ease: EASE }}
                        />
                      )}
                    </button>
                  );
                })}
              </div>

              {!isLoading && !isError && (
                <p className={cn(MONO, 'text-concrete')} aria-live="polite">
                  <span className="text-paper">{String(filtered.length).padStart(2, '0')}</span>{' '}
                  {filtered.length === 1 ? 'Espacio' : 'Espacios'}
                </p>
              )}
            </motion.div>

            <div className="mt-10 md:mt-14">
              <AnimatePresence mode="wait" initial={false}>
                <motion.div
                  key={filter ?? 'all'}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: reduce ? 0.15 : 0.3, ease: EASE }}
                >
                  {filter === null ? (
                    <SpacesGrid spaces={filtered} isLoading={isLoading} isError={isError} />
                  ) : (
                    <SpacesIndex
                      spaces={filtered}
                      isLoading={isLoading}
                      isError={isError}
                      emptyMessage="No hay espacios de este tipo por ahora."
                    />
                  )}
                </motion.div>
              </AnimatePresence>
            </div>
          </div>
        </main>

        <LandingFooter />
      </div>
    </div>
  );
}
