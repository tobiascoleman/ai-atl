// API Response Types
export interface User {
  id: string
  email: string
  username: string
  created_at: string
}

export interface AuthResponse {
  token: string
  expires_at: string
  user: User
}

export interface Player {
  id: string
  nfl_id: string
  name: string
  team: string
  position: string
  weekly_stats: WeeklyStat[]
  epa_per_play: number
  success_rate: number
  snap_share: number
  target_share: number
  injury_status?: string
  updated_at: string
}

export interface WeeklyStat {
  week: number
  season: number
  opponent?: string
  yards: number
  touchdowns: number
  receptions?: number
  targets?: number
  carries?: number
  passing_yards?: number
  passing_tds?: number
  interceptions?: number
  epa: number
  projected_points: number
  actual_points: number
}

export interface Game {
  id: string
  game_id: string
  season: number
  week: number
  home_team: string
  away_team: string
  start_time: string
  status: 'scheduled' | 'live' | 'final'
  vegas_line: number
  over_under: number
  home_score: number
  away_score: number
}

export interface GameScriptPrediction {
  game_id: string
  predicted_flow: string
  player_impacts: PlayerImpact[]
  confidence_score: number
  key_factors: string[]
}

export interface PlayerImpact {
  player_name: string
  impact: string
  reasoning: string
}

export interface FantasyLineup {
  id: string
  user_id: string
  week: number
  season: number
  positions: Record<string, string>
  projected_points: number
  actual_points: number
  created_at: string
  updated_at: string
}

export interface TradeAnalysis {
  team_a_grade: string
  team_b_grade: string
  fairness_score: number
  ai_analysis: string
  team_a_value_change: string
  team_b_value_change: string
}

export interface ChatMessage {
  question: string
  response: string
  timestamp?: string
}

export interface Streak {
  player_id: string
  player_name: string
  streak_type: 'over' | 'under' | 'hot' | 'cold'
  stat_line: string
  games_in_streak: number
  ai_explanation: string
  confidence: number
}

export interface WaiverGem {
  player: Player
  ownership: number
  epa_per_play: number
  ai_analysis: string
  recommendation: string
}

export interface Vote {
  id: string
  user_id: string
  player_id: string
  prediction_type: 'over' | 'under' | 'lock' | 'fade'
  stat_line: number
  week: number
  created_at: string
}

export interface VoteConsensus {
  player_id: string
  week: number
  total_votes: number
  consensus: Record<string, number>
  percentages: Record<string, number>
}

