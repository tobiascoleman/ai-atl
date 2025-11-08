'use client'

import { useState } from 'react'
import { Brain, TrendingUp, Activity, AlertCircle } from 'lucide-react'
import { insightsAPI } from '@/lib/api/insights'
import { GameScriptPrediction } from '@/types/api'

export default function InsightsPage() {
  const [gameId, setGameId] = useState('2024_09_KC_BUF')
  const [prediction, setPrediction] = useState<GameScriptPrediction | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handlePredict = async () => {
    setLoading(true)
    setError('')
    try {
      const data = await insightsAPI.getGameScript(gameId)
      setPrediction(data)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch prediction')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">AI Game Script Predictor</h1>
        <p className="text-gray-600 mt-2">
          Predict how games will unfold and which players will benefit
        </p>
      </div>

      {/* Input Section */}
      <div className="bg-white rounded-xl shadow-sm p-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Select Game
        </label>
        <div className="flex gap-3">
          <input
            type="text"
            value={gameId}
            onChange={(e) => setGameId(e.target.value)}
            placeholder="Enter game ID (e.g., 2024_09_KC_BUF)"
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
          />
          <button
            onClick={handlePredict}
            disabled={loading || !gameId}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition flex items-center gap-2"
          >
            <Brain size={18} />
            {loading ? 'Analyzing...' : 'Predict'}
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 flex items-center gap-3">
          <AlertCircle size={20} />
          {error}
        </div>
      )}

      {/* Prediction Results */}
      {prediction && (
        <div className="space-y-6">
          {/* Game Flow */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <div className="flex items-center gap-3 mb-4">
              <Activity className="text-blue-600" size={24} />
              <h2 className="text-xl font-bold">Predicted Game Flow</h2>
            </div>
            <p className="text-gray-700 whitespace-pre-wrap">{prediction.predicted_flow}</p>
          </div>

          {/* Confidence Score */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <h3 className="text-lg font-bold mb-3">Confidence Score</h3>
            <div className="flex items-center gap-4">
              <div className="flex-1 bg-gray-200 rounded-full h-4">
                <div
                  className="bg-green-500 h-4 rounded-full transition-all"
                  style={{ width: `${prediction.confidence_score * 100}%` }}
                ></div>
              </div>
              <span className="text-2xl font-bold text-gray-900">
                {Math.round(prediction.confidence_score * 100)}%
              </span>
            </div>
          </div>

          {/* Player Impacts */}
          <div className="bg-white rounded-xl shadow-sm p-6">
            <div className="flex items-center gap-3 mb-4">
              <TrendingUp className="text-green-600" size={24} />
              <h2 className="text-xl font-bold">Player Impact Predictions</h2>
            </div>
            <div className="space-y-3">
              {prediction.player_impacts && prediction.player_impacts.length > 0 ? (
                prediction.player_impacts.map((impact, i) => (
                  <div
                    key={i}
                    className="flex items-center justify-between p-4 bg-gradient-to-r from-blue-50 to-green-50 rounded-lg"
                  >
                    <div className="flex-1">
                      <h4 className="font-bold text-gray-900">{impact.player_name}</h4>
                      <p className="text-sm text-gray-600 mt-1">{impact.reasoning}</p>
                    </div>
                    <div className="ml-4 px-4 py-2 bg-green-100 text-green-700 font-bold rounded-lg">
                      {impact.impact}
                    </div>
                  </div>
                ))
              ) : (
                <p className="text-gray-500 text-center py-4">No player impacts predicted</p>
              )}
            </div>
          </div>

          {/* Key Factors */}
          {prediction.key_factors && prediction.key_factors.length > 0 && (
            <div className="bg-white rounded-xl shadow-sm p-6">
              <h3 className="text-lg font-bold mb-3">Key Factors</h3>
              <ul className="space-y-2">
                {prediction.key_factors.map((factor, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <div className="w-2 h-2 bg-blue-600 rounded-full mt-2"></div>
                    <span className="text-gray-700">{factor}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* Empty State */}
      {!prediction && !error && !loading && (
        <div className="bg-white rounded-xl shadow-sm p-12 text-center">
          <Brain className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-xl font-bold text-gray-900 mb-2">
            Ready to Predict Game Flow
          </h3>
          <p className="text-gray-600">
            Enter a game ID and click "Predict" to see AI-powered game script analysis
          </p>
        </div>
      )}
    </div>
  )
}

