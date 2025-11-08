'use client'

import { Brain, TrendingUp, Users, Zap } from 'lucide-react'
import Link from 'next/link'

export default function DashboardPage() {
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
        <StatCard
          title="AI Predictions Made"
          value="1,234"
          change="+12%"
          positive
        />
        <StatCard
          title="Accuracy Rate"
          value="78.5%"
          change="+5.2%"
          positive
        />
        <StatCard
          title="Fantasy Points Gained"
          value="342"
          change="+23%"
          positive
        />
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
  change,
  positive,
}: {
  title: string
  value: string
  change: string
  positive: boolean
}) {
  return (
    <div className="bg-white rounded-xl shadow-sm p-6">
      <p className="text-gray-600 text-sm mb-2">{title}</p>
      <p className="text-3xl font-bold text-gray-900 mb-2">{value}</p>
      <p className={`text-sm ${positive ? 'text-green-600' : 'text-red-600'}`}>
        {change} from last week
      </p>
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

