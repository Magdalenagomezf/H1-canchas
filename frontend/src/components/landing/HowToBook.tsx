import { useRef } from 'react';
import { motion, useInView, useReducedMotion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { STEPS } from './content';
import { FadeIn } from './Reveal';
import { SectionLabel } from './SectionLabel';
import { CONTAINER, EASE, SECTION_Y, VIEWPORT } from './ui';

const LINE = 'absolute bg-white/70';

/** One court marking that draws in when `show` flips to true. */
function CourtLine({
  axis,
  origin,
  show,
  delay = 0,
  className,
}: {
  axis: 'x' | 'y';
  origin: string;
  show: boolean;
  delay?: number;
  className?: string;
}) {
  const reduce = useReducedMotion();
  const hidden = reduce ? { opacity: 0 } : axis === 'x' ? { scaleX: 0 } : { scaleY: 0 };
  const shown = reduce ? { opacity: 1 } : axis === 'x' ? { scaleX: 1 } : { scaleY: 1 };
  return (
    <motion.div
      aria-hidden="true"
      className={cn(LINE, origin, className)}
      initial={hidden}
      animate={show ? shown : hidden}
      transition={{ duration: reduce ? 0.4 : 1.1, ease: EASE, delay }}
    />
  );
}

/**
 * Three steps joined by a single continuous line (no boxes).
 * Desktop: horizontal line above the steps. Mobile: vertical line on the left.
 */
export function HowToBook() {
  const courtRef = useRef<HTMLDivElement>(null);
  // Observe the unclipped/unscaled wrapper, never the drawn lines.
  const show = useInView(courtRef, VIEWPORT);

  return (
    <section aria-label="Cómo reservar" className={cn('on-dark bg-dusk text-white', SECTION_Y)}>
      <div className={CONTAINER}>
        <SectionLabel number="04" label="Cómo reservar" tone="court" />

        <div ref={courtRef} className="relative mt-14 md:mt-24">
          {/* One continuous line joining the steps: horizontal on desktop, vertical on mobile */}
          <CourtLine axis="x" origin="origin-left" show={show} className="left-0 top-0 hidden h-px w-full md:block" />
          <CourtLine axis="y" origin="origin-top" show={show} className="left-0 top-0 h-full w-px md:hidden" />

          <ol className="relative grid grid-cols-1 gap-y-14 pl-8 md:grid-cols-3 md:gap-x-10 md:pl-0 md:pt-12 lg:gap-x-16 lg:pt-16">
            {STEPS.map((step, i) => (
              <li key={step.number}>
                <FadeIn delay={i * 0.12}>
                  <span className="block font-arch font-expanded text-[clamp(3.5rem,7vw,6.5rem)] font-extrabold leading-[0.9] tracking-[-0.02em] text-white">
                    {step.number}
                    <span className="text-light">/</span>
                  </span>
                  <h3 className="mt-6 text-xl font-semibold leading-tight text-white">{step.title}</h3>
                  <p className="mt-2 max-w-xs text-base leading-relaxed text-white/85">{step.text}</p>
                </FadeIn>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}
