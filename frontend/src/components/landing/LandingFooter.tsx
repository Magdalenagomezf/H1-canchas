import { ArrowUpRight } from 'lucide-react';
import { buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { FOOTER } from './content';
import { LightLine } from './LightLine';
import { CONTAINER, GRID, MONO } from './ui';

function ColumnLabel({ children }: { children: string }) {
  return <h2 className={cn(MONO, 'mb-5 text-concrete')}>{children}</h2>;
}

export function LandingFooter() {
  return (
    <footer id="ubicacion" className="on-dark bg-night pb-10 text-paper">
      <LightLine className="mb-16 md:mb-24" />
      <div className={CONTAINER}>
        <div className={cn(GRID, 'gap-y-14')}>
          <div className="col-span-12 md:col-span-5">
            <ColumnLabel>Ubicación</ColumnLabel>
            <p className="text-xl font-medium leading-snug text-paper">{FOOTER.address}</p>
            <a
              href={FOOTER.mapsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(buttonVariants({ variant: 'line' }), 'px-0 mt-6 h-auto gap-2 pb-1 text-sm font-semibold text-paper')}
            >
              Ver en Google Maps
              <ArrowUpRight size={16} className="transition-transform duration-300 group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
            </a>
          </div>

          <div className="col-span-12 md:col-span-4">
            <ColumnLabel>Contacto</ColumnLabel>
            <a
              href={FOOTER.whatsappUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(buttonVariants({ variant: 'court' }), 'h-12 gap-3 px-6 text-sm font-semibold')}
            >
              {FOOTER.whatsappLabel}
              <ArrowUpRight size={16} className="transition-transform duration-300 group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
            </a>
          </div>

          <div className="col-span-12 md:col-span-3">
            <ColumnLabel>Horarios</ColumnLabel>
            <p className={cn(MONO, 'space-y-1 text-paper')}>
              {FOOTER.hours.map((line) => (
                <span key={line} className="block">
                  {line}
                </span>
              ))}
            </p>
          </div>
        </div>

        <p className={cn(MONO, 'mt-20 border-t border-paper/20 pt-6 text-concrete md:mt-28')}>{FOOTER.credit}</p>
      </div>
    </footer>
  );
}
