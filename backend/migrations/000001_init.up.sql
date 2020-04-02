CREATE TABLE "event" (
	"key" varchar(25) NOT NULL,
	end_date varchar(25) NULL,
	event_type int4 NULL,
	short_name varchar(250) NULL,
	"year" int4 NULL,
	CONSTRAINT event_pkey PRIMARY KEY (key)
);

CREATE TABLE team (
	"key" varchar(8) NOT NULL,
	team_number int4 NULL,
 	"name" varchar(1000) NULL,
 	CONSTRAINT team_pkey PRIMARY KEY (key)
);

CREATE TABLE "match" (
	"key" varchar(25) NOT NULL,
 	comp_level varchar(2) NULL,
 	set_number int4 NULL,
 	match_number int4 NULL,
 	winning_alliance varchar(5) NULL,
 	event_key varchar(25) NULL,
 	"time" int4 NULL,
 	actual_time int4 NULL,
 	predicted_time int4 NULL,
 	post_result_time int4 NULL,
 	score_breakdown json NULL,
 	CONSTRAINT match_pkey PRIMARY KEY (key),
 	CONSTRAINT match_event_key_fkey FOREIGN KEY (event_key) REFERENCES event(key)
);

CREATE TABLE alliance (
	"key" varchar(25) NOT NULL,
	score int4 NULL,
	color varchar(10) NULL,
	match_key varchar(25) NULL,
	CONSTRAINT alliance_pkey PRIMARY KEY (key),
	CONSTRAINT alliance_match_key_fkey FOREIGN KEY (match_key) REFERENCES match(key)
);


CREATE TABLE prediction_history (
	"match" varchar(25) NULL,
	model varchar(100) NULL,
	prediction json NULL,
	CONSTRAINT prediction_history_pkey PRIMARY KEY("match", model),
	CONSTRAINT prediction_history_match_fkey FOREIGN KEY (match) REFERENCES match(key)
);

CREATE TABLE forecast_history (
	model varchar(100) NULL,
	match_key varchar(25) NOT NULL,
	team_key varchar(8) NOT NULL,
	forecast float NULL,
	CONSTRAINT forecast_pkey PRIMARY KEY (model, match_key, team_key),
	CONSTRAINT forecast_match_fkey FOREIGN KEY (match_key) REFERENCES match(key)
);

CREATE TABLE alliance_teams (
	alliance_id varchar(25) NOT NULL,
	team_key varchar(8) NOT NULL,
	position int4 NOT NULL,
 	CONSTRAINT alliance_teams_pkey PRIMARY KEY (alliance_id, team_key),
	CONSTRAINT alliance_teams_alliance_id_fkey FOREIGN KEY (alliance_id) REFERENCES alliance(key),
	CONSTRAINT alliance_teams_team_key_fkey FOREIGN KEY (team_key) REFERENCES team(key)
);

CREATE TABLE json_cache (
	"key" varchar(100) PRIMARY KEY,
	value json NULL
);
