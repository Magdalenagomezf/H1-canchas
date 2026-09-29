/**
 * All editable landing copy and placeholders live here.
 * Anything marked TODO must be confirmed / filled by the client.
 */

// TODO: drop the two architecture photos into frontend/public/images/
export const IMAGES = {
  hero: '/images/fachada-noche.jpg',
  complex: '/images/locales.jpg',
} as const;

// Hero title. Option A is live; option B is kept for quick swapping.
export const HERO_TITLE = ['ESPACIO', 'DEPORTIVO', 'H1'];
// export const HERO_TITLE = ['JUGÁ', 'BAJO LA', 'LUZ'];

export const HERO_SUBLINE = 'Pádel, fútbol, padbol, beach vóley y quinchos en Valle Viejo, Catamarca.';
export const HERO_HOURS = 'ABIERTO TODOS LOS DÍAS — 12:00 / 24:00';

// Section 01 slideshow, rotated every PLACE_SLIDE_MS
export const PLACE_IMAGES = [
  { src: '/images/lugar-1.jpeg', alt: 'Ingreso al paseo comercial del complejo al atardecer, con árboles iluminados' },
  { src: '/images/lugar-2.jpeg', alt: 'Estacionamiento y galería de locales del complejo al atardecer' },
  { src: '/images/lugar-3.jpeg', alt: 'Galería peatonal entre los locales y el estacionamiento cubierto, de noche' },
];
export const PLACE_SLIDE_MS = 3500;

export const PLACE_STATEMENT = 'Canchas, quinchos y un espacio pensado para que el partido no termine en la cancha';

// TODO: confirm the numbers (padel courts, quinchos count).
export const FACTS = [
  { value: '04', label: 'Canchas de pádel' },
  { value: '01', label: 'Cancha de fútbol' },
  { value: '02', label: 'Quinchos' },
  { value: '12—24', label: 'Horas, todos los días' },
];

export const COMPLEX_FEATURES = [
  'Estacionamiento',
  'Zona comercial',
  'Canchas techadas',
  'Quinchos con parrilla',
];

export const STEPS = [
  { number: '01', title: 'Elegí el espacio', text: 'Canchas o quincho: el que mejor se adapte a tu plan.' },
  { number: '02', title: 'Elegí fecha y turno', text: 'Mirá la disponibilidad en tiempo real y tomá tu horario.' },
  { number: '03', title: 'Pagá con Mercado Pago', text: 'Señá tu turno online y recibí la confirmación al instante.' },
];

export const SPACE_TYPE_LABEL = {
  cancha_padel: 'PÁDEL',
  cancha_futbol: 'FÚTBOL',
  cancha_padbol: 'PADBOL',
  cancha_beach_voley: 'BEACH VÓLEY',
  quincho: 'QUINCHO',
} as const;

export const NAV_LINKS = [
  { label: 'Espacios', href: '#espacios' },
  { label: 'El complejo', href: '#complejo' },
  { label: 'Ubicación', href: '#ubicacion' },
];

export const FINAL_CTA_TITLE = ['RESERVÁ', 'TU TURNO'];

export const FOOTER = {
  // TODO: add street and number.
  address: 'Valle Viejo, Catamarca',
  // TODO: replace with the real Google Maps link.
  mapsUrl: 'https://www.google.com/maps/search/?api=1&query=Valle+Viejo+Catamarca',
  // TODO: replace with the real WhatsApp number (country code, no symbols).
  whatsappUrl: 'https://wa.me/5490000000000',
  whatsappLabel: 'Escribinos por WhatsApp',
  hours: ['Todos los días', '12:00 — 24:00'],
  credit: '© 2026 H1 ESPACIO DEPORTIVO — ARQUITECTURA: GRUPO MAZZUCCO ARQUITECTOS ASOCIADOS',
};
