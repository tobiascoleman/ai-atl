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
