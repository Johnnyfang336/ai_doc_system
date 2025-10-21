import api from './api';
import { FileItem, UploadResponse, StorageUsage } from '../types';

export const fileService = {
  // Upload file
  async uploadFile(file: File): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await api.post<UploadResponse>('/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    
    return response.data;
  },

// Get user file list
  async getUserFiles(): Promise<FileItem[]> {
    const response = await api.get<{ files: FileItem[] }>('/files');
    return response.data.files || [];
  },

// Get single file information
  async getFile(fileId: number): Promise<FileItem> {
    const response = await api.get<{ file: FileItem }>(`/files/${fileId}`);
    return response.data.file;
  },

// Download file
  async downloadFile(fileId: number): Promise<Blob> {
    const token = localStorage.getItem('token');
    const response = await api.get(`/files/${fileId}/download?token=${token}`, {
      responseType: 'blob',
    });
    return response.data;
  },

// Delete file
  async deleteFile(fileId: number): Promise<void> {
    await api.delete(`/files/${fileId}`);
  },

// Rename file
  async renameFile(fileId: number, newName: string): Promise<FileItem> {
    const response = await api.put<{ file: FileItem }>(`/files/${fileId}/rename`, {
      new_name: newName,
    });
    return response.data.file;
  },

// Get storage usage
  async getStorageUsage(): Promise<StorageUsage> {
    const response = await api.get<StorageUsage>('/storage/usage');
    return response.data;
  },


  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

// Get file icon
  getFileIcon(mimeType: string): string {
    if (mimeType.startsWith('image/')) return '🖼️';
    if (mimeType.startsWith('video/')) return '🎥';
    if (mimeType.startsWith('audio/')) return '🎵';
    if (mimeType.includes('pdf')) return '📄';
    if (mimeType.includes('word')) return '📝';
    if (mimeType.includes('excel') || mimeType.includes('spreadsheet')) return '📊';
    if (mimeType.includes('powerpoint') || mimeType.includes('presentation')) return '📈';
    if (mimeType.includes('zip') || mimeType.includes('rar') || mimeType.includes('7z')) return '🗜️';
    return '📁';
  },
};