package espn

// League represents the top-level ESPN league structure
type League struct {
	ID     int64  `json:"id"`
	Teams  []Team `json:"teams"`
	Status struct {
		IsViewable bool `json:"isViewable"`
	} `json:"status"`
	Settings struct {
		Name string `json:"name"`
	} `json:"settings"`
}

// Team represents an ESPN fantasy team
type Team struct {
	ID           int64  `json:"id"`
	Abbrev       string `json:"abbrev"`
	Location     string `json:"location"`
	Nickname     string `json:"nickname"`
	PrimaryOwner string `json:"primaryOwner"`
	Logo         string `json:"logo"`
	Roster       struct {
		Entries []RosterEntry `json:"entries"`
	} `json:"roster"`
}

// RosterEntry represents a player on a team's roster
type RosterEntry struct {
	PlayerID       int64  `json:"playerId"`
	LineupSlotID   int64  `json:"lineupSlotId"`
	InjuryStatus   string `json:"injuryStatus"`
	PlayerPoolEntry struct {
		Player PlayerData `json:"player"`
	} `json:"playerPoolEntry"`
}

// PlayerData contains detailed player information
type PlayerData struct {
	ID                int64   `json:"id"`
	FirstName         string  `json:"firstName"`
	LastName          string  `json:"lastName"`
	FullName          string  `json:"fullName"`
	DefaultPositionID int64   `json:"defaultPositionId"`
	EligibleSlots     []int64 `json:"eligibleSlots"`
	ProTeamID         int64   `json:"proTeamId"`
	Active            bool    `json:"active"`
	Injured           bool    `json:"injured"`
}

// TeamRoster is the simplified roster structure for API responses
type TeamRoster struct {
	TeamID       int64    `json:"team_id"`
	TeamName     string   `json:"team_name"`
	Abbreviation string   `json:"abbreviation"`
	Owner        string   `json:"owner"`
	Players      []Player `json:"players"`
}

// Player represents a fantasy football player
type Player struct {
	ID                int64   `json:"id"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	FullName          string  `json:"full_name"`
	DefaultPositionID int64   `json:"default_position_id"`
	Position          string  `json:"position"`
	EligibleSlots     []int64 `json:"eligible_slots"`
	ProTeamID         int64   `json:"pro_team_id"`
	Active            bool    `json:"active"`
	Injured           bool    `json:"injured"`
	InjuryStatus      string  `json:"injury_status"`
	LineupSlotID      int64   `json:"lineup_slot_id"`
	LineupSlotName    string  `json:"lineup_slot_name"`
}

// TeamInfo contains basic team information
type TeamInfo struct {
	ID           int64  `json:"id"`
	Abbreviation string `json:"abbreviation"`
	Location     string `json:"location"`
	Nickname     string `json:"nickname"`
	Owner        string `json:"owner"`
	LogoURL      string `json:"logo_url"`
}

// Position ID to name mapping
var PositionMap = map[int64]string{
	0:  "QB",
	1:  "TQB", // Two-QB
	2:  "RB",
	3:  "RB/WR",
	4:  "WR",
	5:  "WR/TE",
	6:  "TE",
	7:  "OP",   // Offensive Player
	8:  "DT",
	9:  "DE",
	10: "LB",
	11: "DL",
	12: "CB",
	13: "S",
	14: "DB",
	15: "DP",   // Defensive Player
	16: "D/ST", // Defense/Special Teams
	17: "K",
	20: "Bench",
	21: "IR",    // Injured Reserve
	23: "FLEX",
	24: "ER",    // Emergency
}

// GetPositionName returns the position name for a given position ID
func GetPositionName(positionID int64) string {
	if name, ok := PositionMap[positionID]; ok {
		return name
	}
	return "UNKNOWN"
}
