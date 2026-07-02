import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Trophy, CalendarDays, CheckCircle2, ArrowRight, Zap, Check } from 'lucide-react';
import { motion } from 'framer-motion';
import { getSpaces } from '../api/spaces';
import { SpaceCard, SpaceCardSkeleton } from '../components/SpaceCard';
import { useAuth } from '../hooks/useAuth';

const STEPS = [
  {
    icon: Trophy,
    label: 'PASO 1',
    title: 'Elegí tu espacio',
    desc: 'Explorá las canchas y quinchos disponibles. Si es tu primera vez, te pedimos que te registres gratis.',
  },
  {
    icon: CalendarDays,
    label: 'PASO 2',
    title: 'Seleccioná fecha y turno',
    desc: 'Chequeá la disponibilidad en tiempo real y elegí el horario que más te convenga.',
  },
  {
    icon: CheckCircle2,
    label: 'PASO 3',
    title: 'Confirmá tu reserva',
    desc: 'Recibís la confirmación al instante. ¡Listo para jugar!',
  },
];

const FEATURES = [
  'Estacionamiento',
  'Quinchos con parrilla',
  'Zona comercial',
  'Canchas al aire libre',
  'Canchas techadas',
];

export default function HomePage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const { data: spaces, isLoading } = useQuery({ queryKey: ['spaces'], queryFn: getSpaces });

  return (
    <div className="animate-fade-up">

      {/* ── Hero — único bloque oscuro ── */}
      <section className="relative min-h-[88vh] bg-dark-2 flex items-center overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-dark-2 to-dark pointer-events-none" />
        <div className="absolute top-1/3 right-[10%] w-[400px] h-[400px] rounded-full bg-primary/[0.15] blur-[120px] pointer-events-none" />
        <div className="absolute bottom-1/4 left-[15%] w-[240px] h-[240px] rounded-full bg-white/[0.04] blur-[80px] pointer-events-none" />

        <div className="max-w-[1140px] mx-auto px-6 py-24 relative z-10">
          <motion.div
            initial={{ opacity: 0, y: 32 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, ease: [0.4, 0, 0.2, 1] }}
            className="max-w-2xl mx-auto text-center"
          >
            <span className="inline-block text-2xs font-bold tracking-[0.12em] uppercase text-lime mb-5">
              COMPLEJO DEPORTIVO · VALLE VIEJO
            </span>
            <h1 className="font-serif text-5xl text-white tracking-tight leading-[1.06] mb-6">
              Pádel, fútbol y quinchos a minutos de casa
            </h1>
            <p className="text-sm text-white/60 leading-relaxed mb-10 max-w-md mx-auto">
              Espacios modernos y cómodos para toda la familia. Todo lo que necesitás en un mismo lugar. Reservá tu turno en minutos.
            </p>

            <div className="flex flex-wrap gap-3 justify-center">
              <motion.button
                whileHover={{ scale: 1.03 }}
                whileTap={{ scale: 0.97 }}
                onClick={() => navigate('/canchas')}
                className="bg-primary text-white font-bold rounded-lg px-6 py-3 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:shadow-glow active:scale-[0.98] flex items-center gap-2"
              >
                Reservar ahora <ArrowRight size={15} />
              </motion.button>

              {!isAuthenticated && (
                <motion.button
                  whileHover={{ scale: 1.03 }}
                  whileTap={{ scale: 0.97 }}
                  onClick={() => navigate('/login', { state: { tab: 'register' } })}
                  className="bg-transparent border-[1.5px] border-white/25 text-white/80 font-semibold rounded-lg px-6 py-3 text-sm transition-all hover:border-lime/60 hover:text-white active:scale-[0.98]"
                >
                  Registrarme
                </motion.button>
              )}
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.35, duration: 0.5 }}
            className="flex flex-wrap gap-3 mt-16 pt-10 border-t border-white/[0.10] justify-center"
          >
            {FEATURES.map((f) => (
              <span
                key={f}
                className="inline-flex items-center gap-2 rounded-full px-4 py-2 bg-white/[0.08] border border-white/[0.12] text-sm font-medium text-white/80"
              >
                <Check size={13} className="text-lime shrink-0" />
                {f}
              </span>
            ))}
          </motion.div>
        </div>
      </section>

      {/* ── Spaces grid — fondo claro ── */}
      <section className="bg-bg">
        <div className="max-w-[1140px] mx-auto px-6 py-20">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="mb-10"
          >
            <span className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-1.5 block">
              Disponibles ahora
            </span>
            <h2 className="font-serif text-4xl text-ink tracking-tight">
              Nuestros espacios
            </h2>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {isLoading
              ? Array.from({ length: 3 }).map((_, i) => <SpaceCardSkeleton key={i} />)
              : spaces?.map((space, i) => (
                <SpaceCard key={space.id} space={space} index={i} />
              ))}
          </div>
        </div>
      </section>

      {/* ── How it works — surface claro ── */}
      <section className="bg-surface">
        <div className="max-w-[1140px] mx-auto px-6 py-20">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="mb-12"
          >
            <span className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-1.5 block">
              SIMPLE Y RÁPIDO
            </span>
            <h2 className="font-serif text-4xl text-ink tracking-tight">
              ¿Cómo funciona?
            </h2>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
            {STEPS.map((step, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 24 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.12, duration: 0.45 }}
                className="bg-cream-white rounded-xl border-[1.5px] border-black/[0.07] shadow-sm p-6"
              >
                <div className="w-11 h-11 rounded-lg bg-primary/10 flex items-center justify-center mb-5">
                  <step.icon size={20} className="text-primary" />
                </div>
                <div className="text-2xs font-bold tracking-[0.1em] uppercase text-ink-2 mb-2">
                  {step.label}
                </div>
                <h3 className="text-base font-bold text-ink mb-2">{step.title}</h3>
                <p className="text-sm text-ink-2 leading-relaxed">{step.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA (solo no autenticados) — cream ── */}
      {!isAuthenticated && (
        <section className="bg-cream py-24">
          <div className="max-w-[520px] mx-auto px-6 text-center">
            <motion.div
              initial={{ opacity: 0, scale: 0.96 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.45 }}
            >
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center mx-auto mb-6">
                <Zap size={22} className="text-primary" />
              </div>
              <h2 className="font-serif text-4xl text-ink tracking-tight mb-3">
                ¿Listo para jugar?
              </h2>
              <p className="text-sm text-ink-2 leading-relaxed mb-8">
                Creá tu cuenta gratis y empezá a reservar hoy mismo.
              </p>
              <motion.button
                whileHover={{ scale: 1.03 }}
                whileTap={{ scale: 0.97 }}
                onClick={() => navigate('/login', { state: { tab: 'register' } })}
                className="bg-primary text-white font-bold rounded-lg px-8 py-3.5 text-sm transition-all duration-normal ease-smooth hover:bg-primary-dark hover:shadow-glow active:scale-[0.98]"
              >
                Crear cuenta gratis
              </motion.button>
            </motion.div>
          </div>
        </section>
      )}

      {/* ── Footer — surface suave ── */}
      <footer className="bg-surface border-t border-black/[0.06] py-8 text-center">
        <p className="text-2xs text-ink-2 tracking-wide">
          © {new Date().getFullYear()} H1 Canchas. Todos los derechos reservados.
        </p>
      </footer>
    </div>
  );
}
