import { useQuery } from '@tanstack/react-query';
import { getSpaces } from '@/api/spaces';
import { cn } from '@/lib/utils';
import { LineTitle } from './Reveal';
import { SectionLabel } from './SectionLabel';
import { SpacesIndex } from './SpacesIndex';
import { CONTAINER, SECTION_Y } from './ui';

export function Spaces() {
  const { data: spaces, isLoading, isError } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });

  return (
    <section id="espacios" aria-label="Espacios" className={cn('on-dark bg-night text-paper', SECTION_Y)}>
      <div className={CONTAINER}>
        <SectionLabel number="02" label="Espacios" />
        <LineTitle
          lines={['Elegí dónde', 'jugar']}
          className="mt-10 font-arch font-expanded text-[clamp(2.25rem,6vw,5.5rem)] font-extrabold uppercase leading-[0.92] tracking-[-0.02em] text-paper"
        />

        <div className="mt-16 md:mt-24">
          <SpacesIndex spaces={spaces ?? []} isLoading={isLoading} isError={isError} />
        </div>
      </div>
    </section>
  );
}
