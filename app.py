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

@app.route('/api/espn/roster', methods=['GET'])
def get_my_roster():
    try:
        # Initialize the private League using credentials
        league = League(
            league_id=YOUR_LEAGUE_ID,
            year=YOUR_YEAR,
            espn_s2=YOUR_ESPN_S2,
            swid=YOUR_SWID
        )
        
        # Find the team that matches YOUR_TEAM_ID
        team = None
        for t in league.teams:
            if t.team_id == YOUR_TEAM_ID:
                team = t
                break
        
        # If team isn't found, return 404
        if not team:
            return jsonify({'error': f'Team with ID {YOUR_TEAM_ID} not found'}), 404
        
        # Create roster data list
        roster_data = []
        for player in team.roster:
            roster_data.append({
                'name': player.name,
                'position': player.position,
                'proTeam': player.proTeam,
                'lineupSlot': player.lineupSlot
            })
        
        return jsonify(roster_data)
    
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(port=5002, debug=True)
