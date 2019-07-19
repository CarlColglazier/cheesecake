package main

type Match struct {
	Key             string `db:"key"`
	CompLevel       string `db:"comp_level"`
	SetNumber       int    `db:"set_number"`
	MatchNumber     int    `db:"match_number"`
	WinningAlliance string `db:"winning_alliance"`
	EventKey        string `db:"event_key"`
}

type Alliance struct {
	Key      string `db:"key"`
	Score    int    `db:"score"`
	Color    string `db:"color"`
	MatchKey string `db:"match_key"`
}

type AllianceTeam struct {
	Position   int    `db:"position"`
	AllianceId string `db:"alliance_id"`
	TeamKey    string `db:"team_key"`
}

type AllianceEntry struct {
	Alliance Alliance
	Teams    []string
}

type MatchEntry struct {
	Match     Match
	Alliances map[string]*AllianceEntry
}

func (config *Config) getMatches() (map[string]MatchEntry, error) {
	rows, err := config.Conn.Query(`SELECT * FROM match JOIN alliance on (match.key = alliance.match_key) JOIN alliance_teams on (alliance_teams.alliance_id = alliance.key)`)
	if err != nil {
		return nil, err
	}
	matches := make(map[string]MatchEntry)
	for rows.Next() {
		var match Match
		var alliance Alliance
		var aTeam AllianceTeam
		rows.Scan(
			&match.Key,
			&match.CompLevel,
			&match.SetNumber,
			&match.MatchNumber,
			&match.WinningAlliance,
			&match.EventKey,
			&alliance.Key,
			&alliance.Score,
			&alliance.Color,
			&alliance.MatchKey,
			&aTeam.Position,
			&aTeam.AllianceId,
			&aTeam.TeamKey,
		)
		if _, ok := matches[match.Key]; !ok {
			dict := make(map[string]*AllianceEntry)
			matches[match.Key] = MatchEntry{match, dict}
		}
		if _, ok := matches[match.Key].Alliances[alliance.Color]; !ok {
			list := make([]string, 0)
			matches[match.Key].Alliances[alliance.Color] = &AllianceEntry{alliance, list}
		}
		matches[match.Key].Alliances[alliance.Color].Teams = append(
			matches[match.Key].Alliances[alliance.Color].Teams,
			aTeam.TeamKey,
		)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return matches, nil
}
