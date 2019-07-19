package tba

// Starting point for these models was taken from
// https://github.com/frc1418/tbago/blob/master/models.go

type Team struct {
	Key              string         `json:"key"`
	TeamNumber       int            `json:"team_number"`
	Nickname         string         `json:"nickname"`
	Name             string         `json:"name"`
	City             string         `json:"city"`
	StateProv        string         `json:"state_prov"`
	Country          string         `json:"country"`
	Address          string         `json:"address"`
	PostalCode       string         `json:"postal_code"`
	GMapsPlaceID     string         `json:"gmaps_place_id"`
	GMapsURL         string         `json:"gmaps_url"`
	Latitude         float64        `json:"lat"`
	Longitude        float64        `json:"lng"`
	LocationName     string         `json:"location_name"`
	Website          string         `json:"website"`
	Motto            string         `json:"motto"`
	RookieYear       int            `json:"rookie_year"`
	HomeChampionship map[int]string `json:"home_championship"`
}

type Event struct {
	Key             string   `json:"key"`
	Name            string   `json:"name"`
	EventCode       string   `json:"event_code"`
	EventType       int      `json:"event_type"`
	District        District `json:"district"`
	City            string   `json:"city"`
	StateProv       string   `json:"state_prov"`
	PostalCode      string   `json:"postal_code"`
	Country         string   `json:"country"`
	Address         string   `json:"address"`
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	Year            int      `json:"year"`
	ShortName       string   `json:"short_name"`
	EventTypeString string   `json:"event_type_string"`
	Week            int      `json:"week"`
	GMapsPlaceID    string   `json:"gmaps_place_id"`
	GMapsURL        string   `json:"gmaps_url"`
	Latitude        float64  `json:"lat"`
	Longitude       float64  `json:"lng"`
	LocationName    string   `json:"location_name"`
	Timezone        string   `json:"timezone"`
	DivisionKeys    []string `json:"division_keys"`
	Website         string   `json:"website"`
	FIRSTEventID    string   `json:"first_event_id"`
	FIRSTEventCode  string   `json:"first_event_code"`
	Webcasts        []struct {
		Channel string `json:"channel"`
		Type    string `json:"type"`
	} `json:"webcasts"`
	ParentEventKey    string `json:"parent_event_key"`
	PlayoffType       int    `json:"playoff_type"`
	PlayoffTypeString string `json:"playoff_type_string"`
}

type Award struct {
	Name          string `json:"name"`
	AwardType     int    `json:"award_type"`
	EventKey      string `json:"event_key"`
	RecipientList []struct {
		TeamKey string `json:"team_key"`
		Awardee string `json:"awardee"`
	} `json:"recipient_list"`
	Year int `json:"year"`
}

type Match struct {
	Key         string `json:"key"`
	CompLevel   string `json:"comp_level"`
	SetNumber   int    `json:"set_number"`
	MatchNumber int    `json:"match_number"`
	Alliances   struct {
		Blue struct {
			DQTeams        []string `json:"dq_team_keys"`
			Score          int      `json:"score"`
			SurrogateTeams []string `json:"surrogate_team_keys"`
			TeamKeys       []string `json:"team_keys"`
		} `json:"blue"`
		Red struct {
			DQTeams        []string `json:"dq_team_keys"`
			Score          int      `json:"score"`
			SurrogateTeams []string `json:"surrogate_team_keys"`
			TeamKeys       []string `json:"team_keys"`
		} `json:"red"`
	} `json:"alliances"`
	WinningAlliance string `json:"winning_alliance"`
	EventKey        string `json:"event_key"`
	Time            int64  `json:"time"`
	ActualTime      int64  `json:"actual_time"`
	PredictedTime   int64  `json:"predicted_time"`
	PostResultTime  int64  `json:"post_result_time"`
	ScoreBreakdown  struct {
		Red  interface{} `json:"red"`
		Blue interface{} `json:"blue"`
	} `json:"score_breakdown"`
	Videos []struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	} `json:"videos"`
}

type District struct {
	Abbreviation string `json:"abbreviation"`
	DisplayName  string `json:"display_name"`
	Key          string `json:"key"`
	Year         int    `json:"year"`
}
