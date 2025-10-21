import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Container,
  Avatar,
  Menu,
  MenuItem,
  IconButton,
  Divider,
  Snackbar,
  Alert,
 Dialog,
 DialogTitle,
 DialogContent,
 DialogActions,
 TextField,
 List,
 ListItem,
 ListItemText,
 CircularProgress,
} from '@mui/material';
import {
  AccountCircle,
  ExitToApp,
  Settings,
  CloudUpload,
  People,
  Chat,
 PersonAdd,
 Search,
} from '@mui/icons-material';
import api from '../services/api';
import { User } from '../types';
import { useAuth } from '../hooks/useAuth';
import { useNavigate, useLocation } from 'react-router-dom';
import { fileService } from '../services/fileService';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [uploading, setUploading] = React.useState(false);
  const [success, setSuccess] = React.useState('');
  const [error, setError] = React.useState('');
  const [snackbarOpen, setSnackbarOpen] = React.useState(false);

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
    handleClose();
  };

  const handleProfile = () => {
    navigate('/profile');
    handleClose();
  };

  const navigationItems = [
    { path: '/files', label: 'My Files', icon: <CloudUpload /> },
    { path: '/friends', label: 'Friends', icon: <People /> },
    { path: '/messages', label: 'Messages', icon: <Chat /> },
  ];

  const [addFriendOpen, setAddFriendOpen] = React.useState(false);
  const [searchKeyword, setSearchKeyword] = React.useState('');
  const [searchLoading, setSearchLoading] = React.useState(false);
  const [searchResults, setSearchResults] = React.useState<User[]>([]);
  
  const handleNavFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
     const file = event.target.files?.[0];
     if (!file) return;
     try {
       setUploading(true);
       setSuccess('');
       setError('');
       await fileService.uploadFile(file);
       setSuccess(`File "${file.name}" uploaded successfully`);
        setSnackbarOpen(true);
        navigate('/files');
      } catch (e: any) {
       setError(e.response?.data?.error || 'Upload failed');
        setSnackbarOpen(true);
      } finally {
        setUploading(false);
        event.target.value = '';
      }
   };

  const openAddFriendDialog = () => {
    setAddFriendOpen(true);
    setSearchKeyword('');
    setSearchResults([]);
  };

  const closeAddFriendDialog = () => {
    setAddFriendOpen(false);
  };

  const handleSearchUsers = async () => {
     if (!searchKeyword.trim()) {
      setError('Please enter a search keyword');
       setSnackbarOpen(true);
       return;
     }
     try {
       setSearchLoading(true);
       setError('');
       const resp = await api.get<{ users: User[] }>('/users/search', { params: { keyword: searchKeyword.trim() } });
       setSearchResults(resp.data.users || []);
     } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to search users');
       setSnackbarOpen(true);
     } finally {
       setSearchLoading(false);
     }
   };
 
   const handleSendFriendRequest = async (toUserId: number) => {
     try {
       setSuccess('');
       setError('');
       await api.post('/friends/request', { to_user_id: toUserId });
      setSuccess('Friend request sent');
       setSnackbarOpen(true);
     } catch (e: any) {
      setError(e.response?.data?.error || 'Failed to send friend request');
       setSnackbarOpen(true);
     }
   };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static" sx={{ backgroundColor: '#1976d2' }}>
        <Toolbar>
          <Typography
            variant="h6"
            component="div"
            sx={{ flexGrow: 1, cursor: 'pointer' }}
            onClick={() => navigate('/dashboard')}
          >
            AI Document System
          </Typography>

          {/* Navigation menu */}
          <Box sx={{ display: 'flex', alignItems: 'center', mr: 2 }}>
            {/* Add Friend button placed before Upload File */}
            <Button
              color="inherit"
              startIcon={<PersonAdd />}
              sx={{ mx: 1 }}
              onClick={openAddFriendDialog}
            >
              Add Friend
            </Button>
              {/* Upload File placed before Messages */}
              <Button
                color="inherit"
                component="label"
                startIcon={<CloudUpload />}
                sx={{ mx: 1, opacity: uploading ? 0.7 : 1 }}
                disabled={uploading}
              >
                Upload File
                <input type="file" hidden onChange={handleNavFileUpload} />
              </Button>
 
              {/* Messages item */}
              {navigationItems.filter(i => i.label === 'Messages').map((item) => (
                <Button
                  key={item.path}
                  color="inherit"
                  startIcon={item.icon}
                  onClick={() => navigate(item.path)}
                  sx={{
                    mx: 1,
                    backgroundColor: location.pathname === item.path ? 'rgba(255,255,255,0.1)' : 'transparent',
                  }}
                >
                  {item.label}
                </Button>
              ))}
           </Box>
          {/* Add Friend Dialog */}
          <Dialog open={addFriendOpen} onClose={closeAddFriendDialog} maxWidth="sm" fullWidth>
            <DialogTitle>add friend</DialogTitle>
            <DialogContent>
              <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mb: 2 }}>
                <TextField
                  fullWidth
                  size="small"
                placeholder="Enter username or nickname keyword"
                  value={searchKeyword}
                  onChange={(e) => setSearchKeyword(e.target.value)}
                />
                              <Button variant="contained" startIcon={<Search />} onClick={handleSearchUsers}>Search</Button>
                </Box>
                {searchLoading && (
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <CircularProgress size={20} />           
               <Typography variant="body2">Searching…</Typography>
                </Box>
              )}
                            {!searchLoading && searchResults.length === 0 && (
                              <Alert severity="info">Enter a keyword to search</Alert>
                            )}
                {!searchLoading && searchResults.length > 0 && (
                  <List dense>
                    {searchResults.map((u) => (
                      <ListItem key={u.id}
                        secondaryAction={
                                              <Button variant="outlined" size="small" onClick={() => handleSendFriendRequest(u.id)}>Send Request</Button>
                        }
                      >
                        <ListItemText primary={u.nickname || u.username} secondary={`ID: ${u.id}`} />
                      </ListItem>
                    ))}
                  </List>
                )}
              </DialogContent>
              <DialogActions>
                            <Button onClick={closeAddFriendDialog}>Close</Button>
              </DialogActions>
            </Dialog>

          {/* User menu */}
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Typography variant="body2" sx={{ mr: 1 }}>
              {user?.nickname || user?.username}
            </Typography>
            <IconButton
              size="large"
              aria-label="account of current user"
              aria-controls="menu-appbar"
              aria-haspopup="true"
              onClick={handleMenu}
              color="inherit"
            >
              {user?.avatar ? (
                <Avatar src={user.avatar} sx={{ width: 32, height: 32 }} />
              ) : (
                <AccountCircle />
              )}
            </IconButton>
            <Menu
              id="menu-appbar"
              anchorEl={anchorEl}
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              keepMounted
              transformOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              open={Boolean(anchorEl)}
              onClose={handleClose}
            >
              <MenuItem onClick={handleProfile}>
                <Settings sx={{ mr: 1 }} />
                Profile Settings
              </MenuItem>
              <Divider />
              <MenuItem onClick={handleLogout}>
                <ExitToApp sx={{ mr: 1 }} />
                Logout
              </MenuItem>
            </Menu>
          </Box>
        </Toolbar>
      </AppBar>

      <Container maxWidth="xl" sx={{ mt: 3, mb: 3 }}>
        {children}
      </Container>

      <Snackbar open={snackbarOpen} autoHideDuration={4000} onClose={() => setSnackbarOpen(false)}>
        {(success || error) ? (
          <Alert onClose={() => setSnackbarOpen(false)} severity={success ? 'success' : 'error'} sx={{ width: '100%' }}>
            {success || error}
          </Alert>
        ) : undefined}
      </Snackbar>
    </Box>
  );
};

export default Layout;