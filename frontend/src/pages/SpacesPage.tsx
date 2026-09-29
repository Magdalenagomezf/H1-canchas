import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { LayoutGrid } from 'lucide-react';
import { motion } from 'framer-motion';
import { getSpaces } from '../api/spaces';
import { SpaceCard, SpaceCardSkeleton } from '../components/SpaceCard';
import type { SpaceType } from '../types';
import { cn } from '@/lib/utils';

type Filter = SpaceType | null;

const FILTERS: { label: string; value: Filter }[] = [
  { label: 'Todos', value: null },
  { label: 'Pádel', value: 'cancha_padel' },
  { label: 'Fútbol', value: 'cancha_futbol' },
  { label: 'Padbol', value: 'cancha_padbol' },
  { label: 'Beach vóley', value: 'cancha_beach_voley' },
  { label: 'Quincho', value: 'quincho' },
];

export default function SpacesPage() {
  const [filter, setFilter] = useState<Filter>(null);

  const { data: spaces, isLoading } = useQuery({
    queryKey: ['spaces'],
    queryFn: getSpaces,
  });

  const filtered = filter ? spaces?.filter((s) => s.type === filter) : spaces;

  return (
    <div className="animate-fade-up min-h-[calc(100vh-58px)] bg-bg">

      {/* Header */}
      <div className="bg-surface border-b border-black/[0.06]">
        <div className="max-w-[1140px] mx-auto px-6 py-12">
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
          >
            <span className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-1.5 block">
              Disponibles ahora
            </span>
            <h1 className="font-serif text-4xl text-ink tracking-tight mb-6">
              Nuestros espacios
            </h1>

            {/* Filter chips */}
            <div className="flex flex-wrap gap-2">
              {FILTERS.map((f) => (
                <button
                  key={f.label}
                  onClick={() => setFilter(f.value)}
                  className={cn(
                    'rounded-full px-4 py-1.5 text-sm font-semibold transition-all duration-normal active:scale-[0.98]',
                    filter === f.value
                      ? 'bg-primary text-white shadow-sm'
                      : 'bg-white border-[1.5px] border-black/[0.07] text-ink-2 hover:border-primary/40 hover:text-ink',
                  )}
                >
                  {f.label}
                </button>
              ))}
            </div>
          </motion.div>
        </div>
      </div>

      {/* Grid */}
      <div className="max-w-[1140px] mx-auto px-6 py-10">
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {Array.from({ length: 3 }).map((_, i) => <SpaceCardSkeleton key={i} />)}
          </div>
        ) : !filtered?.length ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            className="flex flex-col items-center justify-center py-24 text-center"
          >
            <div className="w-14 h-14 rounded-xl bg-surface flex items-center justify-center mb-4">
              <LayoutGrid size={24} className="text-ink-2/50" />
            </div>
            <p className="text-base font-semibold text-ink mb-1">Sin resultados</p>
            <p className="text-sm text-ink-2">No hay espacios en esta categoría por el momento.</p>
          </motion.div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {filtered.map((space, i) => (
              <SpaceCard key={space.id} space={space} index={i} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
