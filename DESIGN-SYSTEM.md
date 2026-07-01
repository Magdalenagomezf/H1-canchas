# H1 Canchas — Design System
> Documento de referencia para Claude Code. Seguir al pie de la letra.

---

## Stack UI (DEFINITIVO)

```bash
# Desinstalar MUI
npm uninstall @mui/material @mui/icons-material @emotion/react @emotion/styled

# Instalar stack nuevo
npm install tailwindcss@latest @tailwindcss/vite
npm install shadcn-ui
npm install framer-motion
npm install lucide-react
npm install clsx tailwind-merge
npx shadcn@latest init
```

**Respuestas al init de shadcn:**
- Style: `Default`
- Base color: `Neutral`
- CSS variables: `Yes`

---

## Configuración de Tailwind (`tailwind.config.ts`)

```ts
import type { Config } from "tailwindcss";

export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
  primary:   { DEFAULT: "#3D6B50", dark: "#2D4A3E", light: "#5A7A52" },
  accent:    { DEFAULT: "#A8C97F" },
  lime:      { DEFAULT: "#A8C97F" },
  bg:        { DEFAULT: "#EEF4EC" },
  surface:   { DEFAULT: "#E4EEE1", 2: "#DDD5C4" },
  cream:     { DEFAULT: "#F5F0E8", white: "#FDFAF5" },
  dark:      { DEFAULT: "#1C2E26", 2: "#2D4A3E", surface: "#2D4A3E" },
  ink:       { DEFAULT: "#1C2E26", 2: "#7B6E5A" },
        status: {
          pending:   "#C8972A",
          confirmed: "#3D7A4E",
          cancelled: "#B24040",
          completed: "#5A5A55",
        },
      },
      fontFamily: {
        sans:  ["Outfit", "sans-serif"],
        serif: ["DM Serif Display", "serif"],
      },
      fontSize: {
        "2xs": ["0.68rem", { lineHeight: "1rem" }],
        xs:    ["0.75rem", { lineHeight: "1.1rem" }],
        sm:    ["0.875rem", { lineHeight: "1.35rem" }],
        base:  ["1rem",    { lineHeight: "1.55rem" }],
        lg:    ["1.125rem",{ lineHeight: "1.4rem"  }],
        xl:    ["1.25rem", { lineHeight: "1.3rem"  }],
        "2xl": ["1.5rem",  { lineHeight: "1.2rem"  }],
        "3xl": ["1.875rem",{ lineHeight: "1.1rem"  }],
        "4xl": ["2.25rem", { lineHeight: "1.05rem" }],
        "5xl": ["3rem",    { lineHeight: "1rem"    }],
      },
      borderRadius: {
        sm:  "8px",
        md:  "13px",
        lg:  "18px",
        xl:  "24px",
        "2xl": "32px",
        full: "9999px",
      },
      boxShadow: {
        sm:  "0 1px 4px rgba(28,28,26,.06), 0 2px 8px rgba(28,28,26,.04)",
        md:  "0 4px 16px rgba(28,28,26,.10), 0 1px 4px rgba(28,28,26,.05)",
        lg:  "0 12px 40px rgba(28,28,26,.13), 0 4px 12px rgba(28,28,26,.07)",
        xl:  "0 20px 60px rgba(28,28,26,.15), 0 6px 18px rgba(28,28,26,.08)",
        glow:"0 0 24px rgba(61,122,78,.35)",
        "glow-lime": "0 0 24px rgba(200,240,89,.3)",
      },
      transitionTimingFunction: {
        spring: "cubic-bezier(.34,1.56,.64,1)",
        smooth: "cubic-bezier(.4,0,.2,1)",
      },
      transitionDuration: {
        fast:   "120ms",
        normal: "180ms",
        slow:   "320ms",
      },
      keyframes: {
        "fade-up": {
          from: { opacity: "0", transform: "translateY(16px)" },
          to:   { opacity: "1", transform: "translateY(0)" },
        },
        "fade-in": {
          from: { opacity: "0" },
          to:   { opacity: "1" },
        },
        "scale-in": {
          from: { opacity: "0", transform: "scale(0.92)" },
          to:   { opacity: "1", transform: "scale(1)" },
        },
        "slide-in-right": {
          from: { transform: "translateX(100%)" },
          to:   { transform: "translateX(0)" },
        },
        "slide-up": {
          from: { opacity: "0", transform: "translateY(24px)" },
          to:   { opacity: "1", transform: "translateY(0)" },
        },
        shimmer: {
          from: { transform: "translateX(-100%)" },
          to:   { transform: "translateX(100%)" },
        },
        pulse: {
          "0%,100%": { opacity: "1", transform: "scale(1)" },
          "50%":     { opacity: "0.6", transform: "scale(0.85)" },
        },
        float: {
          "0%,100%": { transform: "translateY(0px)" },
          "50%":     { transform: "translateY(-6px)" },
        },
      },
      animation: {
        "fade-up":        "fade-up 0.4s ease both",
        "fade-up-slow":   "fade-up 0.6s ease both",
        "fade-in":        "fade-in 0.25s ease both",
        "scale-in":       "scale-in 0.3s cubic-bezier(.34,1.56,.64,1) both",
        "slide-in-right": "slide-in-right 0.3s cubic-bezier(.4,0,.2,1) both",
        "slide-up":       "slide-up 0.28s ease both",
        shimmer:          "shimmer 1.8s ease infinite",
        pulse:            "pulse 2s ease infinite",
        float:            "float 3s ease-in-out infinite",
      },
    },
  },
  plugins: [],
} satisfies Config;
```

