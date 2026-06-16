import { createTheme } from '@mui/material/styles';

declare module '@mui/material/styles' {
  interface Palette {
    surface: Palette['primary'];
  }
  interface PaletteOptions {
    surface?: PaletteOptions['primary'];
  }
}

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#3D7A4E',
      dark: '#2D5A3D',
      light: '#5A9E6A',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#7B5E3A',
      dark: '#5C4429',
      light: '#A07D55',
      contrastText: '#FFFFFF',
    },
    background: {
      default: '#EFEFEC',
      paper: '#E5E3DF',
    },
    surface: {
      main: '#2A2A28',
      light: '#3A3A38',
      dark: '#1A1A18',
      contrastText: '#FFFFFF',
    },
    text: {
      primary: '#1A1A18',
      secondary: '#5A5A58',
    },
  },
  typography: {
    fontFamily: '"Outfit", "Roboto", "Helvetica", "Arial", sans-serif',
    h1: {
      fontFamily: '"DM Serif Display", serif',
      fontWeight: 400,
    },
    h2: {
      fontFamily: '"DM Serif Display", serif',
      fontWeight: 400,
    },
    h3: {
      fontFamily: '"DM Serif Display", serif',
      fontWeight: 400,
    },
    h4: {
      fontFamily: '"DM Serif Display", serif',
      fontWeight: 400,
    },
    h5: {
      fontFamily: '"Outfit", sans-serif',
      fontWeight: 600,
    },
    h6: {
      fontFamily: '"Outfit", sans-serif',
      fontWeight: 600,
    },
    button: {
      fontFamily: '"Outfit", sans-serif',
      fontWeight: 600,
      textTransform: 'none',
      letterSpacing: '0.02em',
    },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          padding: '10px 24px',
        },
        containedPrimary: {
          background: 'linear-gradient(135deg, #3D7A4E, #2D5A3D)',
          '&:hover': {
            background: 'linear-gradient(135deg, #4A8F5E, #3D7A4E)',
          },
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 16,
          boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 8,
        },
      },
    },
  },
});

export default theme;
