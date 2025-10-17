import React, { useState, useEffect } from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  LinearProgress,
  Button,
  Paper,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Divider,
} from '@mui/material';
import {
  CloudUpload,
  People,
  Chat,
  Share,
  Storage,
  InsertDriveFile,
} from '@mui/icons-material';
import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';
import { fileService } from '../services/fileService';
import { StorageUsage, FileItem } from '../types';

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [storageUsage, setStorageUsage] = useState<StorageUsage | null>(null);
  const [recentFiles, setRecentFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      const [usage, files] = await Promise.all([
        fileService.getStorageUsage(),
        fileService.getUserFiles(),
      ]);
      
      setStorageUsage(usage);
      // Get recent 5 files
      setRecentFiles(files.slice(0, 5));
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const quickActions = [
    {
      title: 'Upload File',
      description: 'Upload new document files',
      icon: <CloudUpload sx={{ fontSize: 40 }} />,
      color: '#1976d2',
      action: () => navigate('/files'),
    },
    {
      title: 'Friend Management',
      description: 'Manage your friend relationships',
      icon: <People sx={{ fontSize: 40 }} />,
      color: '#388e3c',
      action: () => navigate('/friends'),
    },
    {
      title: 'Message Center',
      description: 'View and send messages',
      icon: <Chat sx={{ fontSize: 40 }} />,
      color: '#f57c00',
      action: () => navigate('/messages'),
    },
    {
      title: 'File Sharing',
      description: 'Share files with friends',
      icon: <Share sx={{ fontSize: 40 }} />,
      color: '#7b1fa2',
      action: () => navigate('/shares'),
    },
  ];

  if (loading) {
    return (
      <Box sx={{ width: '100%', mt: 2 }}>
        <LinearProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Welcome back, {user?.nickname || user?.username}!
      </Typography>
      
      <Grid container spacing={3}>
        {/* Storage Usage */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <Storage sx={{ mr: 1, color: '#1976d2' }} />
                <Typography variant="h6">Storage Usage</Typography>
              </Box>
              {storageUsage && (
                <>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Used {fileService.formatFileSize(storageUsage.used)} / {fileService.formatFileSize(storageUsage.limit)}
                  </Typography>
                  <LinearProgress
                    variant="determinate"
                    value={storageUsage.percentage}
                    sx={{ height: 8, borderRadius: 4 }}
                  />
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                    {storageUsage.percentage.toFixed(1)}% Used
                  </Typography>
                </>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Quick Actions */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Quick Actions
              </Typography>
              <Grid container spacing={2}>
                {quickActions.map((action, index) => (
                  <Grid item xs={6} key={index}>
                    <Button
                      fullWidth
                      variant="outlined"
                      onClick={action.action}
                      sx={{
                        height: 80,
                        flexDirection: 'column',
                        borderColor: action.color,
                        color: action.color,
                        '&:hover': {
                          borderColor: action.color,
                          backgroundColor: `${action.color}10`,
                        },
                      }}
                    >
                      {action.icon}
                      <Typography variant="caption" sx={{ mt: 1 }}>
                        {action.title}
                      </Typography>
                    </Button>
                  </Grid>
                ))}
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* Recent Files */}
        <Grid item xs={12}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Recent Files
            </Typography>
            {recentFiles.length > 0 ? (
              <List>
                {recentFiles.map((file, index) => (
                  <React.Fragment key={file.id}>
                    <ListItem
                      button
                      onClick={() => navigate('/files')}
                    >
                      <ListItemIcon>
                        <InsertDriveFile />
                      </ListItemIcon>
                      <ListItemText
                        primary={file.original_name}
                        secondary={
                          <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                            <span>{fileService.formatFileSize(file.size)}</span>
                            <span>{new Date(file.created_at).toLocaleDateString()}</span>
                          </Box>
                        }
                      />
                    </ListItem>
                    {index < recentFiles.length - 1 && <Divider />}
                  </React.Fragment>
                ))}
              </List>
            ) : (
              <Typography color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>
                No files yet, click upload to get started
              </Typography>
            )}
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;