---

## CSS Global (`src/index.css`)

```css
@import url('https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800;900&family=DM+Serif+Display:ital@0;1&display=swap');
@import "tailwindcss";

:root {
  --radius: 13px;
}

* { box-sizing: border-box; }

body {
  font-family: 'Outfit', sans-serif;
background-color: #EEF4EC;
  color: #1C2E26;1
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Scrollbar sutil */
::-webkit-scrollbar { width: 5px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #B8A99A; border-radius: 999px; }

/* Selección */
::selection { background: rgba(61,122,78,.2); color: #1C1C1A; }

/* Focus visible accesible */
:focus-visible {
  outline: 2px solid #3D7A4E;
  outline-offset: 2px;
  border-radius: 6px;
}
```

---

## Componentes shadcn a instalar

```bash
npx shadcn@latest add button card input label badge
npx shadcn@latest add dialog drawer sheet
npx shadcn@latest add select textarea
npx shadcn@latest add table
npx shadcn@latest add toast
npx shadcn@latest add calendar
```

---

## Especificación de componentes

### Button

```tsx
// Variantes a implementar

// Primary — verde bosque, shimmer a/l hover
<Button>
  className="relative overflow-hidden bg-primary text-white font-bold
             rounded-lg px-5 py-2.5 text-sm
             transition-all duration-normal ease-smooth
             hover:bg-primary-dark hover:scale-[1.02] hover:shadow-glow
             active:scale-[0.98]
             before:absolute before:inset-0
             before:bg-gradient-to-r before:from-transparent
             before:via-white/10 before:to-transparent
             before:-translate-x-full hover:before:translate-x-full
             before:transition-transform before:duration-500"
/>

// Dark — oscuro con flecha lima
<Button variant="dark">
  className="bg-dark-surface text-white font-bold rounded-lg
             px-5 py-2.5 transition-all hover:scale-[1.02]
             active:scale-[0.98] shadow-md"
/>

// Ghost
<Button variant="ghost">
  className="bg-transparent border-[1.5px] border-surface-2
             text-ink-2 font-semibold rounded-lg px-5 py-2.5
             transition-all hover:border-primary hover:text-primary
             active:scale-[0.98]"
/>

// Destructive
<Button variant="destructive">
  className="bg-status-cancelled/10 border-[1.5px] border-status-cancelled/20
             text-status-cancelled font-semibold rounded-lg px-5 py-2.5
             transition-all hover:bg-status-cancelled/18 active:scale-[0.98]"
/>
```

**Regla:** TODOS los botones tienen `active:scale-[0.98]`. Sin excepciones.

---

### Card

```tsx
// Base card
className="bg-white rounded-xl border-[1.5px] border-black/[0.07]
           shadow-sm transition-all duration-normal ease-smooth
           hover:-translate-y-1 hover:shadow-md"

// Dark card (panel lateral, headers)
className="bg-dark-surface rounded-xl overflow-hidden"

// Surface card (secundaria)
className="bg-surface rounded-xl border-[1.5px] border-black/[0.07]"
```

---

### Input / TextField

```tsx
className="w-full px-3.5 py-3 bg-surface
           border-[1.5px] border-black/10 rounded-lg
           text-sm font-medium text-ink placeholder:text-ink-2/60
           outline-none transition-all duration-normal
           focus:border-primary focus:ring-2 focus:ring-primary/[0.13]"
```

---

### Badge / Status chip

```tsx
// Pending
className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1
           bg-status-pending/13 text-status-pending
           text-2xs font-bold uppercase tracking-wide"

// Confirmed
className="... bg-status-confirmed/13 text-status-confirmed ..."

// Cancelled
className="... bg-status-cancelled/11 text-status-cancelled ..."

// Completed
className="... bg-status-completed/11 text-status-completed ..."
```

---

### Navbar

```tsx
className="fixed top-0 inset-x-0 z-50 h-[58px]
           bg-dark-surface border-b border-white/[0.07]
           flex items-center justify-between px-6
           shadow-[0_2px_16px_rgba(0,0,0,0.2)]"

// Logo
className="font-serif text-xl text-white"
// Logo span (acento)
className="text-lime italic"
```

---

### Section headers con eyebrow

```tsx
// Eyebrow label
className="text-2xs font-bold tracking-[0.1em] uppercase text-primary mb-1.5"

// Title principal
className="font-serif text-4xl text-ink tracking-tight leading-[1.08]"

// Title con acento
<>Reservá tu <em className="text-primary not-italic">turno</em></>

// Subtitle
className="text-sm text-ink-2 leading-relaxed mt-2 max-w-md"
```

