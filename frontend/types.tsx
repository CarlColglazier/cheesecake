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
}

export type TeamSummaryType = {
    team: string;
    median: number;
    qlow: number;
    qhigh: number;
}
