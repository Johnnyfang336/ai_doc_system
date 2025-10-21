// User related types
export interface User {
  id: number;
  username: string;
  email: string;
  nickname: string;
  avatar?: string;
  role: 'user' | 'admin';
  storage_used: number;
  storage_limit: number;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  nickname: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// File related types
export interface FileItem {
  id: number;
  original_name: string;
  filename: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface UploadResponse {
  message: string;
  file: FileItem;
}

export interface StorageUsage {
  used: number;
  limit: number;
  percentage: number;
}

// Friend related types
export interface Friend {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
  status: 'pending' | 'accepted' | 'rejected';
  created_at: string;
}

export interface FriendRequest {
  id: number;
  requester_id: number;
  receiver_id: number;
  status: 'pending' | 'accepted' | 'rejected';
  requester: User;
  created_at: string;
}

export interface FriendGroup {
  id: number;
  name: string;
  user_id: number;
  created_at: string;
}

// Message related types
export interface Message {
  id: number;
  sender_id: number;
  receiver_id: number;
  content: string;
  is_read: boolean;
  created_at: string;
  sender?: User;
}

export interface ChatItem {
  friend_id: number;
  friend: User;
  last_message: Message;
  unread_count: number;
}

// File sharing related types
export interface FileShare {
  id: number;
  file_id: number;
  owner_id: number;
  shared_with_user_id?: number;
  share_token?: string;
  expires_at?: string;
  created_at: string;
  file: FileItem;
  owner: User;
  shared_with?: User;
}

// API response types
export interface ApiResponse<T = any> {
  message?: string;
  data?: T;
  error?: string;
}

// Common pagination types
export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}