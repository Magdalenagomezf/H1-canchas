import { cn } from '@/lib/utils';
import { COMPLEX_FEATURES, IMAGES } from './content';
import { FadeIn, ImageReveal } from './Reveal';
import { SectionLabel } from './SectionLabel';
import { CONTAINER, MONO } from './ui';

export function Complex() {
  return (
    <section id="complejo" aria-label="El complejo" className="on-dark relative bg-night text-paper">
      <ImageReveal
        src={IMAGES.complex}
        alt="Locales comerciales y galería del complejo H1"
        className="h-[80vh] min-h-[480px] w-full"
        parallax
      />
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 bg-gradient-to-t from-night/90 via-night/10 to-night/40" />

      <div className="absolute inset-0">
       <div className={cn(CONTAINER, 'flex h-full flex-col justify-between py-12 md:py-16')}>
        <SectionLabel number="03" label="El complejo" />
        <FadeIn>
          <ul className={cn(MONO, 'flex flex-wrap gap-x-3 gap-y-2 border-t border-paper/30 pt-6 text-paper')}>
            {COMPLEX_FEATURES.map((feature, i) => (
              <li key={feature} className="flex items-center gap-3">
                {i > 0 && (
                  <span aria-hidden="true" className="text-light">
                    ·
                  </span>
                )}
                {feature}
              </li>
            ))}
          </ul>
        </FadeIn>
       </div>
      </div>
    </section>
  );
}
