import apiClient from './client'
import { Player } from '@/types/api'

export const playersAPI = {
  getPlayers: async (params?: {
    search?: string
    team?: string
    position?: string
    current?: string
    page?: number
    limit?: number
    sort?: string
    order?: 'asc' | 'desc'
  }) => {
    const { data } = await apiClient.get<{
      players: Player[]
      total: number
      page: number
      limit: number
    }>('/players', { params })
    return data
  },

  getPlayer: async (id: string): Promise<Player> => {
    const { data } = await apiClient.get<Player>(`/players/${id}`)
    return data
  },

  getPlayerStats: async (id: string, season?: number, week?: number) => {
    const { data } = await apiClient.get(`/players/${id}/stats`, {
      params: { season, week },
    })
    return data
  },
}

