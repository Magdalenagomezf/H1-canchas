import { useState } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import type { Location } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import VisibilityIcon from '@mui/icons-material/Visibility';
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff';
import { motion } from 'framer-motion';
import { login, register } from '../api/auth';
import { useAuth } from '../hooks/useAuth';

const MotionCard = motion(Card);

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { setAuth } = useAuth();

  const state = location.state as { from?: Location; tab?: string } | null;
  const redirectTo = state?.from ? state.from.pathname + (state.from.search ?? '') : '/';

  const [tab, setTab] = useState<0 | 1>(state?.tab === 'register' ? 1 : 0);
  const [showPassword, setShowPassword] = useState(false);

  const [loginForm, setLoginForm] = useState({ phone: '', password: '' });
  const [registerForm, setRegisterForm] = useState({ name: '', phone: '', password: '' });

  const loginMutation = useMutation({
    mutationFn: () => login(loginForm.phone, loginForm.password),
    onSuccess: (data) => {
      setAuth(data.user, data.token);
      navigate(redirectTo, { replace: true });
    },
  });

  const registerMutation = useMutation({
    mutationFn: () => register(registerForm.name, registerForm.phone, registerForm.password),
    onSuccess: (data) => {
      setAuth(data.user, data.token);
      navigate(redirectTo, { replace: true });
    },
  });

  const isLogin = tab === 0;
  const mutation = isLogin ? loginMutation : registerMutation;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate();
  };

  return (
    <Box
      sx={{
        minHeight: 'calc(100vh - 64px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(160deg, #2D5A3D 0%, #3D7A4E 50%, #5A9E6A 100%)',
        p: 2,
      }}
    >
      <MotionCard
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.45, ease: 'easeOut' }}
        sx={{ width: '100%', maxWidth: 420, overflow: 'visible' }}
      >
        <CardContent sx={{ p: { xs: 3, sm: 4 } }}>
          {/* Logo */}
          <Typography
            component={Link}
            to="/"
            variant="h4"
            sx={{
              display: 'block',
              textAlign: 'center',
              textDecoration: 'none',
              color: 'primary.dark',
              mb: 3,
            }}
          >
            H1 Canchas
          </Typography>

          {/* Tabs */}
          <Tabs
            value={tab}
            onChange={(_, v) => {
              setTab(v);
              mutation.reset();
            }}
            variant="fullWidth"
            sx={{ mb: 3, '& .MuiTab-root': { fontWeight: 600 } }}
          >
            <Tab label="Iniciar sesión" />
            <Tab label="Registrarse" />
          </Tabs>

          {/* Error */}
          {mutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {(mutation.error as Error).message}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit} sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {/* Name — solo en registro */}
            {!isLogin && (
              <TextField
                label="Nombre"
                value={registerForm.name}
                onChange={(e) => setRegisterForm((f) => ({ ...f, name: e.target.value }))}
                required
                autoFocus
                fullWidth
              />
            )}

            {/* Phone */}
            <TextField
              label="Teléfono"
              type="tel"
              value={isLogin ? loginForm.phone : registerForm.phone}
              onChange={(e) =>
                isLogin
                  ? setLoginForm((f) => ({ ...f, phone: e.target.value }))
                  : setRegisterForm((f) => ({ ...f, phone: e.target.value }))
              }
              required
              autoFocus={isLogin}
              fullWidth
            />

            {/* Password */}
            <TextField
              label="Contraseña"
              type={showPassword ? 'text' : 'password'}
              value={isLogin ? loginForm.password : registerForm.password}
              onChange={(e) =>
                isLogin
                  ? setLoginForm((f) => ({ ...f, password: e.target.value }))
                  : setRegisterForm((f) => ({ ...f, password: e.target.value }))
              }
              required
              fullWidth
              slotProps={{
                input: {
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton onClick={() => setShowPassword((v) => !v)} edge="end">
                        {showPassword ? <VisibilityOffIcon /> : <VisibilityIcon />}
                      </IconButton>
                    </InputAdornment>
                  ),
                },
              }}
            />

            <Button
              type="submit"
              variant="contained"
              size="large"
              disabled={mutation.isPending}
              fullWidth
              sx={{ mt: 1 }}
            >
              {mutation.isPending
                ? isLogin
                  ? 'Ingresando...'
                  : 'Creando cuenta...'
                : isLogin
                  ? 'Iniciar sesión'
                  : 'Crear cuenta'}
            </Button>
          </Box>
        </CardContent>
      </MotionCard>
    </Box>
  );
}
