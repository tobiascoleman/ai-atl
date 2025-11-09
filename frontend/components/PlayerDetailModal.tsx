"use client";

import { useEffect, useState } from "react";
import { X, TrendingUp, Activity, Award, AlertCircle } from "lucide-react";
import { Player } from "@/types/api";
import apiClient from "@/lib/api/client";

interface PlayerDetailModalProps {
  player: Player;
  isOpen: boolean;
  onClose: () => void;
}

interface PlayerDetailData {
  player: Player;
  all_seasons: Player[];
  stats: any[];
  all_stats: any[];
  all_ngs: any[];
  weekly_stats: any[];
  all_weekly_stats: any[];
  epa: number;
  ngs: any[];
  play_count: number;
  epa_by_season: { [season: number]: { epa: number; play_count: number } };
  lifetime_epa: number;
  lifetime_plays: number;
}

type ViewMode = "lifetime" | "season" | "week";

export default function PlayerDetailModal({
  player,
  isOpen,
  onClose,
}: PlayerDetailModalProps) {
  const [loading, setLoading] = useState(true);
  const [detailData, setDetailData] = useState<PlayerDetailData | null>(null);
  const [error, setError] = useState("");
  const [viewMode, setViewMode] = useState<ViewMode>("season");
  const [selectedSeason, setSelectedSeason] = useState<number>(player.season);

  useEffect(() => {
    if (isOpen && player.nfl_id) {
      fetchPlayerDetails();
    }
  }, [isOpen, player.nfl_id]);

  const fetchPlayerDetails = async () => {
    setLoading(true);
    setError("");
    try {
      console.log("🔍 Full player object:", player);
      console.log(
        `🔍 Fetching player details: nfl_id="${player.nfl_id}", season=${player.season}`
      );

      if (!player.nfl_id) {
        throw new Error("Player NFL ID is missing");
      }

      // Fetch comprehensive player data using the API client (which points to localhost:8080)
      // Pass season=0 to get ALL seasons data
      const { data } = await apiClient.get(
        `/data/players/${player.nfl_id}/summary?season=0`
      );

      console.log("✅ Player details loaded:", data);
      console.log("📊 EPA by season:", data.epa_by_season);
      console.log(
        "📈 Lifetime EPA:",
        data.lifetime_epa,
        "Lifetime plays:",
        data.lifetime_plays
      );
      console.log("🔢 All NGS count:", data.all_ngs?.length);
      setDetailData(data);
    } catch (err: any) {
      console.error("❌ Error fetching player details:", err);
      const errorMessage =
        err.response?.data?.error ||
        err.message ||
        "Failed to load player details";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div
          className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
          onClick={onClose}
        />

        {/* Modal */}
        <div className="relative bg-white rounded-2xl shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
          {/* Header */}
          <div className="sticky top-0 bg-gradient-to-r from-blue-600 to-indigo-600 text-white p-6 rounded-t-2xl">
            <button
              onClick={onClose}
              className="absolute top-4 right-4 p-2 hover:bg-white/20 rounded-full transition"
              aria-label="Close player details"
            >
              <X size={24} />
            </button>

            <div className="flex items-start gap-4">
              <div className="flex-1">
                <h2 className="text-3xl font-bold mb-2">{player.name}</h2>
                <div className="flex items-center gap-4 text-blue-100">
                  <span className="px-3 py-1 bg-white/20 rounded-full font-semibold">
                    {player.position}
                  </span>
                  <span className="text-lg">{player.team}</span>
                  <span>Season {player.season}</span>
                </div>
              </div>

              {player.status_description && (
                <div className="text-right">
                  <div
                    className={`px-3 py-1 rounded-full text-sm font-semibold ${
                      player.status_description.includes("Injured")
                        ? "bg-red-500"
                        : player.status_description.includes("Active")
                        ? "bg-green-500"
                        : "bg-blue-500"
                    }`}
                  >
                    {player.status_description}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Tabs and Season Selector */}
          {!loading && !error && detailData && (
            <div className="border-b bg-gray-50 px-6 py-3">
              <div className="flex items-center justify-between">
                {/* View Mode Tabs */}
                <div className="flex gap-2">
                  <button
                    onClick={() => setViewMode("lifetime")}
                    className={`px-4 py-2 rounded-lg font-medium transition ${
                      viewMode === "lifetime"
                        ? "bg-blue-600 text-white"
                        : "bg-white text-gray-700 hover:bg-gray-100"
                    }`}
                  >
                    Lifetime
                  </button>
                  <button
                    onClick={() => setViewMode("season")}
                    className={`px-4 py-2 rounded-lg font-medium transition ${
                      viewMode === "season"
                        ? "bg-blue-600 text-white"
                        : "bg-white text-gray-700 hover:bg-gray-100"
                    }`}
                  >
                    Season
                  </button>
                  <button
                    onClick={() => setViewMode("week")}
                    className={`px-4 py-2 rounded-lg font-medium transition ${
                      viewMode === "week"
                        ? "bg-blue-600 text-white"
                        : "bg-white text-gray-700 hover:bg-gray-100"
                    }`}
                  >
                    Week
                  </button>
                </div>

                {/* Season Selector */}
                {viewMode !== "lifetime" &&
                  detailData.all_seasons &&
                  detailData.all_seasons.length > 0 && (
                    <div className="flex items-center gap-2">
                      <label
                        htmlFor="season-selector"
                        className="text-sm font-medium text-gray-700"
                      >
                        Season:
                      </label>
                      <select
                        id="season-selector"
                        value={selectedSeason}
                        onChange={(e) =>
                          setSelectedSeason(parseInt(e.target.value))
                        }
                        className="px-3 py-1 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 text-gray-900"
                      >
                        {detailData.all_seasons.map((season: Player) => (
                          <option key={season.season} value={season.season}>
                            {season.season} - {season.team}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}
              </div>
            </div>
          )}

          {/* Content */}
          <div className="p-6">
            {loading ? (
              <div className="flex items-center justify-center py-12">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
              </div>
            ) : error ? (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
                <AlertCircle className="inline mr-2" size={20} />
                {error}
              </div>
            ) : detailData ? (
              <div className="space-y-6">
                {viewMode === "lifetime" && (
                  <LifetimeView detailData={detailData} player={player} />
                )}

                {viewMode === "season" && (
                  <SeasonView
                    detailData={detailData}
                    player={player}
                    selectedSeason={selectedSeason}
                  />
                )}

                {viewMode === "week" && (
                  <WeekView
                    detailData={detailData}
                    player={player}
                    selectedSeason={selectedSeason}
                  />
                )}
              </div>
            ) : (
              <div className="text-center py-12 text-gray-500">
                No detailed data available
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

// Lifetime View - Aggregates all seasons
function LifetimeView({
  detailData,
  player,
}: {
  detailData: PlayerDetailData;
  player: Player;
}) {
  // Aggregate lifetime stats
  const lifetimeStats =
    detailData.all_stats?.reduce((acc, stat) => {
      return {
        passing_yards: (acc.passing_yards || 0) + (stat.passing_yards || 0),
        passing_tds: (acc.passing_tds || 0) + (stat.passing_tds || 0),
        interceptions: (acc.interceptions || 0) + (stat.interceptions || 0),
        rushing_yards: (acc.rushing_yards || 0) + (stat.rushing_yards || 0),
        rushing_tds: (acc.rushing_tds || 0) + (stat.rushing_tds || 0),
        receptions: (acc.receptions || 0) + (stat.receptions || 0),
        receiving_yards:
          (acc.receiving_yards || 0) + (stat.receiving_yards || 0),
        receiving_tds: (acc.receiving_tds || 0) + (stat.receiving_tds || 0),
        tackles: (acc.tackles || 0) + (stat.tackles || 0),
        sacks: (acc.sacks || 0) + (stat.sacks || 0),
        def_interceptions:
          (acc.def_interceptions || 0) + (stat.def_interceptions || 0),
      };
    }, {} as any) || {};

  const seasonsPlayed = detailData.all_seasons?.length || 0;

  return (
    <>
      {/* Career Summary */}
      <div className="bg-gradient-to-r from-indigo-50 to-purple-50 rounded-xl p-6">
        <div className="flex items-center gap-2 mb-4">
          <Award className="text-indigo-600" size={24} />
          <h3 className="text-xl font-bold text-gray-900">Career Summary</h3>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <MetricCard label="Seasons Played" value={seasonsPlayed.toString()} />
          <MetricCard
            label="Teams"
            value={[
              ...new Set(detailData.all_seasons?.map((s) => s.team)),
            ].join(", ")}
          />
        </div>
      </div>

      {/* Lifetime Performance Metrics */}
      {detailData.lifetime_plays > 0 && (
        <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp className="text-green-600" size={24} />
            <h3 className="text-xl font-bold text-gray-900">
              Career Performance Metrics
            </h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <MetricCard
              label="Career Avg EPA"
              value={detailData.lifetime_epa.toFixed(3)}
              description="Expected Points Added per play"
              color={detailData.lifetime_epa > 0 ? "green" : "red"}
            />
            <MetricCard
              label="Total Career Plays"
              value={detailData.lifetime_plays.toLocaleString()}
              description="Plays involved in across all seasons"
            />
          </div>
        </div>
      )}

      {/* Lifetime Stats */}
      <div className="bg-gray-50 rounded-xl p-6">
        <div className="flex items-center gap-2 mb-4">
          <Activity size={24} />
          <h3 className="text-xl font-bold text-gray-900">
            Lifetime Statistics
          </h3>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {player.position === "QB" && (
            <>
              <StatCard
                label="Passing Yards"
                value={lifetimeStats.passing_yards?.toLocaleString() || "0"}
              />
              <StatCard
                label="Passing TDs"
                value={lifetimeStats.passing_tds || 0}
              />
              <StatCard
                label="Interceptions"
                value={lifetimeStats.interceptions || 0}
              />
              <StatCard
                label="Rushing Yards"
                value={lifetimeStats.rushing_yards?.toLocaleString() || "0"}
              />
            </>
          )}

          {player.position === "RB" && (
            <>
              <StatCard
                label="Rushing Yards"
                value={lifetimeStats.rushing_yards?.toLocaleString() || "0"}
              />
              <StatCard
                label="Rushing TDs"
                value={lifetimeStats.rushing_tds || 0}
              />
              <StatCard
                label="Receptions"
                value={lifetimeStats.receptions || 0}
              />
              <StatCard
                label="Receiving Yards"
                value={lifetimeStats.receiving_yards?.toLocaleString() || "0"}
              />
              <StatCard
                label="Receiving TDs"
                value={lifetimeStats.receiving_tds || 0}
              />
            </>
          )}

          {(player.position === "WR" || player.position === "TE") && (
            <>
              <StatCard
                label="Receptions"
                value={lifetimeStats.receptions || 0}
              />
              <StatCard
                label="Receiving Yards"
                value={lifetimeStats.receiving_yards?.toLocaleString() || "0"}
              />
              <StatCard
                label="Receiving TDs"
                value={lifetimeStats.receiving_tds || 0}
              />
            </>
          )}

          {[
            "LB",
            "DE",
            "DT",
            "CB",
            "S",
            "ILB",
            "OLB",
            "MLB",
            "NT",
            "SS",
            "FS",
            "DB",
            "DL",
          ].includes(player.position) && (
            <>
              <StatCard label="Tackles" value={lifetimeStats.tackles || 0} />
              <StatCard
                label="Sacks"
                value={lifetimeStats.sacks?.toFixed(1) || "0"}
              />
              <StatCard
                label="Interceptions"
                value={lifetimeStats.def_interceptions || 0}
              />
            </>
          )}
        </div>
      </div>

      {/* Season-by-Season Breakdown */}
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <div className="p-6 border-b">
          <h3 className="text-xl font-bold text-gray-900">
            Season-by-Season Breakdown
          </h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Season
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Team
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Stats
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  EPA
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {detailData.all_seasons?.map((season) => {
                const seasonStats = detailData.all_stats?.find(
                  (s) =>
                    s.season === season.season && s.season_type === "REGPOST"
                );
                const seasonEPA = detailData.epa_by_season?.[season.season];
                return (
                  <tr key={season.season} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">
                      {season.season}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {season.team}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {seasonStats
                        ? player.position === "QB"
                          ? `${seasonStats.passing_yards?.toLocaleString()} pass yds, ${
                              seasonStats.passing_tds
                            } TDs`
                          : player.position === "RB"
                          ? `${seasonStats.rushing_yards?.toLocaleString()} rush yds, ${
                              seasonStats.rushing_tds
                            } TDs`
                          : player.position === "WR" || player.position === "TE"
                          ? `${
                              seasonStats.receptions
                            } rec, ${seasonStats.receiving_yards?.toLocaleString()} yds`
                          : "No stats"
                        : "No stats"}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {seasonEPA && seasonEPA.play_count > 0 ? (
                        <span
                          className={`font-medium ${
                            seasonEPA.epa > 0
                              ? "text-green-600"
                              : "text-red-600"
                          }`}
                        >
                          {seasonEPA.epa.toFixed(3)}
                        </span>
                      ) : (
                        <span className="text-gray-400">N/A</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

// Season View - Shows selected season
function SeasonView({
  detailData,
  player,
  selectedSeason,
}: {
  detailData: PlayerDetailData;
  player: Player;
  selectedSeason: number;
}) {
  const seasonData =
    detailData.all_seasons?.find((s) => s.season === selectedSeason) ||
    detailData.player;
  const seasonStats = detailData.all_stats?.find(
    (s) => s.season === selectedSeason && s.season_type === "REGPOST"
  );
  const seasonNGS = detailData.all_ngs?.filter(
    (n) => n.season === selectedSeason
  );

  const enrichedPlayer = {
    ...seasonData,
    passing_yards: seasonStats?.passing_yards,
    passing_tds: seasonStats?.passing_tds,
    rushing_yards: seasonStats?.rushing_yards,
    rushing_tds: seasonStats?.rushing_tds,
    receiving_yards: seasonStats?.receiving_yards,
    receiving_tds: seasonStats?.receiving_tds,
    receptions: seasonStats?.receptions,
    tackles: seasonStats?.tackles,
    sacks: seasonStats?.sacks,
    tackles_for_loss: seasonStats?.tackles_for_loss,
    def_interceptions: seasonStats?.def_interceptions,
    pass_defended: seasonStats?.pass_defended,
  };

  return (
    <>
      {/* Season Stats */}
      <StatSection
        title={`${selectedSeason} Season Statistics`}
        icon={<Activity />}
        player={enrichedPlayer as Player}
      />

      {/* Performance Metrics */}
      {detailData.epa_by_season && detailData.epa_by_season[selectedSeason] && (
        <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp className="text-green-600" size={24} />
            <h3 className="text-xl font-bold text-gray-900">
              {selectedSeason} Performance Metrics
            </h3>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <MetricCard
              label="Season Avg EPA"
              value={detailData.epa_by_season[selectedSeason].epa.toFixed(3)}
              description="Expected Points Added per play"
              color={
                detailData.epa_by_season[selectedSeason].epa > 0
                  ? "green"
                  : "red"
              }
            />
            <MetricCard
              label="Total Plays"
              value={detailData.epa_by_season[
                selectedSeason
              ].play_count.toLocaleString()}
              description="Plays involved in this season"
            />
          </div>
        </div>
      )}

      {/* Next Gen Stats */}
      {seasonNGS && seasonNGS.length > 0 && (
        <div className="bg-purple-50 rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <Award className="text-purple-600" size={24} />
            <h3 className="text-xl font-bold text-gray-900">Next Gen Stats</h3>
          </div>
          <NextGenStatsDisplay ngs={seasonNGS} position={player.position} />
        </div>
      )}
    </>
  );
}

// Week View - Shows week-by-week breakdown
function WeekView({
  detailData,
  player,
  selectedSeason,
}: {
  detailData: PlayerDetailData;
  player: Player;
  selectedSeason: number;
}) {
  // Get weekly stats for selected season
  const weeklyStats =
    detailData.all_weekly_stats?.filter((w) => w.season === selectedSeason) ||
    [];

  const isQB = player.position === "QB";
  const isRB = player.position === "RB";
  const isWRTE = ["WR", "TE"].includes(player.position);

  return (
    <>
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <div className="p-6 border-b">
          <div className="flex items-center gap-2">
            <Activity size={24} />
            <h3 className="text-xl font-bold text-gray-900">
              {selectedSeason} Week-by-Week Stats
            </h3>
          </div>
        </div>

        {weeklyStats.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Week
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    vs
                  </th>
                  {isQB && (
                    <>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Pass Yds
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Pass TDs
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        INTs
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rush Yds
                      </th>
                    </>
                  )}
                  {isRB && (
                    <>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rush Yds
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rush TDs
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rec
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rec Yds
                      </th>
                    </>
                  )}
                  {isWRTE && (
                    <>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rec
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Tgts
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rec Yds
                      </th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                        Rec TDs
                      </th>
                    </>
                  )}
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                    PPR
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                    EPA
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {weeklyStats
                  .sort((a, b) => a.week - b.week)
                  .map((week) => (
                    <tr key={week.week} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm font-medium text-gray-900">
                        Week {week.week}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700">
                        {week.opponent || "-"}
                      </td>
                      {isQB && (
                        <>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.passing_yards || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.passing_tds || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.interceptions || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.rushing_yards || 0}
                          </td>
                        </>
                      )}
                      {isRB && (
                        <>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.rushing_yards || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.rushing_tds || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.receptions || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.receiving_yards || 0}
                          </td>
                        </>
                      )}
                      {isWRTE && (
                        <>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.receptions || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.targets || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.receiving_yards || 0}
                          </td>
                          <td className="px-4 py-3 text-sm text-right">
                            {week.receiving_tds || 0}
                          </td>
                        </>
                      )}
                      <td className="px-4 py-3 text-sm text-right font-semibold">
                        {week.fantasy_points_ppr
                          ? week.fantasy_points_ppr.toFixed(1)
                          : "0.0"}
                      </td>
                      <td className="px-4 py-3 text-sm text-right">
                        <span
                          className={
                            week.epa >= 0 ? "text-green-600" : "text-red-600"
                          }
                        >
                          {week.epa ? week.epa.toFixed(2) : "0.00"}
                        </span>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="p-12 text-center text-gray-500">
            No weekly data available for {selectedSeason}
          </div>
        )}
      </div>
    </>
  );
}

function StatSection({
  title,
  icon,
  player,
}: {
  title: string;
  icon: React.ReactNode;
  player: Player;
}) {
  const defensivePositions = [
    "LB",
    "DE",
    "DT",
    "CB",
    "S",
    "ILB",
    "OLB",
    "MLB",
    "NT",
    "SS",
    "FS",
    "DB",
    "DL",
  ];
  const isDefensive = defensivePositions.includes(player.position);

  return (
    <div className="bg-gray-50 rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        {icon}
        <h3 className="text-xl font-bold text-gray-900">{title}</h3>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {player.position === "QB" && (
          <>
            <StatCard
              label="Passing Yards"
              value={player.passing_yards?.toLocaleString() || "0"}
            />
            <StatCard label="Passing TDs" value={player.passing_tds || 0} />
            <StatCard
              label="Rushing Yards"
              value={player.rushing_yards?.toLocaleString() || "0"}
            />
            <StatCard label="Rushing TDs" value={player.rushing_tds || 0} />
          </>
        )}

        {player.position === "RB" && (
          <>
            <StatCard
              label="Rushing Yards"
              value={player.rushing_yards?.toLocaleString() || "0"}
            />
            <StatCard label="Rushing TDs" value={player.rushing_tds || 0} />
            <StatCard label="Receptions" value={player.receptions || 0} />
            <StatCard
              label="Receiving Yards"
              value={player.receiving_yards?.toLocaleString() || "0"}
            />
            <StatCard label="Receiving TDs" value={player.receiving_tds || 0} />
          </>
        )}

        {(player.position === "WR" || player.position === "TE") && (
          <>
            <StatCard label="Receptions" value={player.receptions || 0} />
            <StatCard
              label="Receiving Yards"
              value={player.receiving_yards?.toLocaleString() || "0"}
            />
            <StatCard label="Receiving TDs" value={player.receiving_tds || 0} />
          </>
        )}

        {isDefensive && (
          <>
            <StatCard label="Tackles" value={player.tackles || 0} />
            <StatCard label="Solo Tackles" value={player.tackles_solo || 0} />
            <StatCard label="Sacks" value={player.sacks || 0} />
            <StatCard label="TFL" value={player.tackles_for_loss || 0} />
            <StatCard
              label="Interceptions"
              value={player.def_interceptions || 0}
            />
            <StatCard label="Pass Defended" value={player.pass_defended || 0} />
            <StatCard
              label="Forced Fumbles"
              value={player.forced_fumbles || 0}
            />
            <StatCard
              label="Fumble Recoveries"
              value={player.fumble_recoveries || 0}
            />
          </>
        )}
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="bg-white rounded-lg p-4 shadow-sm">
      <p className="text-sm text-gray-600 mb-1">{label}</p>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
    </div>
  );
}

function MetricCard({
  label,
  value,
  description,
  color = "blue",
}: {
  label: string;
  value: string;
  description?: string;
  color?: "blue" | "green" | "red";
}) {
  const colorClasses = {
    blue: "bg-blue-100 text-blue-900",
    green: "bg-green-100 text-green-900",
    red: "bg-red-100 text-red-900",
  };

  return (
    <div className={`rounded-lg p-4 ${colorClasses[color]}`}>
      <p className="text-sm opacity-75 mb-1">{label}</p>
      <p className="text-3xl font-bold mb-1">{value}</p>
      {description && <p className="text-xs opacity-75">{description}</p>}
    </div>
  );
}

function NextGenStatsDisplay({
  ngs,
  position,
}: {
  ngs: any[];
  position: string;
}) {
  if (!ngs || ngs.length === 0)
    return <p className="text-gray-500">No Next Gen Stats available</p>;

  // Get the most recent NGS entry
  const latestNGS = ngs[0];

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
      {position === "QB" && latestNGS.stat_type === "passing" && (
        <>
          <StatCard
            label="Avg Time to Throw"
            value={`${latestNGS.avg_time_to_throw?.toFixed(2)}s` || "N/A"}
          />
          <StatCard
            label="Avg Intended Air Yds"
            value={
              `${latestNGS.avg_intended_air_yards?.toFixed(1)} yds` || "N/A"
            }
          />
          <StatCard
            label="Avg Completed Air Yds"
            value={
              `${latestNGS.avg_completed_air_yards?.toFixed(1)} yds` || "N/A"
            }
          />
          <StatCard
            label="CPOE"
            value={
              `${latestNGS.completion_percentage_above_expectation?.toFixed(
                1
              )}%` || "N/A"
            }
          />
          <StatCard
            label="Max Air Distance"
            value={`${latestNGS.max_air_distance?.toFixed(1)} yds` || "N/A"}
          />
          <StatCard
            label="Air Yards Differential"
            value={
              `${latestNGS.avg_air_yards_differential?.toFixed(1)} yds` || "N/A"
            }
          />
        </>
      )}

      {position === "RB" && latestNGS.stat_type === "rushing" && (
        <>
          <StatCard
            label="Avg Time to LOS"
            value={`${latestNGS.avg_time_to_los?.toFixed(2)}s` || "N/A"}
          />
          <StatCard
            label="Rush Yds Over Expected"
            value={
              `${latestNGS.rush_yards_over_expected?.toFixed(1)} yds` || "N/A"
            }
          />
          <StatCard
            label="Expected Rush Yards"
            value={`${latestNGS.expected_rush_yards?.toFixed(1)} yds` || "N/A"}
          />
          <StatCard
            label="Efficiency"
            value={`${(latestNGS.efficiency * 100)?.toFixed(1)}%` || "N/A"}
          />
          <StatCard
            label="Rush vs 8+ Defenders"
            value={
              `${(latestNGS.rush_pct_8_defenders * 100)?.toFixed(1)}%` || "N/A"
            }
          />
        </>
      )}

      {(position === "WR" || position === "TE") &&
        latestNGS.stat_type === "receiving" && (
          <>
            <StatCard
              label="Avg Separation"
              value={`${latestNGS.avg_separation?.toFixed(1)} yds` || "N/A"}
            />
            <StatCard
              label="Avg Cushion"
              value={`${latestNGS.avg_cushion?.toFixed(1)} yds` || "N/A"}
            />
            <StatCard
              label="Catch Percentage"
              value={`${latestNGS.catch_percentage?.toFixed(1)}%` || "N/A"}
            />
            <StatCard
              label="Target Share"
              value={
                `${(latestNGS.share_of_team_targets * 100)?.toFixed(1)}%` ||
                "N/A"
              }
            />
            <StatCard
              label="Avg YAC"
              value={`${latestNGS.avg_yac?.toFixed(1)} yds` || "N/A"}
            />
            <StatCard
              label="YAC Above Expected"
              value={
                `${latestNGS.avg_yac_above_expectation?.toFixed(1)} yds` ||
                "N/A"
              }
            />
          </>
        )}
    </div>
  );
}
