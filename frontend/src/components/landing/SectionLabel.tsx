import { cn } from '@/lib/utils';
import { LightLine } from './LightLine';
import { MONO } from './ui';

type Tone = 'dark' | 'light' | 'court';

const NUMBER: Record<Tone, string> = { dark: 'text-paper', light: 'text-night', court: 'text-white' };
const SLASH: Record<Tone, string> = { dark: 'text-light', light: 'text-court', court: 'text-light' };
const LABEL: Record<Tone, string> = { dark: 'text-concrete', light: 'text-graphite', court: 'text-white/80' };

/** Building-signage style label: LED accent + "01/" large mono + small text. */
export function SectionLabel({
  number,
  label,
  tone = 'dark',
  className,
}: {
  number: string;
  label: string;
  /** `dark` = on a dark surface, `light` = on paper, `court` = on court blue. */
  tone?: Tone;
  className?: string;
}) {
  return (
    <div className={cn('flex items-baseline gap-4', className)}>
      {/* LED accent only on dark / blue surfaces: warm light disappears on paper. */}
      {tone !== 'light' && <LightLine className="w-8 shrink-0 self-center md:w-12" delay={0.1} />}
      <span className={cn('font-jb text-[clamp(1.5rem,2.4vw,2.25rem)] leading-none', NUMBER[tone])}>
        {number}
        <span className={SLASH[tone]}>/</span>
      </span>
      <span className={cn(MONO, LABEL[tone])}>{label}</span>
    </div>
  );
}
