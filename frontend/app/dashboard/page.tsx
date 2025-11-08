'use client'

import { useEffect, useState } from 'react'
import { Brain, TrendingUp, Users, Zap } from 'lucide-react'
import Link from 'next/link'
import { statsAPI, DashboardStats } from '@/lib/api/stats'

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await statsAPI.getDashboardStats()
        setStats(data)
      } catch (error) {
        console.error('Failed to fetch dashboard stats:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-2">Your AI-powered fantasy command center</p>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <QuickActionCard
          href="/dashboard/insights"
          icon={<Brain className="w-8 h-8 text-blue-600" />}
          title="AI Game Script"
          description="Predict game flow"
          color="bg-blue-50"
        />
        <QuickActionCard
          href="/dashboard/chat"
          icon={<Zap className="w-8 h-8 text-purple-600" />}
          title="AI Chat"
          description="Get instant advice"
          color="bg-purple-50"
        />
        <QuickActionCard
          href="/dashboard/trades"
          icon={<TrendingUp className="w-8 h-8 text-green-600" />}
          title="Trade Analyzer"
          description="Evaluate trades"
          color="bg-green-50"
        />
        <QuickActionCard
          href="/dashboard/players"
          icon={<Users className="w-8 h-8 text-orange-600" />}
          title="Players"
          description="Browse all players"
          color="bg-orange-50"
        />
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {loading ? (
          <>
            <StatCardSkeleton />
            <StatCardSkeleton />
            <StatCardSkeleton />
          </>
        ) : stats ? (
          <>
            <StatCard
              title="Total Players"
              value={stats.total_players.toLocaleString()}
              subtitle={`${stats.current_season_year} Season`}
            />
            <StatCard
              title="Plays Analyzed"
              value={stats.total_plays.toLocaleString()}
              subtitle={`${stats.total_games.toLocaleString()} Games`}
            />
            <StatCard
              title="Next Gen Stats"
              value={stats.next_gen_stats.toLocaleString()}
              subtitle={`${stats.injured_players} Injured Players`}
            />
          </>
        ) : (
          <>
            <StatCard title="Total Players" value="0" subtitle="No data" />
            <StatCard title="Plays Analyzed" value="0" subtitle="No data" />
            <StatCard title="Next Gen Stats" value="0" subtitle="No data" />
          </>
        )}
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-xl shadow-sm p-6">
        <h2 className="text-xl font-bold mb-4">Quick Tips</h2>
        <div className="space-y-4">
          <TipItem
            title="Game Script Predictions"
            description="Check AI game flow predictions for this week's matchups"
          />
          <TipItem
            title="Waiver Wire Gems"
            description="AI has identified 3 high-value waiver pickups"
          />
          <TipItem
            title="Trade Opportunities"
            description="Analyze potential trades before your deadline"
          />
        </div>
      </div>
    </div>
  )
}

function QuickActionCard({
  href,
  icon,
  title,
  description,
  color,
}: {
  href: string
  icon: React.ReactNode
  title: string
  description: string
  color: string
}) {
  return (
    <Link
      href={href}
      className="block bg-white rounded-xl shadow-sm hover:shadow-md transition p-6"
    >
      <div className={`w-16 h-16 ${color} rounded-lg flex items-center justify-center mb-4`}>
        {icon}
      </div>
      <h3 className="font-bold text-lg mb-1">{title}</h3>
      <p className="text-gray-600 text-sm">{description}</p>
    </Link>
  )
}

function StatCard({
  title,
  value,
  subtitle,
}: {
  title: string
  value: string
  subtitle: string
}) {
  return (
    <div className="bg-white rounded-xl shadow-sm p-6">
      <p className="text-gray-600 text-sm mb-2">{title}</p>
      <p className="text-3xl font-bold text-gray-900 mb-2">{value}</p>
      <p className="text-sm text-gray-500">{subtitle}</p>
    </div>
  )
}

function StatCardSkeleton() {
  return (
    <div className="bg-white rounded-xl shadow-sm p-6 animate-pulse">
      <div className="h-4 bg-gray-200 rounded w-24 mb-2"></div>
      <div className="h-8 bg-gray-200 rounded w-32 mb-2"></div>
      <div className="h-4 bg-gray-200 rounded w-20"></div>
    </div>
  )
}

function TipItem({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex items-start gap-3 p-3 bg-blue-50 rounded-lg">
      <div className="w-2 h-2 bg-blue-600 rounded-full mt-2"></div>
      <div>
        <h4 className="font-semibold text-gray-900">{title}</h4>
        <p className="text-gray-600 text-sm">{description}</p>
      </div>
    </div>
  )
}

