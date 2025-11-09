import apiClient from './client'
import { Game } from '@/types/api'

export const gamesAPI = {
  getScheduledGames: async (season: number, week?: number): Promise<Game[]> => {
    const params = new URLSearchParams({ season: season.toString() })
    if (week) {
      params.append('week', week.toString())
    }
    const { data } = await apiClient.get<{ games: Game[] }>(
      `/data/games/scheduled?${params.toString()}`
    )
    return data.games
  },

  getGame: async (gameId: string): Promise<Game> => {
    const { data } = await apiClient.get<Game>(`/data/games/${gameId}`)
    return data
  },

  getGames: async (season: number, week?: number): Promise<Game[]> => {
    const params = new URLSearchParams({ season: season.toString() })
    if (week) {
      params.append('week', week.toString())
    }
    const { data } = await apiClient.get<{ games: Game[] }>(
      `/data/games?${params.toString()}`
    )
    return data.games
  },
}

