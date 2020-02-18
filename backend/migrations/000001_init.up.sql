CREATE TABLE "event" (
	"key" varchar(25) NOT NULL,
--	address varchar(1000) NULL,
--	city varchar(100) NULL,
--	country varchar(50) NULL,
	end_date varchar(25) NULL,
--	event_code varchar(10) NULL,
	event_type int4 NULL,
--	event_type_string varchar(50) NULL,
--	first_event_code varchar(25) NULL,
--	first_event_id varchar(50) NULL,
--	gmaps_place_id varchar(100) NULL,
--	gmaps_url varchar(250) NULL,
--	lat float8 NULL,
--	lng float8 NULL,
--	location_name varchar(100) NULL,
--	"name" varchar(250) NULL,
--	playoff_type int4 NULL,
--	playoff_type_string varchar(100) NULL,
--	postal_code varchar(50) NULL,
	short_name varchar(250) NULL,
--	start_date varchar(10) NULL,
--	state_prov varchar(50) NULL,
--	timezone varchar(50) NULL,
--	website varchar(100) NULL,
--	week int4 NULL,
	"year" int4 NULL,
	CONSTRAINT event_pkey PRIMARY KEY (key)
);

CREATE TABLE team (
	"key" varchar(8) NOT NULL,
	team_number int4 NULL,
-- 	nickname varchar(100) NULL,
 	"name" varchar(1000) NULL,
-- 	city varchar(100) NULL,
-- 	state_prov varchar(100) NULL,
-- 	country varchar(100) NULL,
-- 	address varchar(1000) NULL,
-- 	postal_code varchar(25) NULL,
-- 	website varchar(250) NULL,
-- 	rookie_year int4 NULL,
-- 	motto varchar(250) NULL,
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
	--	ptime integer NULL,
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
 	CONSTRAINT alliance_teams_pkey PRIMARY KEY (alliance_id, team_key),
	CONSTRAINT alliance_teams_alliance_id_fkey FOREIGN KEY (alliance_id) REFERENCES alliance(key),
	CONSTRAINT alliance_teams_team_key_fkey FOREIGN KEY (team_key) REFERENCES team(key)
);

CREATE TABLE json_cache (
       "key" varchar(100) PRIMARY KEY,
       value json NULL
);
