import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Typography,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  IconButton,
  Button,
  Divider,
  Collapse,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  Snackbar,
  Radio,
 Tooltip,
} from '@mui/material';
import {
  InsertDriveFile,
  Edit,
  Share,
  Delete,
  ExpandLess,
  ExpandMore,
  People,
} from '@mui/icons-material';
import { fileService } from '../services/fileService';
import api from '../services/api';
import { FileItem, Friend } from '../types';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();

  // Sidebar collapse states
  const [filesOpen, setFilesOpen] = useState(true);
  const [friendsOpen, setFriendsOpen] = useState(true);

  // Data
  const [files, setFiles] = useState<FileItem[]>([]);
  const [friends, setFriends] = useState<Friend[]>([]);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [loadingFriends, setLoadingFriends] = useState(false);

  // Share dialog state
  const [shareDialogOpen, setShareDialogOpen] = useState(false);
  const [shareMode, setShareMode] = useState<'friend' | 'public'>('friend');
  const [shareFile, setShareFile] = useState<FileItem | null>(null);
  const [selectedFriendId, setSelectedFriendId] = useState<number | null>(null);
  const [generatedLink, setGeneratedLink] = useState<string>('');
 
  // Feedback
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [snackbarOpen, setSnackbarOpen] = useState(false);
 
  // Inline editor state
  const [editingFile, setEditingFile] = useState<FileItem | null>(null);

  useEffect(() => {
    loadFiles();
    loadFriends();
  }, []);

  const loadFiles = async () => {
    try {
      setLoadingFiles(true);
      const userFiles = await fileService.getUserFiles();
      setFiles(userFiles);
    } catch (e) {
      console.error(e);
      setError('Failed to load files');
      setSnackbarOpen(true);
    } finally {
      setLoadingFiles(false);
    }
  };

  const loadFriends = async () => {
    try {
      setLoadingFriends(true);
      const resp = await api.get<{ friends: Friend[] }>('/friends');
      setFriends(resp.data.friends || []);
    } catch (e) {
      console.error(e);
      setError('Failed to load friends');
      setSnackbarOpen(true);
    } finally {
      setLoadingFriends(false);
    }
  };

  const isSupportedFileType = (mimeType: string): boolean => {
    const supportedTypes = [
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document', // .docx
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', // .xlsx
      'application/vnd.openxmlformats-officedocument.presentationml.presentation', // .pptx
      'application/msword', // .doc
      'application/vnd.ms-excel', // .xls
      'application/vnd.ms-powerpoint', // .ppt
      'application/pdf', // .pdf
      'text/plain', // .txt
      'application/vnd.oasis.opendocument.text', // .odt
      'application/vnd.oasis.opendocument.spreadsheet', // .ods
      'application/vnd.oasis.opendocument.presentation', // .odp
    ];
    return supportedTypes.includes(mimeType);
  };

  const getShortName = (name: string) => {
    if (!name) return '';
    const lastDot = name.lastIndexOf('.');
    const base = lastDot > 0 ? name.slice(0, lastDot) : name;
    const ext = lastDot > 0 ? name.slice(lastDot) : '';
    const head = base.slice(0, 4);
    return ext ? `${head}…${ext}` : `${head}…`;
  };

  // Build editor iframe src when a file is selected
  const token = localStorage.getItem('token') || '';
  const editorSrc = editingFile
    ? `/api/files/${editingFile.id}/edit?token=${encodeURIComponent(token)}`
    : '';
 

  const handleEdit = (file: FileItem) => {
    // Open editor inline inside Dashboard (2/3 width area)
    setEditingFile(file);
  };

  const handleDelete = async (file: FileItem) => {
    try {
      await fileService.deleteFile(file.id);
      setSuccess(`Deleted file "${file.original_name}"`);
      setSnackbarOpen(true);
      await loadFiles();
    } catch (e) {
      console.error(e);
      setError('Delete failed');
      setSnackbarOpen(true);
    }
  };

  const openShareDialog = (file: FileItem) => {
    setShareFile(file);
    setShareMode('friend');
    setSelectedFriendId(null);
    setGeneratedLink('');
    setShareDialogOpen(true);
  };

  const closeShareDialog = () => {
    setShareDialogOpen(false);
  };

  const handleShareToFriend = async () => {
    if (!shareFile || !selectedFriendId) return;
    try {
      await api.post('/shares/friend', {
        file_id: shareFile.id,
        friend_id: selectedFriendId,
      });
      setSuccess('Successfully shared to friend');
      setSnackbarOpen(true);
      setShareDialogOpen(false);
    } catch (e: any) {
      console.error(e);
      setError(e.response?.data?.error || 'Share failed');
      setSnackbarOpen(true);
    }
  };

  const handleGenerateLink = async () => {
    if (!shareFile) return;
    try {
      const resp = await api.post<{ share: { share_token: string } }>(
        '/shares/public',
        { file_id: shareFile.id }
      );
      const token = (resp.data as any)?.share?.share_token || (resp.data as any)?.share_token;
      if (token) {
        const origin = window.location.origin;
        const url = `${origin}/api/share/${token}`;
        setGeneratedLink(url);
      } else {
        setError('No share link returned by server');
        setSnackbarOpen(true);
      }
    } catch (e: any) {
      console.error(e);
      setError(e.response?.data?.error || 'Failed to generate share link');
      setSnackbarOpen(true);
    }
  };

  return (
    <Box sx={{ display: 'flex', gap: 2 }}>
      {/* Left sidebar */}
      <Box
        sx={{
       width: 94,
       minWidth: 94,

          flexShrink: 0,
          borderRight: '1px solid #e0e0e0',
          pr: 1,
        }}
      >
        {/* My Files header */}
        <Box
          sx={{ display: 'flex', alignItems: 'center', cursor: 'pointer', py: 1 }}
          onClick={() => setFilesOpen((prev) => !prev)}
        >
          <InsertDriveFile sx={{ mr: 1 }} />
          <Typography variant="h6" sx={{ fontSize: '0.625rem', overflowWrap: 'break-word', whiteSpace: 'normal', lineHeight: 1 }}>My Files</Typography>
          <Box sx={{ flexGrow: 1 }} />
          {filesOpen ? <ExpandLess /> : <ExpandMore />}
        </Box>
        <Divider />
        <Collapse in={filesOpen} timeout="auto" unmountOnExit>
          <List dense>
            {loadingFiles ? (
              <ListItem>
                <ListItemText primary="Loading files..." />
              </ListItem>
            ) : files.length === 0 ? (
              <ListItem>
                <ListItemText primary="No files" />
              </ListItem>
            ) : (
              files.map((file) => (
                <ListItem
                  key={file.id}
                  disableGutters
                  sx={{ px: 0.5, flexDirection: 'column', alignItems: 'flex-start' }}
                >
                  <Box sx={{ display: 'flex', width: '100%', alignItems: 'center' }}>
                    <ListItemIcon sx={{ minWidth: 0, mr: 0.5 }}>
                      <span style={{ fontSize: 16 }}>{fileService.getFileIcon(file.mime_type)}</span>
                    </ListItemIcon>
                    <ListItemText
                     primary={file.original_name}
                      primaryTypographyProps={{ sx: { fontSize: '1rem', whiteSpace: 'normal', overflowWrap: 'break-word', wordBreak: 'break-word' } }}
                     />
                  </Box>
                  <Box sx={{ display: 'flex', gap: 0.5, mt: 0.5 }}>
                    <Tooltip title="Edit">
                      <span>
                        <IconButton
                          size="small"
                          aria-label="edit"
                          disabled={!isSupportedFileType(file.mime_type)}
                          onClick={() => handleEdit(file)}
                        >
                          <Edit sx={{ fontSize: 12 }} />
                        </IconButton>
                      </span>
                    </Tooltip>
                    <Tooltip title="Share">
                      <IconButton size="small" aria-label="share" onClick={() => openShareDialog(file)}>
                        <Share sx={{ fontSize: 12 }} />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton size="small" aria-label="delete" color="error" onClick={() => handleDelete(file)}>
                        <Delete sx={{ fontSize: 12 }} />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </ListItem>
              ))
            )}
          </List>
        </Collapse>

        {/* Friends header */}
        <Box
          sx={{ display: 'flex', alignItems: 'center', cursor: 'pointer', py: 1, mt: 2 }}
          onClick={() => setFriendsOpen((prev) => !prev)}
        >
          <People sx={{ mr: 1 }} />
          <Typography variant="h6" sx={{ fontSize: '0.625rem', overflowWrap: 'break-word', whiteSpace: 'normal', lineHeight: 1 }}>Friends</Typography>
          <Box sx={{ flexGrow: 1 }} />
          {friendsOpen ? <ExpandLess /> : <ExpandMore />}
        </Box>
        <Divider />
        <Collapse in={friendsOpen} timeout="auto" unmountOnExit>
          <List dense>
            {loadingFriends ? (
              <ListItem>
                <ListItemText primary="Loading friends..." />
              </ListItem>
            ) : friends.length === 0 ? (
              <ListItem>
                <ListItemText primary="No friends" />
              </ListItem>
            ) : (
              friends.map((f) => (
                <ListItem key={f.id}>
                  <ListItemIcon>
                    <People />
                  </ListItemIcon>
                  <ListItemText primary={f.nickname || (f as any).username || ''} />
                </ListItem>
              ))
            )}
          </List>
        </Collapse>
      </Box>

      {/* Right content area with inline editor (occupies 2/3 of right area) */}
      <Box sx={{ flexGrow: 1, p: 2, display: 'flex', gap: 2 }}>
      {/* Editor column: 2/3 width */}
      <Box sx={{ flex: 2 }}>
      {editorSrc ? (
      <iframe
      src={editorSrc}
      title="OnlyOffice Editor"
      style={{ width: '100%', height: '80vh', border: 'none' }}
      allow="clipboard-read; clipboard-write"
      />
      ) : (
      <Alert severity="info">Please select the file on the left and click "Edit" to open it here.</Alert>
      )}
      </Box>
      {/* Auxiliary content column: 1/3 width */}
      <Box sx={{ flex: 1 }}>
      <Typography variant="h5" sx={{ mb: 2 }}>
      Dashboard
      </Typography>
      <Alert severity="info">Select a file on the left to Edit / Share / Delete.</Alert>
      </Box>
      </Box>

      {/* Share Dialog */}
      <Dialog open={shareDialogOpen} onClose={closeShareDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Share File{shareFile ? `: ${shareFile.original_name}` : ''}</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
            <Button
              variant={shareMode === 'friend' ? 'contained' : 'outlined'}
              onClick={() => setShareMode('friend')}
            >
              Share to Friend
            </Button>
            <Button
              variant={shareMode === 'public' ? 'contained' : 'outlined'}
              onClick={() => setShareMode('public')}
            >
              Generate Share Link
            </Button>
          </Box>

          {shareMode === 'friend' ? (
            <Box>
              <Typography variant="subtitle1" sx={{ mb: 1 }}>Select one friend (single choice)</Typography>
              <List dense>
                {friends.length === 0 ? (
                  <ListItem>
                    <ListItemText primary="No friends available to share" />
                  </ListItem>
                ) : (
                  friends.map((f) => (
                    <ListItem
                      key={f.id}
                      secondaryAction={
                        <Radio
                          checked={selectedFriendId === f.id}
                          onChange={() => setSelectedFriendId(f.id)}
                        />
                      }
                    >
                      <ListItemIcon>
                        <People />
                      </ListItemIcon>
                      <ListItemText primary={f.nickname || (f as any).username || ''} />
                    </ListItem>
                  ))
                )}
              </List>
            </Box>
          ) : (
            <Box>
              <Typography variant="subtitle1" sx={{ mb: 1 }}>Click the button below to generate a link</Typography>
              <Button variant="contained" onClick={handleGenerateLink}>Generate Link</Button>
              {generatedLink && (
                <Box sx={{ mt: 2 }}>
                  <Alert severity="success">Share link generated:</Alert>
                  <Typography variant="body2" sx={{ mt: 1, wordBreak: 'break-all' }}>{generatedLink}</Typography>
                  <Button
                    sx={{ mt: 1 }}
                    variant="outlined"
                    onClick={() => navigator.clipboard.writeText(generatedLink)}
                  >
                    Copy Link
                  </Button>
                </Box>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={closeShareDialog}>Cancel</Button>
          {shareMode === 'friend' ? (
            <Button
              onClick={handleShareToFriend}
              variant="contained"
              disabled={!selectedFriendId}
            >
              Confirm Share
            </Button>
          ) : (
            <Button onClick={closeShareDialog} variant="contained">Done</Button>
          )}
        </DialogActions>
      </Dialog>

      {/* Snackbar */}
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

export default Dashboard;