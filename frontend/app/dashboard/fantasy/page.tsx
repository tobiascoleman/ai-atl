"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Link2, RefreshCcw, Trophy } from "lucide-react";

import {
  fetchFantasyStatus,
  fetchFantasyTeams,
  getYahooAuthUrl,
} from "@/lib/api/fantasy";
import {
  fetchESPNStatus,
  fetchESPNRoster,
  saveESPNCredentials,
  optimizeESPNLineup,
  fetchFreeAgents,
  getAIStartSitAdvice,
  ESPNPlayer,
  ESPNCredentials,
  FreeAgentPlayer,
  AIStartSitResponse,
} from "@/lib/api/espn";
import { insightsAPI } from "@/lib/api/insights";
import { FantasyStatusResponse, YahooTeam, WaiverGem } from "@/types/api";

export default function FantasyPage() {
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<FantasyStatusResponse | null>(null);
  const [teams, setTeams] = useState<YahooTeam[]>([]);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // ESPN state
  const [espnConnected, setEspnConnected] = useState(false);
  const [espnRoster, setEspnRoster] = useState<ESPNPlayer[]>([]);
  const [espnLoading, setEspnLoading] = useState(false);
  const [espnError, setEspnError] = useState<string | null>(null);
  const [showESPNForm, setShowESPNForm] = useState(false);
  const [showOptimized, setShowOptimized] = useState(false);
  const [optimizedLineup, setOptimizedLineup] = useState<ESPNPlayer[]>([]);
  const [benchPlayers, setBenchPlayers] = useState<ESPNPlayer[]>([]);
  const [totalProjected, setTotalProjected] = useState(0);
  const [espnCreds, setEspnCreds] = useState<ESPNCredentials>({
    espn_s2: "",
    espn_swid: "",
    league_id: 0,
    team_id: 0,
    year: 2025,
  });
  const [freeAgents, setFreeAgents] = useState<FreeAgentPlayer[]>([]);
  const [selectedPosition, setSelectedPosition] = useState<string>("");
  const [showFreeAgents, setShowFreeAgents] = useState(false);
  const [aiAdvice, setAiAdvice] = useState<AIStartSitResponse | null>(null);
  const [showAIModal, setShowAIModal] = useState(false);
  const [comparingPlayers, setComparingPlayers] = useState<{playerA: ESPNPlayer | null, playerB: ESPNPlayer | null}>({
    playerA: null,
    playerB: null,
  });
  
  // Waiver Wire AI Scout state
  const [waiverGems, setWaiverGems] = useState<WaiverGem[]>([]);
  const [waiverPosition, setWaiverPosition] = useState<string>("ALL");
  const [showWaiverGems, setShowWaiverGems] = useState(false);
  const [waiverLoading, setWaiverLoading] = useState(false);

  const connectedQuery = useMemo(
    () => searchParams?.get("connected"),
    [searchParams]
  );

  useEffect(() => {
    if (connectedQuery === "1") {
      setSuccess("Yahoo account connected successfully!");
    }
  }, [connectedQuery]);

  const loadFantasyData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const fantasyStatus = await fetchFantasyStatus();
      setStatus(fantasyStatus);

      if (fantasyStatus.connected) {
        const fantasyTeams = await fetchFantasyTeams();
        setTeams(fantasyTeams.teams);
      } else {
        setTeams([]);
      }
    } catch (err) {
      console.error(err);
      setError(
        "Unable to reach the fantasy service right now. Please try again shortly."
      );
    } finally {
      setLoading(false);
    }
  }, []);

  const loadESPNData = useCallback(async () => {
    setEspnLoading(true);
    setEspnError(null);
    try {
      const espnStatus = await fetchESPNStatus();
      setEspnConnected(espnStatus.connected);

      if (espnStatus.connected) {
        const rosterData = await fetchESPNRoster();
        setEspnRoster(rosterData.players);
      } else {
        setEspnRoster([]);
      }
    } catch (err) {
      console.error(err);
      setEspnError("Unable to reach ESPN service. Please try again.");
    } finally {
      setEspnLoading(false);
    }
  }, []);

  const handleESPNConnect = async () => {
    setEspnLoading(true);
    setEspnError(null);
    try {
      await saveESPNCredentials(espnCreds);
      setEspnConnected(true);
      setShowESPNForm(false);
      await loadESPNData();
      setSuccess("ESPN account connected successfully!");
    } catch (err) {
      console.error(err);
      setEspnError("Failed to save ESPN credentials. Please check your input.");
    } finally {
      setEspnLoading(false);
    }
  };

  const handleOptimizeLineup = async () => {
    setEspnLoading(true);
    setEspnError(null);
    try {
      const result = await optimizeESPNLineup();
      setOptimizedLineup(result.optimalLineup);
      setBenchPlayers(result.bench);
      setTotalProjected(result.totalProjected);
      setShowOptimized(true);
    } catch (err) {
      console.error(err);
      setEspnError("Failed to optimize lineup.");
    } finally {
      setEspnLoading(false);
    }
  };

  const handleLoadFreeAgents = async (position?: string) => {
    setEspnLoading(true);
    setEspnError(null);
    try {
      const result = await fetchFreeAgents(position, 50);
      setFreeAgents(result.players);
      setShowFreeAgents(true);
    } catch (err: any) {
      console.error("Free agents error:", err);
      const errorMsg = err?.response?.data?.error || err?.message || "Failed to load free agents.";
      setEspnError(errorMsg);
    } finally {
      setEspnLoading(false);
    }
  };

  const handleSelectPlayerForComparison = (player: ESPNPlayer) => {
    if (!comparingPlayers.playerA) {
      setComparingPlayers({ playerA: player, playerB: null });
    } else if (!comparingPlayers.playerB) {
      setComparingPlayers({ ...comparingPlayers, playerB: player });
    } else {
      // Reset and start over
      setComparingPlayers({ playerA: player, playerB: null });
    }
  };

  const handleGetAIAdvice = async () => {
    if (!comparingPlayers.playerA || !comparingPlayers.playerB) {
      setEspnError("Please select two players to compare");
      return;
    }

    setEspnLoading(true);
    setEspnError(null);
    try {
      const result = await getAIStartSitAdvice(comparingPlayers.playerA, comparingPlayers.playerB);
      setAiAdvice(result);
      setShowAIModal(true);
    } catch (err: any) {
      console.error("AI advice error:", err);
      const errorMsg = err?.response?.data?.error || err?.message || "Failed to get AI recommendation.";
      setEspnError(errorMsg);
    } finally {
      setEspnLoading(false);
    }
  };

  const resetComparison = () => {
    setComparingPlayers({ playerA: null, playerB: null });
    setAiAdvice(null);
    setShowAIModal(false);
  };

  const loadWaiverGems = async (position: string = "ALL") => {
    setWaiverLoading(true);
    setEspnError(null);
    try {
      // If we have ESPN roster data, get personalized recommendations
      if (espnRoster && espnRoster.length > 0) {
        const rosterData = espnRoster.map(player => ({
          name: player.name,
          position: player.position,
          projectedPoints: player.projectedPoints,
          lineupSlot: player.lineupSlot,
        }));
        
        const result = await insightsAPI.getPersonalizedWaiverGems(rosterData);
        setWaiverGems(result.gems);
        setShowWaiverGems(true);
      } else {
        // Fallback to generic recommendations
        const result = await insightsAPI.getWaiverGems(position);
        setWaiverGems(result.gems);
        setShowWaiverGems(true);
      }
    } catch (err: any) {
      console.error("Waiver gems error:", err);
      const errorMsg = err?.response?.data?.error || err?.message || "Failed to load waiver wire gems.";
      setEspnError(errorMsg);
    } finally {
      setWaiverLoading(false);
    }
  };

  useEffect(() => {
    loadFantasyData();
    loadESPNData();
  }, [loadFantasyData, loadESPNData]);

  const handleConnect = async () => {
    setConnecting(true);
    setError(null);
    try {
      const url = await getYahooAuthUrl();
      window.location.href = url;
    } catch (err) {
      console.error(err);
      setError(
        "Could not start Yahoo authentication. Double-check your connection and try again."
      );
      setConnecting(false);
    }
  };

  const handleRefresh = () => {
    loadFantasyData();
  };

  const renderStatus = () => {
    if (loading && !status) {
      return (
        <p className="text-gray-500">Checking your fantasy integration...</p>
      );
    }

    if (!status?.enabled) {
      return (
        <div className="rounded-lg bg-yellow-50 p-4 text-sm text-yellow-800">
          Yahoo Fantasy integration isn&apos;t configured for this environment
          yet. Ask your admin to add credentials to enable it.
        </div>
      );
    }

    if (!status.connected) {
      return (
        <div className="space-y-4">
          <p className="text-gray-600">
            Link your Yahoo! Fantasy Sports account to pull in your NFL teams.
            We&apos;ll only read your roster details.
          </p>
          <button
            onClick={handleConnect}
            disabled={connecting}
            className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
          >
            <Link2 size={18} />
            {connecting ? "Connecting…" : "Connect Yahoo Account"}
          </button>
        </div>
      );
    }

    return (
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="font-medium text-green-600">Yahoo account connected</p>
          <p className="text-sm text-gray-500">
            Your latest fantasy teams are synced below.
          </p>
        </div>
        <button
          onClick={handleRefresh}
          className="inline-flex items-center gap-2 rounded-lg border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 transition hover:border-gray-300 hover:text-gray-900"
        >
          <RefreshCcw size={16} />
          Refresh data
        </button>
      </div>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2">
        <h1 className="flex items-center gap-2 text-3xl font-semibold text-gray-900">
          <Trophy className="text-blue-600" size={28} />
          Fantasy Central
        </h1>
        <p className="text-gray-600">
          Connect your Yahoo! Fantasy account to preview your teams inside
          AI-ATL. This proof of concept syncs a snapshot of your rosters.
        </p>
      </div>

      {success && (
        <div className="rounded-lg bg-green-50 p-4 text-sm text-green-700">
          {success}
        </div>
      )}

      {error && (
        <div className="rounded-lg bg-red-50 p-4 text-sm text-red-700">
          {error}
        </div>
      )}

      <section className="rounded-xl bg-white p-6 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-800">
          Account connection
        </h2>
        <div className="mt-4">{renderStatus()}</div>
      </section>

      {status?.connected && (
        <section className="rounded-xl bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-800">Yahoo teams</h2>
            {!loading && (
              <span className="text-sm text-gray-500">
                {teams.length} linked team{teams.length === 1 ? "" : "s"}
              </span>
            )}
          </div>

          {loading && (
            <p className="mt-6 text-sm text-gray-500">
              Loading your fantasy teams…
            </p>
          )}

          {!loading && teams.length === 0 && (
            <p className="mt-6 text-sm text-gray-500">
              We couldn&apos;t find any NFL teams tied to your Yahoo account.
              Double-check you&apos;re in an active league this season.
            </p>
          )}

          <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {teams.map((team) => (
              <div
                key={team.team_key}
                className="rounded-lg border border-gray-100 bg-gray-50 p-4 transition hover:border-blue-200 hover:bg-white"
              >
                <p className="text-sm uppercase tracking-wide text-gray-500">
                  {team.league_name || "NFL League"}
                </p>
                <h3 className="mt-1 text-lg font-semibold text-gray-900">
                  {team.name}
                </h3>
                <p className="mt-3 text-xs text-gray-500">
                  Team key: {team.team_key}
                </p>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* ESPN Fantasy Section */}
      <section className="rounded-xl bg-white p-6 shadow-sm border-2 border-orange-200">
        <h2 className="text-lg font-semibold text-gray-800 flex items-center gap-2">
          <span className="text-orange-600">ESPN</span> Fantasy Football
        </h2>

        {espnError && (
          <div className="mt-4 rounded-lg bg-red-50 p-4 text-sm text-red-700">
            {espnError}
          </div>
        )}

        {!espnConnected && !showESPNForm && (
          <div className="mt-4 space-y-4">
            <p className="text-gray-600">
              Connect your ESPN Fantasy Football account to view your roster
              directly in AI-ATL.
            </p>
            <button
              onClick={() => setShowESPNForm(true)}
              className="inline-flex items-center gap-2 rounded-lg bg-orange-600 px-4 py-2 font-medium text-white transition hover:bg-orange-700"
            >
              <Link2 size={18} />
              Connect ESPN Account
            </button>
          </div>
        )}

        {showESPNForm && (
          <div className="mt-4 space-y-4">
            <div className="rounded-lg bg-blue-50 p-4 text-sm text-blue-800">
              <p className="font-medium">How to get your ESPN credentials:</p>
              <ol className="mt-2 list-decimal list-inside space-y-1">
                <li>Log into ESPN Fantasy Football in your browser</li>
                <li>Open Developer Tools (F12)</li>
                <li>Go to Application → Cookies → fantasy.espn.com</li>
                <li>Copy the <code>espn_s2</code> and <code>SWID</code> values</li>
                <li>Get your league ID and team ID from your team URL</li>
              </ol>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  ESPN_S2 Cookie
                </label>
                <input
                  type="text"
                  value={espnCreds.espn_s2}
                  onChange={(e) =>
                    setEspnCreds({ ...espnCreds, espn_s2: e.target.value })
                  }
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900"
                  placeholder="Long cookie string..."
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  SWID Cookie
                </label>
                <input
                  type="text"
                  value={espnCreds.espn_swid}
                  onChange={(e) =>
                    setEspnCreds({ ...espnCreds, espn_swid: e.target.value })
                  }
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900"
                  placeholder="{XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}"
                />
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    League ID
                  </label>
                  <input
                    type="number"
                    value={espnCreds.league_id || ""}
                    onChange={(e) =>
                      setEspnCreds({
                        ...espnCreds,
                        league_id: parseInt(e.target.value) || 0,
                      })
                    }
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Team ID
                  </label>
                  <input
                    type="number"
                    value={espnCreds.team_id || ""}
                    onChange={(e) =>
                      setEspnCreds({
                        ...espnCreds,
                        team_id: parseInt(e.target.value) || 0,
                      })
                    }
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Year
                  </label>
                  <input
                    type="number"
                    value={espnCreds.year}
                    onChange={(e) =>
                      setEspnCreds({
                        ...espnCreds,
                        year: parseInt(e.target.value) || 2025,
                      })
                    }
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900"
                  />
                </div>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={handleESPNConnect}
                  disabled={espnLoading}
                  className="inline-flex items-center gap-2 rounded-lg bg-orange-600 px-4 py-2 font-medium text-white transition hover:bg-orange-700 disabled:bg-orange-300"
                >
                  {espnLoading ? "Connecting..." : "Submit"}
                </button>
                <button
                  onClick={() => setShowESPNForm(false)}
                  className="inline-flex items-center gap-2 rounded-lg border border-gray-300 px-4 py-2 font-medium text-gray-700 transition hover:bg-gray-50"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {espnConnected && (
          <div className="mt-4">
            <div className="flex items-center justify-between mb-4">
              <p className="font-medium text-green-600">
                ESPN account connected
              </p>
              <div className="flex gap-2">
                <button
                  onClick={handleOptimizeLineup}
                  disabled={espnLoading}
                  className="inline-flex items-center gap-2 rounded-lg bg-orange-600 px-3 py-1.5 text-sm font-medium text-white transition hover:bg-orange-700 disabled:bg-orange-300"
                >
                  🎯 Optimize Lineup
                </button>
                <button
                  onClick={loadESPNData}
                  disabled={espnLoading}
                  className="inline-flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-gray-300"
                >
                  <RefreshCcw size={14} />
                  Refresh
                </button>
              </div>
            </div>

            {/* AI Start/Sit Advisor */}
            {(comparingPlayers.playerA || comparingPlayers.playerB) && (
              <div className="mb-4 rounded-lg bg-purple-50 border-2 border-purple-300 p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-lg font-semibold text-purple-900">
                    🤖 AI Start/Sit Advisor
                  </h3>
                  <button
                    onClick={resetComparison}
                    className="text-sm text-purple-700 hover:text-purple-900 underline"
                  >
                    Reset
                  </button>
                </div>
                <div className="grid grid-cols-2 gap-4 mb-3">
                  <div className={`p-3 rounded border-2 ${comparingPlayers.playerA ? 'border-purple-500 bg-white' : 'border-dashed border-purple-300 bg-purple-100'}`}>
                    {comparingPlayers.playerA ? (
                      <div>
                        <p className="font-semibold text-gray-900">{comparingPlayers.playerA.name}</p>
                        <p className="text-sm text-gray-600">{comparingPlayers.playerA.position} - {comparingPlayers.playerA.proTeam}</p>
                        <p className="text-sm text-purple-700">Proj: {comparingPlayers.playerA.projectedPoints.toFixed(1)} pts</p>
                      </div>
                    ) : (
                      <p className="text-sm text-gray-500 text-center">Select Player A</p>
                    )}
                  </div>
                  <div className={`p-3 rounded border-2 ${comparingPlayers.playerB ? 'border-purple-500 bg-white' : 'border-dashed border-purple-300 bg-purple-100'}`}>
                    {comparingPlayers.playerB ? (
                      <div>
                        <p className="font-semibold text-gray-900">{comparingPlayers.playerB.name}</p>
                        <p className="text-sm text-gray-600">{comparingPlayers.playerB.position} - {comparingPlayers.playerB.proTeam}</p>
                        <p className="text-sm text-purple-700">Proj: {comparingPlayers.playerB.projectedPoints.toFixed(1)} pts</p>
                      </div>
                    ) : (
                      <p className="text-sm text-gray-500 text-center">Select Player B</p>
                    )}
                  </div>
                </div>
                <button
                  onClick={handleGetAIAdvice}
                  disabled={!comparingPlayers.playerA || !comparingPlayers.playerB || espnLoading}
                  className="w-full rounded-lg bg-purple-600 px-4 py-2 font-medium text-white transition hover:bg-purple-700 disabled:bg-purple-300"
                >
                  {espnLoading ? "Getting AI Recommendation..." : "Get AI Recommendation"}
                </button>
              </div>
            )}

            {/* AI Recommendation Modal */}
            {showAIModal && aiAdvice && (
              <div className="mb-4 rounded-lg bg-gradient-to-r from-purple-50 to-blue-50 border-2 border-purple-400 p-6 shadow-lg">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-xl font-bold text-purple-900 mb-2">
                      🤖 AI Recommendation
                    </h3>
                    <div className="flex items-center gap-3 mb-3">
                      <div className="text-center">
                        <p className="text-sm text-gray-600">Start</p>
                        <p className="text-2xl font-bold text-purple-700">
                          {aiAdvice.recommendation === 'A' ? aiAdvice.playerAName : aiAdvice.playerBName}
                        </p>
                      </div>
                      <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
                        <div 
                          className="h-full bg-gradient-to-r from-purple-500 to-blue-500"
                          style={{ width: `${aiAdvice.confidence}%` }}
                        />
                      </div>
                      <div className="text-center">
                        <p className="text-sm text-gray-600">Confidence</p>
                        <p className="text-2xl font-bold text-blue-700">{aiAdvice.confidence}%</p>
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={() => setShowAIModal(false)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    ✕
                  </button>
                </div>
                <div className="bg-white rounded-lg p-4 border border-purple-200">
                  <p className="text-sm font-semibold text-gray-700 mb-2">Reasoning:</p>
                  <p className="text-gray-800">{aiAdvice.reasoning}</p>
                </div>
              </div>
            )}

            {espnLoading && (
              <p className="text-sm text-gray-500">Loading your roster...</p>
            )}

            {!espnLoading && espnRoster.length === 0 && (
              <p className="text-sm text-gray-500">
                No roster data available. Please check your credentials.
              </p>
            )}

            {showOptimized && optimizedLineup.length > 0 && (
              <div className="mb-6 rounded-lg bg-orange-50 border-2 border-orange-300 p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-lg font-semibold text-orange-900">
                    🎯 Optimized Lineup
                  </h3>
                  <div className="text-right">
                    <p className="text-xs text-orange-600">Total Projected</p>
                    <p className="text-2xl font-bold text-orange-900">
                      {totalProjected.toFixed(1)}
                    </p>
                  </div>
                </div>
                
                <div className="grid gap-2 md:grid-cols-2 lg:grid-cols-3">
                  {optimizedLineup.map((player, idx) => (
                    <div
                      key={idx}
                      className="rounded-lg border-2 border-orange-300 bg-white p-3"
                    >
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <p className="font-semibold text-gray-900">
                              {player.name}
                            </p>
                            {player.injured && (
                              <span className="text-xs font-medium text-red-600 bg-red-100 px-1.5 py-0.5 rounded">
                                {player.injuryStatus || "INJ"}
                              </span>
                            )}
                          </div>
                          <p className="text-sm text-gray-600">
                            {player.position} - {player.proTeam}
                          </p>
                        </div>
                        <span className="text-xs font-bold text-orange-700 bg-orange-200 px-2 py-1 rounded">
                          {player.recommendedSlot || player.lineupSlot}
                        </span>
                      </div>
                      
                      <div className="flex items-center justify-between pt-2 border-t border-gray-200">
                        <div className="text-center flex-1">
                          <p className="text-xs text-gray-500">Projected</p>
                          <p className="text-lg font-bold text-orange-600">
                            {player.projectedPoints.toFixed(1)}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>

                {benchPlayers.length > 0 && (
                  <div className="mt-4">
                    <p className="text-sm font-medium text-gray-700 mb-2">
                      Bench ({benchPlayers.length} players):
                    </p>
                    <div className="grid gap-2 grid-cols-2 md:grid-cols-4">
                      {benchPlayers.map((player, idx) => (
                        <div
                          key={idx}
                          className="rounded border border-gray-300 bg-gray-50 p-2 text-xs"
                        >
                          <p className="font-medium text-gray-900">{player.name}</p>
                          <p className="text-gray-600">{player.position} - {player.projectedPoints.toFixed(1)} pts</p>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                <button
                  onClick={() => setShowOptimized(false)}
                  className="mt-3 text-sm text-orange-700 hover:text-orange-900 underline"
                >
                  Hide Optimized Lineup
                </button>
              </div>
            )}

            {!espnLoading && espnRoster.length > 0 && (
              <div className="space-y-4">
                <p className="text-sm font-medium text-gray-700">
                  Your Current ESPN Roster ({espnRoster.length} players):
                </p>
                
                {/* Starters Section */}
                <div>
                  <p className="text-sm font-semibold text-gray-800 mb-3">
                    Starting Lineup:
                  </p>
                  <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                    {espnRoster
                      .filter((player) => player.lineupSlot !== "BE")
                      .map((player, idx) => (
                        <div
                          key={idx}
                          className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm hover:shadow-md transition"
                        >
                          <div className="flex items-start justify-between mb-2">
                            <div className="flex-1">
                              <div className="flex items-center gap-2">
                                <p className="font-semibold text-gray-900">
                                  {player.name}
                                </p>
                                {player.injured && (
                                  <span className="text-xs font-medium text-red-600 bg-red-100 px-1.5 py-0.5 rounded">
                                    {player.injuryStatus || "INJ"}
                                  </span>
                                )}
                              </div>
                              <p className="text-sm text-gray-600">
                                {player.position} - {player.proTeam}
                              </p>
                            </div>
                            <span className="text-xs font-medium text-orange-600 bg-orange-100 px-2 py-1 rounded">
                              {player.lineupSlot}
                            </span>
                          </div>
                          
                          <div className="flex items-center justify-between pt-2 border-t border-gray-100">
                            <div className="text-center flex-1">
                              <p className="text-xs text-gray-500">Projected</p>
                              <p className="text-lg font-bold text-orange-600">
                                {player.projectedPoints.toFixed(1)}
                              </p>
                            </div>
                            <div className="text-center flex-1 border-l border-gray-200">
                              <p className="text-xs text-gray-500">Actual</p>
                              <p className="text-lg font-bold text-gray-900">
                                {player.points.toFixed(1)}
                              </p>
                            </div>
                          </div>
                          
                          <button
                            onClick={() => handleSelectPlayerForComparison(player)}
                            className={`mt-2 w-full text-xs px-2 py-1 rounded transition ${
                              comparingPlayers.playerA === player || comparingPlayers.playerB === player
                                ? 'bg-purple-600 text-white'
                                : 'bg-purple-100 text-purple-700 hover:bg-purple-200'
                            }`}
                          >
                            {comparingPlayers.playerA === player ? '✓ Player A' : 
                             comparingPlayers.playerB === player ? '✓ Player B' : 
                             '🤖 Compare'}
                          </button>
                        </div>
                      ))}
                  </div>
                </div>

                {/* Bench Section */}
                {espnRoster.filter((player) => player.lineupSlot === "BE").length > 0 && (
                  <div>
                    <p className="text-sm font-semibold text-gray-700 mb-3">
                      Bench ({espnRoster.filter((player) => player.lineupSlot === "BE").length} players):
                    </p>
                    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                      {espnRoster
                        .filter((player) => player.lineupSlot === "BE")
                        .map((player, idx) => (
                          <div
                            key={idx}
                            className="rounded-lg border border-gray-200 bg-gray-50 p-3 shadow-sm hover:shadow-md transition"
                          >
                            <div className="flex items-start justify-between mb-2">
                              <div className="flex-1">
                                <div className="flex items-center gap-2">
                                  <p className="font-semibold text-gray-900">{player.name}</p>
                                  {player.injured && (
                                    <span className="text-xs font-medium text-red-600 bg-red-100 px-1.5 py-0.5 rounded">
                                      {player.injuryStatus || "INJ"}
                                    </span>
                                  )}
                                </div>
                                <p className="text-sm text-gray-600">
                                  {player.position} - {player.proTeam}
                                </p>
                              </div>
                              <span className="text-xs font-medium text-gray-600 bg-gray-200 px-2 py-1 rounded">
                                BE
                              </span>
                            </div>
                            
                            <div className="flex items-center justify-between pt-2 border-t border-gray-200">
                              <div className="text-center flex-1">
                                <p className="text-xs text-gray-500">Projected</p>
                                <p className="text-lg font-bold text-orange-600">
                                  {player.projectedPoints.toFixed(1)}
                                </p>
                              </div>
                              <div className="text-center flex-1 border-l border-gray-200">
                                <p className="text-xs text-gray-500">Actual</p>
                                <p className="text-lg font-bold text-gray-900">
                                  {player.points.toFixed(1)}
                                </p>
                              </div>
                            </div>
                            
                            <button
                              onClick={() => handleSelectPlayerForComparison(player)}
                              className={`mt-2 w-full text-xs px-2 py-1 rounded transition ${
                                comparingPlayers.playerA === player || comparingPlayers.playerB === player
                                  ? 'bg-purple-600 text-white'
                                  : 'bg-purple-100 text-purple-700 hover:bg-purple-200'
                              }`}
                            >
                              {comparingPlayers.playerA === player ? '✓ Player A' : 
                               comparingPlayers.playerB === player ? '✓ Player B' : 
                               '🤖 Compare'}
                            </button>
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Free Agents Section */}
            {!espnLoading && espnConnected && (
              <div className="mt-6 pt-6 border-t border-gray-200">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-semibold text-gray-800">
                    Available Free Agents
                  </h3>
                  <div className="flex gap-2 items-center">
                    <select
                      value={selectedPosition}
                      onChange={(e) => setSelectedPosition(e.target.value)}
                      className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-900"
                    >
                      <option value="">All Positions</option>
                      <option value="QB">QB</option>
                      <option value="RB">RB</option>
                      <option value="WR">WR</option>
                      <option value="TE">TE</option>
                      <option value="K">K</option>
                      <option value="D/ST">D/ST</option>
                    </select>
                    <button
                      onClick={() => handleLoadFreeAgents(selectedPosition || undefined)}
                      disabled={espnLoading}
                      className="inline-flex items-center gap-2 rounded-lg bg-green-600 px-3 py-1.5 text-sm font-medium text-white transition hover:bg-green-700 disabled:bg-green-300"
                    >
                      🔍 Search Free Agents
                    </button>
                  </div>
                </div>

                {showFreeAgents && freeAgents.length > 0 && (
                  <div>
                    <p className="text-sm text-gray-600 mb-3">
                      Showing top {freeAgents.length} available players
                      {selectedPosition && ` at ${selectedPosition}`}
                    </p>
                    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3 max-h-96 overflow-y-auto">
                      {freeAgents.map((player, idx) => (
                        <div
                          key={idx}
                          className="rounded-lg border border-green-200 bg-green-50 p-3 hover:shadow-md transition"
                        >
                          <div className="flex items-start justify-between mb-2">
                            <div className="flex-1">
                              <div className="flex items-center gap-2">
                                <p className="font-semibold text-gray-900">
                                  {player.name}
                                </p>
                                {player.injured && (
                                  <span className="text-xs font-medium text-red-600 bg-red-100 px-1.5 py-0.5 rounded">
                                    {player.injuryStatus || "INJ"}
                                  </span>
                                )}
                              </div>
                              <p className="text-sm text-gray-600">
                                {player.position} - {player.proTeam}
                              </p>
                            </div>
                          </div>
                          
                          <div className="flex items-center justify-between pt-2 border-t border-green-200">
                            <div className="text-center flex-1">
                              <p className="text-xs text-gray-500">Projected</p>
                              <p className="text-base font-bold text-green-700">
                                {player.projectedPoints.toFixed(1)}
                              </p>
                            </div>
                            <div className="text-center flex-1 border-l border-green-200">
                              <p className="text-xs text-gray-500">% Owned</p>
                              <p className="text-base font-bold text-gray-700">
                                {player.percentOwned.toFixed(0)}%
                              </p>
                            </div>
                            <div className="text-center flex-1 border-l border-green-200">
                              <p className="text-xs text-gray-500">% Started</p>
                              <p className="text-base font-bold text-gray-700">
                                {player.percentStarted.toFixed(0)}%
                              </p>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                    <button
                      onClick={() => setShowFreeAgents(false)}
                      className="mt-3 text-sm text-green-700 hover:text-green-900 underline"
                    >
                      Hide Free Agents
                    </button>
                  </div>
                )}
              </div>
            )}

            {/* Waiver Wire AI Scout Section */}
            {!espnLoading && espnConnected && (
              <div className="mt-6 pt-6 border-t border-gray-200">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-800 flex items-center gap-2">
                      🔮 Waiver Wire AI Scout
                      {espnRoster && espnRoster.length > 0 && (
                        <span className="text-xs font-normal text-purple-600 bg-purple-50 px-2 py-1 rounded-full">
                          Personalized for Your Team
                        </span>
                      )}
                    </h3>
                    <p className="text-sm text-gray-600 mt-1">
                      {espnRoster && espnRoster.length > 0 
                        ? "AI recommendations tailored to your roster needs"
                        : "AI-powered breakout candidate finder"}
                    </p>
                  </div>
                  <div className="flex gap-2 items-center">
                    <select
                      value={waiverPosition}
                      onChange={(e) => setWaiverPosition(e.target.value)}
                      className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm"
                      disabled={waiverLoading}
                    >
                      <option value="ALL">All Positions</option>
                      <option value="QB">QB</option>
                      <option value="RB">RB</option>
                      <option value="WR">WR</option>
                      <option value="TE">TE</option>
                    </select>
                    <button
                      onClick={() => loadWaiverGems(waiverPosition)}
                      disabled={waiverLoading}
                      className="rounded-lg bg-gradient-to-r from-indigo-600 to-purple-600 px-4 py-2 text-sm font-medium text-white hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50"
                    >
                      {waiverLoading ? "Analyzing..." : "🔍 Find Hidden Gems"}
                    </button>
                  </div>
                </div>

                {showWaiverGems && waiverGems && waiverGems.length > 0 && (
                  <div className="space-y-3">
                    <p className="text-sm text-gray-700">
                      <strong>{waiverGems.length} breakout candidates</strong> found with AI analysis:
                    </p>
                    <div className="grid gap-4">
                      {waiverGems.map((gem, idx) => (
                        <div
                          key={idx}
                          className="rounded-xl border-2 border-indigo-200 bg-gradient-to-br from-indigo-50 to-purple-50 p-4 shadow-sm hover:shadow-md transition"
                        >
                          {/* Header */}
                          <div className="flex items-start justify-between mb-3">
                            <div className="flex-1">
                              <div className="flex items-center gap-3 mb-1">
                                <h4 className="text-lg font-bold text-gray-900">{gem.playerName}</h4>
                                <span className="text-xs font-semibold text-indigo-700 bg-indigo-100 px-2 py-1 rounded">
                                  {gem.position}
                                </span>
                                <span className="text-xs text-gray-600">{gem.team}</span>
                              </div>
                              <div className="flex items-center gap-2">
                                <span className="text-sm font-medium text-purple-700">
                                  {gem.recommendation}
                                </span>
                                {gem.trendingUp && (
                                  <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded">
                                    📈 Trending Up
                                  </span>
                                )}
                              </div>
                            </div>
                            <div className="text-right">
                              <p className="text-xs text-gray-500">Breakout Score</p>
                              <p className="text-3xl font-bold text-indigo-600">{gem.breakoutScore.toFixed(0)}</p>
                              <p className="text-xs text-gray-500">/100</p>
                            </div>
                          </div>

                          {/* Key Metrics Grid */}
                          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-3 p-3 bg-white/60 rounded-lg">
                            <div className="text-center">
                              <p className="text-xs text-gray-500">Snap %</p>
                              <p className="text-lg font-bold text-gray-900">
                                {gem.snapCountPct > 0 ? `${gem.snapCountPct.toFixed(0)}%` : 'N/A'}
                              </p>
                            </div>
                            <div className="text-center">
                              <p className="text-xs text-gray-500">EPA/Play</p>
                              <p className="text-lg font-bold text-gray-900">
                                {gem.epaPerPlay > 0 ? gem.epaPerPlay.toFixed(3) : 'N/A'}
                              </p>
                            </div>
                            <div className="text-center">
                              <p className="text-xs text-gray-500">Target Trend</p>
                              <p className="text-sm font-semibold text-gray-900 capitalize">{gem.targetShareTrend}</p>
                            </div>
                            <div className="text-center">
                              <p className="text-xs text-gray-500">Schedule</p>
                              <p className="text-sm font-semibold text-gray-900 capitalize">{gem.upcomingSchedule}</p>
                            </div>
                          </div>

                          {/* Opportunity Badge */}
                          {gem.depthChartStatus && gem.depthChartStatus !== "Normal role" && (
                            <div className="mb-3 p-2 bg-yellow-50 border border-yellow-200 rounded-lg">
                              <p className="text-sm text-yellow-800">
                                ⚡ <strong>Opportunity:</strong> {gem.depthChartStatus}
                              </p>
                            </div>
                          )}

                          {/* Recent Games */}
                          {gem.lastThreeGames && gem.lastThreeGames.length > 0 && (
                            <div className="mb-3">
                              <p className="text-xs font-medium text-gray-700 mb-2">Last 3 Games:</p>
                              <div className="space-y-1">
                                {gem.lastThreeGames.slice(0, 3).map((game, gIdx) => (
                                  <div key={gIdx} className="text-xs text-gray-600 bg-white/50 rounded px-2 py-1">
                                    <span className="font-medium">Week {game.week} vs {game.opponent}:</span> {game.production} ({game.fantasyPoints.toFixed(1)} pts)
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          {/* AI Analysis */}
                          {gem.aiAnalysis && gem.aiAnalysis !== "AI analysis unavailable" && (
                            <div className="p-3 bg-indigo-100 border border-indigo-300 rounded-lg">
                              <p className="text-xs font-semibold text-indigo-900 mb-1">🤖 AI Analysis:</p>
                              <p className="text-sm text-indigo-800">{gem.aiAnalysis}</p>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                    <button
                      onClick={() => setShowWaiverGems(false)}
                      className="mt-3 text-sm text-indigo-700 hover:text-indigo-900 underline"
                    >
                      Hide Waiver Wire Gems
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  );
}
