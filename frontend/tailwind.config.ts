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
        sm:    "8px",
        md:    "13px",
        lg:    "18px",
        xl:    "24px",
        "2xl": "32px",
        full:  "9999px",
      },
      boxShadow: {
        sm:  "0 1px 4px rgba(28,46,38,.06), 0 2px 8px rgba(28,46,38,.04)",
        md:  "0 4px 16px rgba(28,46,38,.10), 0 1px 4px rgba(28,46,38,.05)",
        lg:  "0 12px 40px rgba(28,46,38,.13), 0 4px 12px rgba(28,46,38,.07)",
        xl:  "0 20px 60px rgba(28,46,38,.15), 0 6px 18px rgba(28,46,38,.08)",
        glow:        "0 0 24px rgba(61,107,80,.40)",
        "glow-lime": "0 0 24px rgba(168,201,127,.35)",
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
