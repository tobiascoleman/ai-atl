import apiClient from "./client";
import { FantasyStatusResponse, FantasyTeamsResponse } from "@/types/api";

export async function getYahooAuthUrl(): Promise<string> {
  const { data } = await apiClient.get<{ url: string }>("/fantasy/oauth/url");
  return data.url;
}

export async function fetchFantasyStatus(): Promise<FantasyStatusResponse> {
  const { data } = await apiClient.get<FantasyStatusResponse>(
    "/fantasy/status"
  );
  return data;
}

export async function fetchFantasyTeams(): Promise<FantasyTeamsResponse> {
  const { data } = await apiClient.get<FantasyTeamsResponse>("/fantasy/teams");
  return data;
}
