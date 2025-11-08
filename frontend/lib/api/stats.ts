import apiClient from './client'

export interface DashboardStats {
  total_players: number
  total_games: number
  total_plays: number
  injured_players: number
  next_gen_stats: number
  active_teams: number
  current_season_year: number
}

export const statsAPI = {
  getDashboardStats: async (): Promise<DashboardStats> => {
    const { data } = await apiClient.get<DashboardStats>('/stats/dashboard')
    return data
  },
}

