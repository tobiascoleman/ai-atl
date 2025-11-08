from flask import Flask, jsonify, request
from espn_api.football import League
from flask_cors import CORS
import os

app = Flask(__name__)

# Configure CORS to allow requests from Go backend
CORS(app, resources={r"/api/*": {"origins": "http://localhost:8080"}})

# Default credentials (can be overridden via request headers or environment)
YOUR_LEAGUE_ID = int(os.getenv('ESPN_LEAGUE_ID', 929602296))
YOUR_TEAM_ID = int(os.getenv('ESPN_TEAM_ID', 10))
YOUR_YEAR = int(os.getenv('ESPN_YEAR', 2025))
YOUR_ESPN_S2 = os.getenv('ESPN_S2', 'AEANF5s/YFx8uRBzF0ySSDkyZkZVNuQ95avS3MuJaOMoWTdXFYiRItuIfiDSE/EADpCTJYbypKBuEva4kJ6+3kj/G58wrOwlk+HiORhAHPQeZ/ibNioe6PRhLjSLMttbmV2PKL6SjFT87LpLTYlgYL9Pw3cm32NNS8740CFpIbsUUBGLJ0Ry6dpXGL/dxMhX7AmhmdwQhfV7LsopKrI6tR/YD2NUCxTfs722KQHg0f64uSK3zdXAtNM8wNAkc7K1WsWCY1g35RHzE8esgza5WXwVcld3X7pAdGX6Wa1fn34OPA==')
YOUR_SWID = os.getenv('ESPN_SWID', '{06B8EDC1-CAAD-40F0-A6AB-22C15EDF791B}')

def get_league_and_team():
    """Helper function to initialize league and get team"""
    league = League(
        league_id=YOUR_LEAGUE_ID,
        year=YOUR_YEAR,
        espn_s2=YOUR_ESPN_S2,
        swid=YOUR_SWID
    )
    
    team = None
    for t in league.teams:
        if t.team_id == YOUR_TEAM_ID:
            team = t
            break
    
    if not team:
        return None, None, f'Team with ID {YOUR_TEAM_ID} not found'
    
    return league, team, None