---

## Animaciones por contexto

### Entrada de páginas
```tsx
// Wrapper de cada página
<div className="animate-fade-up">
  {children}
</div>
```

### Cards en lista (stagger)
```tsx
{items.map((item, i) => (
  <div
    key={item.id}
    className="animate-fade-up"
    style={{ animationDelay: `${i * 60}ms` }}
  >
    <ItemCard item={item} />
  </div>
))}
```

### Modales y drawers
```tsx
// Modal: scale-in desde centro
<div className="animate-scale-in">

// Drawer: slide desde derecha
<div className="animate-slide-in-right">

// Bottom sheet: slide desde abajo
<div className="animate-slide-up">
```

### Hover en space cards
```tsx
className="transition-all duration-[200ms] ease-smooth
           hover:-translate-y-[5px] hover:shadow-lg"
```

---

## Utilidades de layout frecuentes

```tsx
// Contenedor de página
className="max-w-[1140px] mx-auto px-6 py-10"

// Grid de espacios
className="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-5"

// Grid 2 columnas (main + panel)
className="grid grid-cols-[1fr_360px] gap-6 items-start"

// Sidebar layout
className="flex min-h-screen"
// Sidebar
className="w-[220px] shrink-0 bg-dark-surface flex flex-col sticky top-0 h-screen"
```

---

## Tipografía — Reglas

| Elemento | Clase |
|---|---|
| Hero title | `font-serif text-5xl text-white tracking-tight` |
| Page title | `font-serif text-4xl text-ink tracking-tight` |
| Section title | `font-serif text-3xl text-ink tracking-tight` |
| Card title | `text-base font-bold text-ink` |
| Body | `text-sm text-ink-2 leading-relaxed` |
| Label | `text-2xs font-bold tracking-[0.06em] uppercase text-ink-2` |
| Price | `font-serif text-3xl text-ink` |

**Regla:** `font-serif` solo para títulos y precios. Todo lo demás es `font-sans` (Outfit).

---

## Patrones de color por sección

| Sección | Fondo | Cards | Texto |
|---|---|---|---|
| Hero / Headers | `bg-dark-surface` | glass con `bg-white/10` | `text-white` |
| Páginas principales | `bg-bg` | `bg-white` | `text-ink` |
| Secciones "How it works" | `bg-dark-surface` | — | `text-white` |
| Paneles admin/recep | `bg-bg` | `bg-white` | `text-ink` |
| Sidebar | `bg-dark-surface` | — | `text-white/50 → text-white` |

---

## Helper `cn()` (OBLIGATORIO usar en todos los componentes)

```ts
// src/lib/utils.ts
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

---

## Reglas que NO se pueden romper

1. **NUNCA** usar colores hardcodeados como `#3D7A4E` en el JSX — siempre usar las clases de Tailwind del tema (`text-primary`, `bg-primary`, etc.)
2. **NUNCA** usar `style={{ color: "..." }}` para colores del sistema
3. **TODOS** los botones tienen `active:scale-[0.98]` y `transition-all`
4. **NUNCA** usar `font-bold uppercase` para botones — los botones van en `font-bold` sin uppercase (excepto badges/chips)
5. **NUNCA** usar sombras de Tailwind genéricas (`shadow-md` de Tailwind default) — usar las del sistema: `shadow-sm`, `shadow-md`, `shadow-lg`, `shadow-xl` definidas en el config
6. Las **transiciones** siempre con `duration-normal ease-smooth` salvo que el efecto lo requiera diferente
7. **SIEMPRE** usar `cn()` para combinar clases condicionales
8. Los **radios** siguen la escala: `rounded-sm` (8px) → `rounded-md` (13px) → `rounded-lg` (18px) → `rounded-xl` (24px) → `rounded-2xl` (32px)
9. **NUNCA** usar `rounded` o `rounded-full` para cards o inputs — solo para pills/chips/avatars
10. Todo texto secundario usa `text-ink-2`, nunca `text-gray-500` ni similares

---

## Referencia visual de estilo

El diseño sigue estas referencias:
- **Fondo oscuro + tarjetas bento:** fondo `dark-surface`, tarjetas con métricas en verde lima `#C8F059`, verde bosque y glass. Tipografía bold. Contrastes fuertes pero sin colores chillones.
- **Hero con foto + bottom sheet:** imagen de fondo con overlay degradado, contenido flotando sobre la imagen, formularios en panel blanco/surface limpio.
- **Estética general:** minimalista-deportiva. Espaciado generoso. Sin bordes duros. Sombras suaves. Animaciones de escala en hover/click. Transiciones suaves entre páginas.

---

## Orden de implementación recomendado

1. Instalar dependencias y configurar Tailwind
2. Crear `tailwind.config.ts` con este archivo
3. Actualizar `src/index.css`
4. Crear `src/lib/utils.ts` con `cn()`
5. Instalar componentes shadcn
6. Actualizar componentes de UI base (Button, Card, Input, Badge)
7. Migrar página por página: Home → Login → Detalle → Checkout → Mis Reservas → Paneles

