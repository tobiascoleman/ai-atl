import apiClient from './client'
import { GameScriptPrediction, Streak, WaiverGem } from '@/types/api'

export const insightsAPI = {
  getGameScript: async (gameId: string): Promise<GameScriptPrediction> => {
    const { data } = await apiClient.get<GameScriptPrediction>(
      `/insights/game_script?game_id=${gameId}`
    )
    return data
  },

  getInjuryImpact: async (playerId: string) => {
    const { data } = await apiClient.post('/insights/injury_impact', {
      player_id: playerId,
    })
    return data
  },

  getStreaks: async (playerId: string): Promise<{ streaks: Streak[] }> => {
    const { data } = await apiClient.get<{ streaks: Streak[] }>(
      `/insights/streaks?player_id=${playerId}`
    )
    return data
  },

  getTopPerformers: async (week: number, type: 'over' | 'under' = 'over') => {
    const { data } = await apiClient.get(
      `/insights/top_performers?week=${week}&type=${type}`
    )
    return data
  },

  getWaiverGems: async (position: string = 'ALL'): Promise<{ gems: WaiverGem[]; count: number }> => {
    const { data} = await apiClient.get<{ gems: WaiverGem[]; count: number }>(
      `/insights/waiver_gems?position=${position}`
    )
    return data
  },

  getPersonalizedWaiverGems: async (roster: any[]): Promise<{ gems: WaiverGem[]; count: number }> => {
    const { data } = await apiClient.post<{ gems: WaiverGem[]; count: number }>(
      '/insights/personalized_waiver_gems',
      { roster }
    )
    return data
  },
}

