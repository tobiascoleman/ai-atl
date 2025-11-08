import apiClient from './client'
import { AuthResponse } from '@/types/api'

export const authAPI = {
  register: async (email: string, username: string, password: string): Promise<AuthResponse> => {
    const { data } = await apiClient.post<AuthResponse>('/auth/register', {
      email,
      username,
      password,
    })
    return data
  },

  login: async (email: string, password: string): Promise<AuthResponse> => {
    const { data } = await apiClient.post<AuthResponse>('/auth/login', {
      email,
      password,
    })
    return data
  },

  refreshToken: async (): Promise<AuthResponse> => {
    const { data } = await apiClient.post<AuthResponse>('/auth/refresh')
    return data
  },
}

