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
  ESPNPlayer,
  ESPNCredentials,
} from "@/lib/api/espn";
import { FantasyStatusResponse, YahooTeam } from "@/types/api";

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
  const [espnCreds, setEspnCreds] = useState<ESPNCredentials>({
    espn_s2: "",
    espn_swid: "",
    league_id: 0,
    team_id: 0,
    year: 2025,
  });

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
                  className="w-full rounded-lg border border-gray-300 px-3 py-2"
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
                  className="w-full rounded-lg border border-gray-300 px-3 py-2"
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
                    className="w-full rounded-lg border border-gray-300 px-3 py-2"
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
                    className="w-full rounded-lg border border-gray-300 px-3 py-2"
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
                    className="w-full rounded-lg border border-gray-300 px-3 py-2"
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
              <button
                onClick={loadESPNData}
                disabled={espnLoading}
                className="inline-flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-gray-300"
              >
                <RefreshCcw size={14} />
                Refresh
              </button>
            </div>

            {espnLoading && (
              <p className="text-sm text-gray-500">Loading your roster...</p>
            )}

            {!espnLoading && espnRoster.length === 0 && (
              <p className="text-sm text-gray-500">
                No roster data available. Please check your credentials.
              </p>
            )}

            {!espnLoading && espnRoster.length > 0 && (
              <div className="space-y-2">
                <p className="text-sm font-medium text-gray-700">
                  Your ESPN Roster ({espnRoster.length} players):
                </p>
                <div className="grid gap-2 md:grid-cols-2 lg:grid-cols-3">
                  {espnRoster.map((player, idx) => (
                    <div
                      key={idx}
                      className="rounded-lg border border-gray-200 bg-gray-50 p-3"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <p className="font-medium text-gray-900">
                            {player.name}
                          </p>
                          <p className="text-sm text-gray-600">
                            {player.position} - {player.proTeam}
                          </p>
                        </div>
                        <span className="text-xs font-medium text-orange-600 bg-orange-100 px-2 py-1 rounded">
                          {player.lineupSlot}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  );
}
