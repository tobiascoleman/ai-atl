'use client'

import { useState, useEffect } from 'react'
import { playersAPI } from '@/lib/api/players'
import { Player } from '@/types/api'
import { Search, Users } from 'lucide-react'
import PlayerDetailModal from '@/components/PlayerDetailModal'

export default function PlayersPage() {
  const [players, setPlayers] = useState<Player[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [position, setPosition] = useState('')
  const [team, setTeam] = useState('')
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)

  // Debounce search to avoid too many API calls
  useEffect(() => {
    const timer = setTimeout(() => {
      loadPlayers()
    }, 300) // Wait 300ms after user stops typing

    return () => clearTimeout(timer)
  }, [search, position, team])

  const loadPlayers = async () => {
    setLoading(true)
    try {
      const data = await playersAPI.getPlayers({
        search: search || undefined,
        position: position || undefined,
        team: team || undefined,
        limit: 100,
        sort: 'name',
        order: 'asc',
      })
      setPlayers(data.players || [])
    } catch (error) {
      console.error('Failed to load players:', error)
    } finally {
      setLoading(false)
    }
  }

  const handlePlayerClick = (player: Player) => {
    console.log('👆 Player clicked:', player)
    console.log('📋 Player nfl_id:', player.nfl_id)
    
    if (!player.nfl_id) {
      console.error('❌ Player has no nfl_id!', player)
      alert(`Cannot view details: Player ${player.name} has no NFL ID`)
      return
    }
    
    setSelectedPlayer(player)
    setIsModalOpen(true)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Players</h1>
        <p className="text-gray-600 mt-2">
          {loading ? (
            'Loading players...'
          ) : (
            `Showing ${players.length} player${players.length !== 1 ? 's' : ''}`
          )}
        </p>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm p-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              <Search size={16} className="inline mr-2" />
              Search
            </label>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search players..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 text-gray-900"
            />
          </div>

          <div>
            <label htmlFor="position-filter" className="block text-sm font-medium text-gray-700 mb-2">Position</label>
            <select
              id="position-filter"
              value={position}
              onChange={(e) => setPosition(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 text-gray-900"
            >
              <option value="">All Positions</option>
              <option value="QB">QB</option>
              <option value="RB">RB</option>
              <option value="WR">WR</option>
              <option value="TE">TE</option>
              <option value="K">K</option>
              <option value="P">P</option>
              <option value="DL">DL</option>
              <option value="LB">LB</option>
              <option value="DB">DB</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Team</label>
            <input
              type="text"
              value={team}
              onChange={(e) => setTeam(e.target.value.toUpperCase())}
              placeholder="e.g., KC"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 text-gray-900"
            />
          </div>
        </div>
      </div>

      {/* Players List */}
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-12 text-center text-gray-500">Loading players...</div>
        ) : players.length === 0 ? (
          <div className="p-12 text-center text-gray-500">No players found</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Player
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Position
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Team
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Season Stats
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {players.map((player) => {
                  // Determine what stats to show based on position
                  const defensivePositions = ['LB', 'DE', 'DT', 'CB', 'S', 'ILB', 'OLB', 'MLB', 'NT', 'SS', 'FS', 'DB', 'DL']
                  const isDefensive = defensivePositions.includes(player.position)
                  
                  let statsDisplay = 'No stats'
                  
                  if (isDefensive && (player.tackles || player.sacks)) {
                    // Defensive stats
                    const parts = []
                    if (player.tackles) parts.push(`${player.tackles} tackles`)
                    if (player.sacks) parts.push(`${player.sacks} sacks`)
                    if (player.tackles_for_loss) parts.push(`${player.tackles_for_loss} TFL`)
                    if (player.def_interceptions) parts.push(`${player.def_interceptions} INTs`)
                    if (player.pass_defended) parts.push(`${player.pass_defended} PD`)
                    statsDisplay = parts.join(', ')
                  } else if (player.position === 'QB' && player.passing_yards) {
                    statsDisplay = `${player.passing_yards.toLocaleString()} pass yds, ${player.passing_tds ?? 0} pass TDs`
                  } else if (player.position === 'RB' && (player.rushing_yards || player.receiving_yards)) {
                    // RBs: Show rushing and receiving stats separately with clear labels
                    const parts = []
                    if (player.rushing_yards) {
                      parts.push(`${player.rushing_yards.toLocaleString()} rush yds, ${player.rushing_tds ?? 0} rush TDs`)
                    }
                    if (player.receiving_yards) {
                      parts.push(`${player.receptions ?? 0} rec, ${player.receiving_yards} rec yds, ${player.receiving_tds ?? 0} rec TDs`)
                    }
                    statsDisplay = parts.join(' | ')
                  } else if ((player.position === 'WR' || player.position === 'TE') && player.receiving_yards) {
                    statsDisplay = `${player.receptions ?? 0} rec, ${player.receiving_yards.toLocaleString()} yds, ${player.receiving_tds ?? 0} TDs`
                  }

                  return (
                    <tr
                      key={player.id}
                      className="hover:bg-blue-50 cursor-pointer transition-colors"
                      onClick={() => handlePlayerClick(player)}
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="font-medium text-gray-900">{player.name}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                          {player.position}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {player.team}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                        {statsDisplay}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {(() => {
                          const status = player.status_description || 'Active'
                          const isRetired = status.includes('Retired')
                          const isInjured = player.status === 'INA' || 
                                          status.includes('Injured') || 
                                          status.includes('PUP') ||
                                          status.includes('Non-Football Injury')
                          const isPracticeSquad = status.includes('Practice Squad')
                          const isWaived = status.includes('Waived')
                          const isSuspended = status.includes('Suspended')
                          
                          let colorClass = 'bg-green-100 text-green-800' // Active
                          if (isRetired) colorClass = 'bg-gray-100 text-gray-700'
                          else if (isInjured) colorClass = 'bg-red-100 text-red-800'
                          else if (isWaived) colorClass = 'bg-orange-100 text-orange-800'
                          else if (isSuspended) colorClass = 'bg-purple-100 text-purple-800'
                          else if (isPracticeSquad) colorClass = 'bg-blue-100 text-blue-800'
                          
                          return (
                            <span className={`px-2 py-1 text-xs font-semibold rounded-full ${colorClass}`}>
                              {status}
                            </span>
                          )
                        })()}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Player Detail Modal */}
      {selectedPlayer && (
        <PlayerDetailModal
          player={selectedPlayer}
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
        />
      )}
    </div>
  )
}

