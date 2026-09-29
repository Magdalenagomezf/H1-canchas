import { Link } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { FINAL_CTA_TITLE } from './content';
import { FadeIn, LineTitle } from './Reveal';
import { CONTAINER, SECTION_Y } from './ui';

export function FinalCta() {
  return (
    <section aria-label="Reservá tu turno" className={cn('on-dark bg-night text-paper', SECTION_Y)}>
      <div className={CONTAINER}>
        <LineTitle
          lines={FINAL_CTA_TITLE}
          className="font-arch font-expanded text-[clamp(3rem,11vw,10rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper"
        />
        <FadeIn delay={0.3} className="mt-12 md:mt-16">
          <Link to="/canchas" className={cn(buttonVariants({ variant: 'court' }), 'h-14 gap-3 px-9 text-base font-semibold')}>
            Reservar turno
            <ArrowRight size={18} className="transition-transform duration-300 group-hover:translate-x-1" />
          </Link>
        </FadeIn>
      </div>
    </section>
  );
}
