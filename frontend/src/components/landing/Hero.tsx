import { Link } from 'react-router-dom';
import { motion, useReducedMotion } from 'framer-motion';
import { ArrowRight } from 'lucide-react';
import { buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { HERO_HOURS, HERO_SUBLINE, HERO_TITLE, IMAGES } from './content';
import { LightLine } from './LightLine';
import { FadeIn, ImageReveal, LineTitle } from './Reveal';
import { CONTAINER, MONO } from './ui';

export function Hero() {
  const reduce = useReducedMotion();
  return (
    <section
      aria-label="Inicio"
      className="on-dark relative flex h-[100svh] min-h-[600px] flex-col justify-end overflow-hidden bg-night text-paper"
    >
      <ImageReveal
        src={IMAGES.hero}
        alt="Fachada nocturna del complejo H1 Espacio Deportivo, iluminada"
        className="absolute inset-0"
        eager
        parallax
        reveal={false}
      />
      <div aria-hidden="true" className="absolute inset-0 bg-gradient-to-t from-night via-night/40 to-transparent" />
      <div aria-hidden="true" className="absolute inset-0 bg-gradient-to-r from-night/80 via-night/20 to-transparent" />

      {/* Title block: lit from above by the LED under the slab */}
      <div className="relative">
        <motion.div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 top-0 h-[180px] bg-gradient-to-b from-light/15 to-transparent md:h-[200px]"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: reduce ? 0.4 : 1.4, ease: 'easeOut', delay: reduce ? 0.5 : 1.7 }}
        />
        <LightLine className="absolute inset-x-0 top-0" delay={0.5} />

        <div className={cn(CONTAINER, 'relative pb-14 pt-12 md:pb-20 md:pt-16')}>
          <LineTitle
            as="h1"
            lines={HERO_TITLE}
            delay={0.2}
            className="font-arch font-expanded text-[clamp(2.25rem,min(7vw,11vh),7.5rem)] font-extrabold uppercase leading-[0.9] tracking-[-0.02em] text-paper"
          />

          <div className="mt-10 flex flex-col gap-10 md:mt-14 md:flex-row md:items-end md:justify-between">
            <FadeIn delay={0.6} className="max-w-md">
              <p className="text-base leading-relaxed text-paper/85 md:text-lg">{HERO_SUBLINE}</p>
              <div className="mt-8 flex flex-wrap items-center gap-x-10 gap-y-5">
                <Link to="/canchas" className={cn(buttonVariants({ variant: 'court' }), 'h-12 gap-3 px-7 text-sm font-semibold')}>
                  Reservar turno
                  <ArrowRight size={16} className="transition-transform duration-300 group-hover:translate-x-1" />
                </Link>
                <a
                  href="#espacios"
                  className={cn(buttonVariants({ variant: 'line' }), 'px-0 h-auto gap-2 pb-1 text-sm font-semibold text-paper')}
                >
                  Ver espacios
                </a>
              </div>
            </FadeIn>

            <FadeIn delay={0.8}>
              <p className={cn(MONO, 'text-paper/80')}>{HERO_HOURS}</p>
            </FadeIn>
          </div>
        </div>
      </div>
    </section>
  );
}
