import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Skeleton from '@mui/material/Skeleton';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Divider from '@mui/material/Divider';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import InputAdornment from '@mui/material/InputAdornment';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import PersonIcon from '@mui/icons-material/Person';
import PhoneIcon from '@mui/icons-material/Phone';
import AddIcon from '@mui/icons-material/Add';
import EventBusyIcon from '@mui/icons-material/EventBusy';
import GroupIcon from '@mui/icons-material/Group';
import SportsTennisIcon from '@mui/icons-material/SportsTennis';
import SportsSoccerIcon from '@mui/icons-material/SportsSoccer';
import CelebrationIcon from '@mui/icons-material/Celebration';
import StorefrontIcon from '@mui/icons-material/Storefront';
import { motion } from 'framer-motion';
import { getAllBookings, createManualBooking, cancelBooking } from '../api/bookings';
import { getSpaces, getSlots, createSpace, updateSpace, deleteSpace } from '../api/spaces';
import type { Space } from '../types';
import { getUsers, createStaffUser, updateUserRole, deleteUser } from '../api/users';
import { useAuth } from '../hooks/useAuth';
import type { BookingStatus, SpaceType } from '../types';

const MotionCard = motion(Card);

const STATUS_CONFIG: Record<BookingStatus, { label: string; color: 'warning' | 'success' | 'default' | 'info' }> = {
  pending: { label: 'Pendiente', color: 'warning' },
  confirmed: { label: 'Confirmada', color: 'success' },
  cancelled: { label: 'Cancelada', color: 'default' },
  completed: { label: 'Completada', color: 'info' },
};

const ROLE_LABELS: Record<string, string> = {
  customer: 'Cliente',
  receptionist: 'Recepcionista',
  admin: 'Admin',
};

const ROLE_COLORS: Record<string, 'default' | 'primary' | 'secondary'> = {
  customer: 'default',
  receptionist: 'primary',
  admin: 'secondary',
};

const SPACE_TYPE_OPTIONS: { value: SpaceType; label: string; icon: React.ReactNode }[] = [
  { value: 'cancha_padel', label: 'Pádel', icon: <SportsTennisIcon fontSize="small" /> },
  { value: 'cancha_futbol', label: 'Fútbol', icon: <SportsSoccerIcon fontSize="small" /> },
  { value: 'quincho', label: 'Quincho / Salón', icon: <CelebrationIcon fontSize="small" /> },
];

const today = new Date().toISOString().split('T')[0];

