export type EVType = {
    team: string;
    points: number[];
    bcount: number[];
}

export type MatchDataType = { [key: string]: any };

export type TeamSimType1 = {
    [mcount: string]: number
}

export type TeamSimsType = {
    [team: string]: {
        [mtype: string]: TeamSimType1
    };
}

export type EventDataType = {
    ev: EVType[];
    matches: MatchDataType[];
    team_sims: TeamSimsType;
    predictions?: {
        [key:string]:{[pkey:string]: number}
    }
    schedule: Schedule2023[];
}

export type Schedule2023 = {
    key: string;
    match_number: number;
    time: number;
    red_teams: number[];
    blue_teams: number[];
    red_score: number;
    blue_score: number;
}

export type TeamSummaryType = {
    team: string;
    median: number;
    qlow: number;
    qhigh: number;
}
