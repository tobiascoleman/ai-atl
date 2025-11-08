import apiClient from "./client";

export interface ESPNCredentials {
  espn_s2: string;
  espn_swid: string;
  league_id: number;
  team_id: number;
  year: number;
}

export interface ESPNPlayer {
  name: string;
  position: string;
  proTeam: string;
  lineupSlot: string;
  projectedPoints: number;
  points: number;
  injured: boolean;
  injuryStatus?: string | null;
  eligibleSlots?: string[];
  recommendedSlot?: string;
  playerId?: number;
}

export interface ESPNStatusResponse {
  connected: boolean;
}

export interface ESPNRosterResponse {
  connected: boolean;
  players: ESPNPlayer[];
}

export async function saveESPNCredentials(
  credentials: ESPNCredentials
): Promise<{ message: string; connected: boolean }> {
  const { data } = await apiClient.post<{ message: string; connected: boolean }>(
    "/espn/credentials",
    credentials
  );
  return data;
}

export async function fetchESPNStatus(): Promise<ESPNStatusResponse> {
  const { data } = await apiClient.get<ESPNStatusResponse>("/espn/status");
  return data;
}

export async function fetchESPNRoster(): Promise<ESPNRosterResponse> {
  const { data } = await apiClient.get<ESPNRosterResponse>("/espn/roster");
  return data;
}

export interface OptimizeLineupResponse {
  optimalLineup: ESPNPlayer[];
  bench: ESPNPlayer[];
  totalProjected: number;
}

export async function optimizeESPNLineup(): Promise<OptimizeLineupResponse> {
  const { data } = await apiClient.get<OptimizeLineupResponse>("/espn/optimize-lineup");
  return data;
}

export interface FreeAgentPlayer {
  name: string;
  position: string;
  proTeam: string;
  projectedPoints: number;
  points: number;
  injured: boolean;
  injuryStatus: string | null;
  playerId?: number;
  percentOwned: number;
  percentStarted: number;
}

export interface FreeAgentsResponse {
  players: FreeAgentPlayer[];
  count: number;
}

export async function fetchFreeAgents(
  position?: string,
  size: number = 50
): Promise<FreeAgentsResponse> {
  const params = new URLSearchParams();
  if (position) params.append("position", position);
  params.append("size", size.toString());
  
  const { data } = await apiClient.get<FreeAgentsResponse>(
    `/espn/free-agents?${params.toString()}`
  );
  return data;
}

export interface AIStartSitRequest {
  playerA: ESPNPlayer;
  playerB: ESPNPlayer;
}

export interface AIStartSitResponse {
  recommendation: string; // "A" or "B"
  confidence: number; // 0-100
  reasoning: string;
  playerAName: string;
  playerBName: string;
}

export async function getAIStartSitAdvice(
  playerA: ESPNPlayer,
  playerB: ESPNPlayer
): Promise<AIStartSitResponse> {
  const { data } = await apiClient.post<AIStartSitResponse>(
    "/espn/ai-start-sit",
    { playerA, playerB }
  );
  return data;
}