function formatDate(dateStr: string): string {
  const [year, month, day] = dateStr.split('-').map(Number);
  return new Date(year, month - 1, day).toLocaleDateString('es-AR', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

const EMPTY_BOOKING_FORM = {
  space_id: '',
  slot_id: '',
  booking_date: today,
  customer_name: '',
  customer_phone: '',
};

const EMPTY_SPACE_FORM = {
  name: '',
  type: '' as SpaceType | '',
  description: '',
  price_per_slot: '',
};

const EMPTY_STAFF_FORM = {
  name: '',
  phone: '',
  password: '',
  role: 'receptionist' as 'receptionist' | 'admin',
};

export default function StaffPanelPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState(0);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
    open: false,
    message: '',
    severity: 'success',
  });

  const showSnack = (message: string, severity: 'success' | 'error' = 'success') =>
    setSnackbar({ open: true, message, severity });

  // — Bookings tab state —
  const [date, setDate] = useState(today);
  const [bookingDialogOpen, setBookingDialogOpen] = useState(false);
  const [bookingForm, setBookingForm] = useState(EMPTY_BOOKING_FORM);

  // — Spaces tab state —
  const [spaceDialogOpen, setSpaceDialogOpen] = useState(false);
  const [spaceForm, setSpaceForm] = useState(EMPTY_SPACE_FORM);
  const [editingSpace, setEditingSpace] = useState<Space | null>(null);
  const [editForm, setEditForm] = useState({ name: '', description: '', price_per_slot: '' });

  // — Users tab state —
  const [staffDialogOpen, setStaffDialogOpen] = useState(false);
  const [staffForm, setStaffForm] = useState(EMPTY_STAFF_FORM);
  const [updatingRoleId, setUpdatingRoleId] = useState<number | null>(null);

  // — Delete confirmation state —
  const [deleteConfirm, setDeleteConfirm] = useState<{ type: 'user' | 'space'; id: number; name: string } | null>(null);

  // — Cancel booking confirmation state —
  const [cancelConfirm, setCancelConfirm] = useState<{ id: number; customer: string } | null>(null);

  // — Queries —
  const { data: bookings, isLoading: loadingBookings } = useQuery({
    queryKey: ['all-bookings', date],
    queryFn: () => getAllBookings(date),
  });

  const { data: allSpaces, isLoading: loadingSpaces } = useQuery({
    queryKey: ['spaces'],
    queryFn: getSpaces,
  });

  const { data: slots } = useQuery({
    queryKey: ['slots', bookingForm.space_id],
    queryFn: () => getSlots(Number(bookingForm.space_id)),
    enabled: !!bookingForm.space_id,
  });

  const { data: users, isLoading: loadingUsers } = useQuery({
    queryKey: ['admin-users'],
    queryFn: getUsers,
    enabled: isAdmin && activeTab === 2,
  });

  // — Mutations —
  const manualBookingMutation = useMutation({
    mutationFn: createManualBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['all-bookings'] });
      setBookingDialogOpen(false);
      setBookingForm(EMPTY_BOOKING_FORM);
      showSnack('Reserva manual creada correctamente.');
    },
    onError: (err: Error) => showSnack(err.message ?? 'No se pudo crear la reserva.', 'error'),
  });

  const updateSpaceMutation = useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: { name: string; description?: string; price_per_slot: number } }) =>
      updateSpace(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['spaces'] });
      setEditingSpace(null);
      showSnack('Espacio actualizado correctamente.');
    },
    onError: (err: Error) => showSnack(err.message ?? 'No se pudo actualizar el espacio.', 'error'),
  });

  const createSpaceMutation = useMutation({
    mutationFn: createSpace,
    onSuccess: (newSpace) => {
      queryClient.invalidateQueries({ queryKey: ['spaces'] });
      setSpaceDialogOpen(false);
      setSpaceForm(EMPTY_SPACE_FORM);
      showSnack(`Espacio "${newSpace.name}" creado correctamente.`);
    },
    onError: (err: Error) => showSnack(err.message ?? 'No se pudo crear el espacio.', 'error'),
  });

  const createStaffMutation = useMutation({
    mutationFn: createStaffUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setStaffDialogOpen(false);
      setStaffForm(EMPTY_STAFF_FORM);
      showSnack('Usuario creado correctamente.');
    },
    onError: (err: Error) => showSnack(err.message ?? 'No se pudo crear el usuario.', 'error'),
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ id, role }: { id: number; role: string }) => updateUserRole(id, role),
    onMutate: ({ id }) => setUpdatingRoleId(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      showSnack('Rol actualizado.');
    },
    onError: (err: Error) => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      showSnack(err.message ?? 'No se pudo cambiar el rol.', 'error');
    },
    onSettled: () => setUpdatingRoleId(null),
  });

  const deleteSpaceMutation = useMutation({
    mutationFn: deleteSpace,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['spaces'] });
      setDeleteConfirm(null);
      showSnack('Espacio eliminado correctamente.');
    },
    onError: (err: Error) => {
      setDeleteConfirm(null);
      showSnack(err.message ?? 'No se pudo eliminar el espacio.', 'error');
    },
  });

  const deleteUserMutation = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      setDeleteConfirm(null);
      showSnack('Usuario eliminado correctamente.');
    },
    onError: (err: Error) => {
      setDeleteConfirm(null);
      showSnack(err.message ?? 'No se pudo eliminar el usuario.', 'error');
    },
  });

  const cancelBookingMutation = useMutation({
    mutationFn: (id: number) => cancelBooking(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['all-bookings'] });
      setCancelConfirm(null);
      showSnack('Reserva cancelada correctamente.');
    },
    onError: (err: Error) => {
      setCancelConfirm(null);
      showSnack(err.message ?? 'No se pudo cancelar la reserva.', 'error');
    },
  });

  const handleBookingSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    manualBookingMutation.mutate({
      space_id: Number(bookingForm.space_id),
      slot_id: Number(bookingForm.slot_id),
      booking_date: bookingForm.booking_date,
      customer_name: bookingForm.customer_name,
      customer_phone: bookingForm.customer_phone,
    });
  };

  const handleEditOpen = (space: Space) => {
    setEditingSpace(space);
    setEditForm({
      name: space.name,
      description: space.description ?? '',
      price_per_slot: String(space.price_per_slot),
    });
  };

  const handleEditSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingSpace) return;
    const price = parseFloat(editForm.price_per_slot);
    if (isNaN(price) || price <= 0) return;
    updateSpaceMutation.mutate({
      id: editingSpace.id,
      payload: {
        name: editForm.name,
        description: editForm.description || undefined,
        price_per_slot: price,
      },
    });
  };

  const handleSpaceSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const price = parseFloat(spaceForm.price_per_slot);
    if (!spaceForm.type || isNaN(price) || price <= 0) return;
    createSpaceMutation.mutate({
      name: spaceForm.name,
      type: spaceForm.type as SpaceType,
      description: spaceForm.description || undefined,
      price_per_slot: price,
    });
  };

  const handleStaffSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createStaffMutation.mutate(staffForm);
  };

  return (
    <Box>
      {/* Header */}
      <Box sx={{ background: 'linear-gradient(160deg, #2D5A3D 0%, #3D7A4E 60%, #5A9E6A 100%)', py: { xs: 6, md: 8 } }}>
        <Container maxWidth="lg">
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>
            <Typography variant="h2" sx={{ color: '#fff', fontSize: { xs: '2rem', md: '3rem' }, mb: 1 }}>
              Panel de staff
            </Typography>
            <Typography sx={{ color: 'rgba(255,255,255,0.8)' }}>
              Gestioná reservas{isAdmin ? ', espacios y usuarios' : ''} desde acá.
            </Typography>
          </motion.div>
        </Container>
      </Box>

      <Container maxWidth="lg" sx={{ py: { xs: 5, md: 8 } }}>
        {/* Tabs */}
        <Tabs
          value={activeTab}
          onChange={(_, v) => setActiveTab(v)}
          sx={{ mb: 5, borderBottom: '1px solid', borderColor: 'divider' }}
        >
          <Tab icon={<CalendarMonthIcon />} iconPosition="start" label="Reservas" />
          {isAdmin && <Tab icon={<StorefrontIcon />} iconPosition="start" label="Espacios" />}
          {isAdmin && <Tab icon={<GroupIcon />} iconPosition="start" label="Usuarios" />}
        </Tabs>

        {/* — TAB 0: RESERVAS — */}
        {activeTab === 0 && (
          <Box>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 4, flexWrap: 'wrap', gap: 2 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flexWrap: 'wrap' }}>
                <TextField
                  type="date"
                  label="Fecha"
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                  size="small"
                  slotProps={{ inputLabel: { shrink: true } }}
                  sx={{ maxWidth: 200 }}
                />
                {date && (
                  <Typography variant="body2" color="text.secondary">
                    {formatDate(date)}
                  </Typography>
                )}
              </Box>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => {
                  setBookingForm({ ...EMPTY_BOOKING_FORM, booking_date: date });
                  setBookingDialogOpen(true);
                }}
              >
                Nueva reserva manual
              </Button>
            </Box>

            {loadingBookings ? (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                {Array.from({ length: 3 }).map((_, i) => (
                  <Card key={i}>
                    <CardContent sx={{ p: 3 }}>
                      <Skeleton height={32} width="40%" sx={{ mb: 2 }} />
                      <Skeleton height={20} width="60%" sx={{ mb: 1 }} />
                      <Skeleton height={20} width="50%" />
                    </CardContent>
                  </Card>
                ))}
              </Box>
            ) : bookings?.length === 0 ? (
              <Box sx={{ textAlign: 'center', py: 10 }}>
                <EventBusyIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 2 }} />
                <Typography variant="h5" color="text.secondary" sx={{ mb: 1 }}>
                  No hay reservas para esta fecha
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Podés cargar una reserva manual con el botón de arriba.
                </Typography>
              </Box>
            ) : (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                {bookings?.map((booking, i) => {
                  const status = STATUS_CONFIG[booking.status];
                  return (
                    <MotionCard
                      key={booking.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.05 }}
                    >
                      <CardContent sx={{ p: { xs: 2.5, sm: 3 } }}>
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2, gap: 2, flexWrap: 'wrap' }}>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                            <Typography variant="h6">{booking.space.name}</Typography>
                            {booking.customer.is_manual && (
                              <Chip label="Manual" size="small" variant="outlined" color="secondary" sx={{ fontWeight: 600 }} />
                            )}
                          </Box>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Chip label={status.label} color={status.color} size="small" sx={{ fontWeight: 600 }} />
                            {(booking.status === 'pending' || booking.status === 'confirmed') && (
                              <Button
                                size="small"
                                variant="outlined"
                                color="error"
                                onClick={() => setCancelConfirm({ id: booking.id, customer: booking.customer.name })}
                              >
                                Cancelar
                              </Button>
                            )}
                          </Box>
                        </Box>
                        <Divider sx={{ mb: 2 }} />
                        <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', mb: 2 }}>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                            <AccessTimeIcon fontSize="small" color="action" />
                            <Typography variant="body2" color="text.secondary">
                              {booking.slot.label}
                              {booking.slot.start_time && booking.slot.end_time
                                ? ` (${booking.slot.start_time} – ${booking.slot.end_time})`
                                : ''}
                            </Typography>
                          </Box>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                            <PersonIcon fontSize="small" color="action" />
                            <Typography variant="body2" color="text.secondary">{booking.customer.name}</Typography>
                          </Box>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                            <PhoneIcon fontSize="small" color="action" />
                            <Typography variant="body2" color="text.secondary">{booking.customer.phone}</Typography>
                          </Box>
                        </Box>
                        <Typography variant="h6" color="primary.main">
                          ${booking.total_price.toLocaleString('es-AR')}
                        </Typography>
                      </CardContent>
                    </MotionCard>
                  );
                })}
              </Box>
            )}
          </Box>
        )}

        {/* — TAB 1: ESPACIOS (admin only) — */}
        {activeTab === 1 && isAdmin && (
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 4 }}>
              <Button variant="contained" startIcon={<AddIcon />} onClick={() => setSpaceDialogOpen(true)}>
                Crear espacio
              </Button>
            </Box>

            {loadingSpaces ? (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                {Array.from({ length: 3 }).map((_, i) => (
                  <Card key={i}>
                    <CardContent sx={{ p: 3 }}>
                      <Skeleton height={32} width="50%" sx={{ mb: 1 }} />
                      <Skeleton height={20} width="30%" />
                    </CardContent>
                  </Card>
                ))}
              </Box>
            ) : allSpaces?.length === 0 ? (
              <Box sx={{ textAlign: 'center', py: 10 }}>
                <StorefrontIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 2 }} />
                <Typography variant="h5" color="text.secondary">
                  Todavía no hay espacios creados
                </Typography>
              </Box>
            ) : (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                {allSpaces?.map((space, i) => {
                  const typeOption = SPACE_TYPE_OPTIONS.find((o) => o.value === space.type);
                  return (
                    <MotionCard
                      key={space.id}
                      initial={{ opacity: 0, y: 16 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.05 }}
                    >
                      <CardContent sx={{ p: { xs: 2.5, sm: 3 }, display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 2, flexWrap: 'wrap' }}>
                        <Box>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                            <Typography variant="h6">{space.name}</Typography>
                            <Chip
                              icon={typeOption?.icon as React.ReactElement}
                              label={typeOption?.label ?? space.type}
                              size="small"
                              sx={{ bgcolor: 'primary.main', color: '#fff', '& .MuiChip-icon': { color: '#fff' } }}
                            />
                          </Box>
                          {space.description && (
                            <Typography variant="body2" color="text.secondary">{space.description}</Typography>
                          )}
                        </Box>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                          <Typography variant="h6" color="primary.main">
                            ${space.price_per_slot.toLocaleString('es-AR')}
                            <Typography component="span" variant="body2" color="text.secondary" sx={{ ml: 0.5 }}>
                              / turno
                            </Typography>
                          </Typography>
                          <Button size="small" variant="outlined" onClick={() => handleEditOpen(space)}>
                            Editar
                          </Button>
                          <Button
                            size="small"
                            variant="outlined"
                            color="error"
                            onClick={() => setDeleteConfirm({ type: 'space', id: space.id, name: space.name })}
                          >
                            Eliminar
                          </Button>
                        </Box>
                      </CardContent>
                    </MotionCard>
                  );
                })}
              </Box>
            )}
          </Box>
        )}

        {/* — TAB 2: USUARIOS (admin only) — */}
        {activeTab === 2 && isAdmin && (
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 4 }}>
              <Button variant="contained" startIcon={<AddIcon />} onClick={() => setStaffDialogOpen(true)}>
                Nuevo usuario staff
              </Button>
            </Box>

            {loadingUsers ? (
              <Card>
                <CardContent>
                  {Array.from({ length: 4 }).map((_, i) => (
                    <Skeleton key={i} height={48} sx={{ mb: 1 }} />
                  ))}
                </CardContent>
              </Card>
            ) : (
              <Card>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Nombre</TableCell>
                      <TableCell>Teléfono</TableCell>
                      <TableCell>Rol actual</TableCell>
                      <TableCell>Cambiar rol</TableCell>
                      <TableCell />
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {users?.map((u) => (
                      <TableRow key={u.id}>
                        <TableCell>{u.name}</TableCell>
                        <TableCell>{u.phone}</TableCell>
                        <TableCell>
                          <Chip
                            label={ROLE_LABELS[u.role] ?? u.role}
                            color={ROLE_COLORS[u.role] ?? 'default'}
                            size="small"
                            sx={{ fontWeight: 600 }}
                          />
                        </TableCell>
                        <TableCell>
                          <FormControl size="small" disabled={u.id === user?.id || updatingRoleId === u.id}>
                            <Select
                              value={u.role}
                              onChange={(e) => updateRoleMutation.mutate({ id: u.id, role: e.target.value })}
                              sx={{ minWidth: 150 }}
                            >
                              <MenuItem value="customer">Cliente</MenuItem>
                              <MenuItem value="receptionist">Recepcionista</MenuItem>
                              <MenuItem value="admin">Admin</MenuItem>
                            </Select>
                          </FormControl>
                          {u.id === user?.id && (
                            <Typography variant="caption" color="text.disabled" sx={{ ml: 1 }}>
                              (tu cuenta)
                            </Typography>
                          )}
                        </TableCell>
                        <TableCell align="right">
                          {u.id !== user?.id && (
                            <Button
                              size="small"
                              variant="outlined"
                              color="error"
                              onClick={() => setDeleteConfirm({ type: 'user', id: u.id, name: u.name })}
                            >
                              Eliminar
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </Card>
            )}
          </Box>
        )}
      </Container>

      {/* Dialog: reserva manual */}
      <Dialog open={bookingDialogOpen} onClose={() => setBookingDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Nueva reserva manual</DialogTitle>
        <Box component="form" onSubmit={handleBookingSubmit}>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2.5, pt: 1 }}>
            <TextField
              select
              label="Espacio"
              value={bookingForm.space_id}
              onChange={(e) => setBookingForm((f) => ({ ...f, space_id: e.target.value, slot_id: '' }))}
              required
              fullWidth
            >
              {allSpaces?.map((s) => (
                <MenuItem key={s.id} value={String(s.id)}>{s.name}</MenuItem>
              ))}
            </TextField>

            <TextField
              select
              label="Turno"
              value={bookingForm.slot_id}
              onChange={(e) => setBookingForm((f) => ({ ...f, slot_id: e.target.value }))}
              required
              fullWidth
              disabled={!bookingForm.space_id}
            >
              {slots?.map((s) => (
                <MenuItem key={s.id} value={String(s.id)}>
                  {s.label}{s.start_time && s.end_time ? ` (${s.start_time} – ${s.end_time})` : ''}
                </MenuItem>
              ))}
            </TextField>

            <TextField
              type="date"
              label="Fecha"
              value={bookingForm.booking_date}
              onChange={(e) => setBookingForm((f) => ({ ...f, booking_date: e.target.value }))}
              required
              fullWidth
              slotProps={{ inputLabel: { shrink: true } }}
            />

            <TextField
              label="Nombre del cliente"
              value={bookingForm.customer_name}
              onChange={(e) => setBookingForm((f) => ({ ...f, customer_name: e.target.value }))}
              required
              fullWidth
            />

            <TextField
              label="Teléfono del cliente"
              type="tel"
              value={bookingForm.customer_phone}
              onChange={(e) => setBookingForm((f) => ({ ...f, customer_phone: e.target.value }))}
              required
              fullWidth
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3, gap: 1 }}>
            <Button onClick={() => setBookingDialogOpen(false)}>Cancelar</Button>
            <Button type="submit" variant="contained" disabled={manualBookingMutation.isPending}>
              {manualBookingMutation.isPending ? 'Guardando...' : 'Confirmar reserva'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Dialog: crear espacio */}
      <Dialog open={spaceDialogOpen} onClose={() => setSpaceDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Crear espacio</DialogTitle>
        <Box component="form" onSubmit={handleSpaceSubmit}>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2.5, pt: 1 }}>
            <TextField
              label="Nombre"
              value={spaceForm.name}
              onChange={(e) => setSpaceForm((f) => ({ ...f, name: e.target.value }))}
              required
              fullWidth
              autoFocus
            />

            <FormControl fullWidth required>
              <InputLabel>Tipo</InputLabel>
              <Select
                value={spaceForm.type}
                label="Tipo"
                onChange={(e) => setSpaceForm((f) => ({ ...f, type: e.target.value as SpaceType }))}
              >
                {SPACE_TYPE_OPTIONS.map((o) => (
                  <MenuItem key={o.value} value={o.value}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      {o.icon}
                      {o.label}
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <TextField
              label="Descripción"
              value={spaceForm.description}
              onChange={(e) => setSpaceForm((f) => ({ ...f, description: e.target.value }))}
              fullWidth
              multiline
              rows={2}
            />

            <TextField
              label="Precio por turno"
              type="number"
              value={spaceForm.price_per_slot}
              onChange={(e) => setSpaceForm((f) => ({ ...f, price_per_slot: e.target.value }))}
              required
              fullWidth
              slotProps={{
                input: {
                  startAdornment: <InputAdornment position="start">$</InputAdornment>,
                },
              }}
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3, gap: 1 }}>
            <Button onClick={() => setSpaceDialogOpen(false)}>Cancelar</Button>
            <Button type="submit" variant="contained" disabled={createSpaceMutation.isPending}>
              {createSpaceMutation.isPending ? 'Creando...' : 'Crear espacio'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Dialog: nuevo usuario staff */}
      <Dialog open={staffDialogOpen} onClose={() => setStaffDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Nuevo usuario staff</DialogTitle>
        <Box component="form" onSubmit={handleStaffSubmit}>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2.5, pt: 1 }}>
            <TextField
              label="Nombre"
              value={staffForm.name}
              onChange={(e) => setStaffForm((f) => ({ ...f, name: e.target.value }))}
              required
              fullWidth
              autoFocus
            />
            <TextField
              label="Teléfono"
              type="tel"
              value={staffForm.phone}
              onChange={(e) => setStaffForm((f) => ({ ...f, phone: e.target.value }))}
              required
              fullWidth
            />
            <TextField
              label="Contraseña"
              type="password"
              value={staffForm.password}
              onChange={(e) => setStaffForm((f) => ({ ...f, password: e.target.value }))}
              required
              fullWidth
            />
            <FormControl fullWidth required>
              <InputLabel>Rol</InputLabel>
              <Select
                value={staffForm.role}
                label="Rol"
                onChange={(e) => setStaffForm((f) => ({ ...f, role: e.target.value as 'receptionist' | 'admin' }))}
              >
                <MenuItem value="receptionist">Recepcionista</MenuItem>
                <MenuItem value="admin">Admin</MenuItem>
              </Select>
            </FormControl>
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3, gap: 1 }}>
            <Button onClick={() => setStaffDialogOpen(false)}>Cancelar</Button>
            <Button type="submit" variant="contained" disabled={createStaffMutation.isPending}>
              {createStaffMutation.isPending ? 'Creando...' : 'Crear usuario'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Dialog: editar espacio */}
      <Dialog open={!!editingSpace} onClose={() => setEditingSpace(null)} maxWidth="sm" fullWidth>
        <DialogTitle>Editar espacio</DialogTitle>
        <Box component="form" onSubmit={handleEditSubmit}>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2.5, pt: 1 }}>
            <TextField
              label="Nombre"
              value={editForm.name}
              onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))}
              required
              fullWidth
              autoFocus
            />
            <TextField
              label="Descripción"
              value={editForm.description}
              onChange={(e) => setEditForm((f) => ({ ...f, description: e.target.value }))}
              fullWidth
              multiline
              rows={2}
            />
            <TextField
              label="Precio por turno"
              type="number"
              value={editForm.price_per_slot}
              onChange={(e) => setEditForm((f) => ({ ...f, price_per_slot: e.target.value }))}
              required
              fullWidth
              slotProps={{
                input: {
                  startAdornment: <InputAdornment position="start">$</InputAdornment>,
                },
              }}
            />
          </DialogContent>
          <DialogActions sx={{ px: 3, pb: 3, gap: 1 }}>
            <Button onClick={() => setEditingSpace(null)}>Cancelar</Button>
            <Button type="submit" variant="contained" disabled={updateSpaceMutation.isPending}>
              {updateSpaceMutation.isPending ? 'Guardando...' : 'Guardar cambios'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* Dialog: confirmación de eliminación */}
      <Dialog open={!!deleteConfirm} onClose={() => setDeleteConfirm(null)} maxWidth="xs" fullWidth>
        <DialogTitle>¿Eliminar definitivamente?</DialogTitle>
        <DialogContent>
          <Typography>
            Vas a eliminar <strong>{deleteConfirm?.name}</strong>. Esta acción no se puede deshacer.
          </Typography>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3, gap: 1 }}>
          <Button onClick={() => setDeleteConfirm(null)}>Cancelar</Button>
          <Button
            color="error"
            variant="contained"
            disabled={deleteSpaceMutation.isPending || deleteUserMutation.isPending}
            onClick={() => {
              if (deleteConfirm?.type === 'space') deleteSpaceMutation.mutate(deleteConfirm.id);
              else if (deleteConfirm?.type === 'user') deleteUserMutation.mutate(deleteConfirm.id);
            }}
          >
            {deleteSpaceMutation.isPending || deleteUserMutation.isPending ? 'Eliminando...' : 'Sí, eliminar'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Dialog: cancelar reserva */}
      <Dialog open={!!cancelConfirm} onClose={() => setCancelConfirm(null)} maxWidth="xs" fullWidth>
        <DialogTitle>¿Cancelar reserva?</DialogTitle>
        <DialogContent>
          <Typography>
            Vas a cancelar la reserva de <strong>{cancelConfirm?.customer}</strong>. Esta acción no se puede deshacer.
          </Typography>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3, gap: 1 }}>
          <Button onClick={() => setCancelConfirm(null)}>Volver</Button>
          <Button
            color="error"
            variant="contained"
            disabled={cancelBookingMutation.isPending}
            onClick={() => { if (cancelConfirm) cancelBookingMutation.mutate(cancelConfirm.id); }}
          >
            {cancelBookingMutation.isPending ? 'Cancelando...' : 'Sí, cancelar'}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3500}
        onClose={() => setSnackbar((s) => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity={snackbar.severity} onClose={() => setSnackbar((s) => ({ ...s, open: false }))}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