@app.route('/api/espn/roster', methods=['GET'])
def get_my_roster():
    try:
        league, team, error = get_league_and_team()
        if error:
            return jsonify({'error': error}), 404
        
        # Get current week for projections
        current_week = league.current_week
        
        # Create roster data list with projected and actual points
        roster_data = []
        for player in team.roster:
            # Get projected points for current week
            projected = 0
            actual = 0
            try:
                if hasattr(player, 'stats') and current_week in player.stats:
                    projected = player.stats[current_week].get('projected_points', 0)
                    actual = player.stats[current_week].get('points', 0)
                # Fallback to season averages
                if projected == 0:
                    projected = getattr(player, 'projected_avg_points', 0)
                if actual == 0:
                    actual = getattr(player, 'avg_points', 0)
            except:
                projected = getattr(player, 'projected_avg_points', 0)
                actual = getattr(player, 'avg_points', 0)
            
            player_data = {
                'name': player.name,
                'position': player.position,
                'proTeam': player.proTeam,
                'lineupSlot': player.lineupSlot,
                'projectedPoints': projected,
                'points': actual,
                'injured': getattr(player, 'injured', False),
                'injuryStatus': getattr(player, 'injuryStatus', None),
            }
            roster_data.append(player_data)
        
        return jsonify(roster_data)
    
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/espn/optimize-lineup', methods=['GET'])
def optimize_lineup():
    try:
        league, team, error = get_league_and_team()
        if error:
            return jsonify({'error': error}), 404
        
        current_week = league.current_week
        
        # Get all players with their projections
        players = []
        for player in team.roster:
            projected = 0
            try:
                if hasattr(player, 'stats') and current_week in player.stats:
                    projected = player.stats[current_week].get('projected_points', 0)
                if projected == 0:
                    projected = getattr(player, 'projected_avg_points', 0)
            except:
                projected = getattr(player, 'projected_avg_points', 0)
            
            players.append({
                'name': player.name,
                'position': player.position,
                'proTeam': player.proTeam,
                'lineupSlot': player.lineupSlot,
                'eligibleSlots': player.eligibleSlots,
                'projectedPoints': projected,
                'injured': getattr(player, 'injured', False),
                'injuryStatus': getattr(player, 'injuryStatus', None),
                'playerId': getattr(player, 'playerId', None)
            })
        
        # Sort by projected points (highest first)
        players.sort(key=lambda x: x['projectedPoints'], reverse=True)
        
        # Define lineup requirements (typical ESPN lineup)
        lineup_slots = {
            'QB': 1,
            'RB': 2,
            'WR': 2,
            'TE': 1,
            'RB/WR/TE': 1,  # FLEX
            'D/ST': 1,
            'K': 1
        }
        
        optimal_lineup = []
        benched = []
        filled_slots = {slot: 0 for slot in lineup_slots.keys()}
        
        # First pass: Fill position-specific slots
        for player in players:
            if player['injured'] and player['injuryStatus'] in ['OUT', 'IR']:
                player['recommendedSlot'] = 'BE'
                benched.append(player)
                continue
            
            # Try to place in specific position slot
            if player['position'] in lineup_slots and filled_slots[player['position']] < lineup_slots[player['position']]:
                player['recommendedSlot'] = player['position']
                filled_slots[player['position']] += 1
                optimal_lineup.append(player)
            else:
                # Check if eligible for flex
                if 'RB/WR/TE' in player['eligibleSlots'] and filled_slots['RB/WR/TE'] < lineup_slots['RB/WR/TE']:
                    if player['position'] in ['RB', 'WR', 'TE']:
                        player['recommendedSlot'] = 'RB/WR/TE'
                        filled_slots['RB/WR/TE'] += 1
                        optimal_lineup.append(player)
                    else:
                        player['recommendedSlot'] = 'BE'
                        benched.append(player)
                else:
                    player['recommendedSlot'] = 'BE'
                    benched.append(player)
        
        return jsonify({
            'optimalLineup': optimal_lineup,
            'bench': benched,
            'totalProjected': sum(p['projectedPoints'] for p in optimal_lineup)
        })
    
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/espn/free-agents', methods=['GET'])
def get_free_agents():
    try:
        league, team, error = get_league_and_team()
        if error:
            return jsonify({'error': error}), 404
        
        # Get query parameters
        position = request.args.get('position', None)  # Filter by position (QB, RB, WR, TE, K, D/ST)
        size = int(request.args.get('size', 50))  # Number of results (default 50)
        
        # Handle empty string as None
        if position == '':
            position = None
        
        current_week = league.current_week
        
        # Get free agents from the league
        # ESPN API provides free_agents method
        free_agents = league.free_agents(size=size, position=position)
        
        # Process free agent data
        free_agent_data = []
        for player in free_agents:
            projected = 0
            actual = 0
            try:
                if hasattr(player, 'stats') and current_week in player.stats:
                    projected = player.stats[current_week].get('projected_points', 0)
                    actual = player.stats[current_week].get('points', 0)
                if projected == 0:
                    projected = getattr(player, 'projected_avg_points', 0)
                if actual == 0:
                    actual = getattr(player, 'avg_points', 0)
            except:
                projected = getattr(player, 'projected_avg_points', 0)
                actual = getattr(player, 'avg_points', 0)
            
            player_data = {
                'name': player.name,
                'position': player.position,
                'proTeam': player.proTeam,
                'projectedPoints': projected,
                'points': actual,
                'injured': getattr(player, 'injured', False),
                'injuryStatus': getattr(player, 'injuryStatus', 'ACTIVE'),
                'playerId': getattr(player, 'playerId', None),
                'percentOwned': getattr(player, 'percent_owned', 0),
                'percentStarted': getattr(player, 'percent_started', 0),
            }
            free_agent_data.append(player_data)
        
        return jsonify({
            'players': free_agent_data,
            'count': len(free_agent_data)
        })
    
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(port=5002, debug=True)
