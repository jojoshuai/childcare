// web/src/api/auth.ts
import api from './axios'

export interface AuthUser {
  id: string
  nickname: string
  family_id: string | null
  role: string | null
}

export interface AuthResponse {
  token: string
  refresh_token: string
  user: AuthUser
}

export const login = (data: { username: string; password: string }) =>
  api.post<AuthResponse>('/api/auth/login', data).then(r => r.data)

export const register = (data: {
  username: string
  password: string
  family_name: string
  nickname: string
}) => api.post<AuthResponse>('/api/auth/register', data).then(r => r.data)
