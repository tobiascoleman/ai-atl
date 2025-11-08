"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Link2, RefreshCcw, Trophy } from "lucide-react";

import {
  fetchFantasyStatus,
  fetchFantasyTeams,
  getYahooAuthUrl,
} from "@/lib/api/fantasy";
import { FantasyStatusResponse, YahooTeam } from "@/types/api";

export default function FantasyPage() {
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<FantasyStatusResponse | null>(null);
  const [teams, setTeams] = useState<YahooTeam[]>([]);
  const [loading, setLoading] = useState(true);
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

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

  useEffect(() => {
    loadFantasyData();
  }, [loadFantasyData]);

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
    </div>
  );
}
