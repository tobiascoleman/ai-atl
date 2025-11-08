'use client'

import { useState } from 'react'
import { TrendingUp, AlertCircle } from 'lucide-react'
import apiClient from '@/lib/api/client'
import { TradeAnalysis } from '@/types/api'

export default function TradesPage() {
  const [teamAGives, setTeamAGives] = useState('')
  const [teamAGets, setTeamAGets] = useState('')
  const [teamBGives, setTeamBGives] = useState('')
  const [teamBGets, setTeamBGets] = useState('')
  const [analysis, setAnalysis] = useState<TradeAnalysis | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleAnalyze = async () => {
    setLoading(true)
    setError('')
    try {
      const { data } = await apiClient.post<TradeAnalysis>('/trades/analyze', {
        team_a_gives: teamAGives.split(',').map((p) => p.trim()),
        team_a_gets: teamAGets.split(',').map((p) => p.trim()),
        team_b_gives: teamBGives.split(',').map((p) => p.trim()),
        team_b_gets: teamBGets.split(',').map((p) => p.trim()),
      })
      setAnalysis(data)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to analyze trade')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Trade Analyzer</h1>
        <p className="text-gray-600 mt-2">
          Get AI-powered analysis of potential trades with fairness scoring
        </p>
      </div>

      {/* Trade Input */}
      <div className="bg-white rounded-xl shadow-sm p-6">
        <div className="grid md:grid-cols-2 gap-6">
          {/* Team A */}
          <div className="space-y-4">
            <h3 className="font-bold text-lg text-blue-600">Team A</h3>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Gives Away
              </label>
              <input
                type="text"
                value={teamAGives}
                onChange={(e) => setTeamAGives(e.target.value)}
                placeholder="Player IDs, comma separated"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 text-gray-900"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Receives
              </label>
              <input
                type="text"
                value={teamAGets}
                onChange={(e) => setTeamAGets(e.target.value)}
                placeholder="Player IDs, comma separated"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 text-gray-900"
              />
            </div>
          </div>

          {/* Team B */}
          <div className="space-y-4">
            <h3 className="font-bold text-lg text-green-600">Team B</h3>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Gives Away
              </label>
              <input
                type="text"
                value={teamBGives}
                onChange={(e) => setTeamBGives(e.target.value)}
                placeholder="Player IDs, comma separated"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 text-gray-900"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Receives
              </label>
              <input
                type="text"
                value={teamBGets}
                onChange={(e) => setTeamBGets(e.target.value)}
                placeholder="Player IDs, comma separated"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 text-gray-900"
              />
            </div>
          </div>
        </div>

        <button
          onClick={handleAnalyze}
          disabled={loading}
          className="mt-6 w-full py-3 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition flex items-center justify-center gap-2"
        >
          <TrendingUp size={18} />
          {loading ? 'Analyzing...' : 'Analyze Trade'}
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 flex items-center gap-3">
          <AlertCircle size={20} />
          {error}
        </div>
      )}

      {/* Analysis Results */}
      {analysis && (
        <div className="space-y-6">
          {/* Grades */}
          <div className="grid md:grid-cols-2 gap-6">
            <div className="bg-white rounded-xl shadow-sm p-6">
              <h3 className="text-lg font-bold mb-4 text-blue-600">Team A Grade</h3>
              <div className="text-5xl font-bold text-center text-blue-600">
                {analysis.team_a_grade}
              </div>
              <p className="text-center text-gray-600 mt-2">{analysis.team_a_value_change}</p>
            </div>

            <div className="bg-white rounded-xl shadow-sm p-6">
              <h3 className="text-lg font-bold mb-4 text-green-600">Team B Grade</h3>
              <div className="text-5xl font-bold text-center text-green-600">
                {analysis.team_b_grade}
              </div>
              <p className="text-center text-gray-600 mt-2">{analysis.team_b_value_change}</p>
            </div>
          </div>

          {/* Fairness Score */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-bold mb-4">Fairness Score</h3>
            <div className="flex items-center gap-4">
              <div className="flex-1 bg-gray-200 rounded-full h-4">
                <div
                  className="bg-gradient-to-r from-blue-500 to-green-500 h-4 rounded-full transition-all"
                  style={{ width: `${(analysis.fairness_score / 10) * 100}%` }}
                ></div>
              </div>
              <span className="text-2xl font-bold text-gray-900">
                {analysis.fairness_score}/10
              </span>
            </div>
          </div>

          {/* AI Analysis */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-bold mb-4">AI Analysis</h3>
            <p className="text-gray-700 whitespace-pre-wrap">{analysis.ai_analysis}</p>
          </div>
        </div>
      )}
    </div>
  )
}

