from cheesecake import create_app, db, tba
import click
from flask import Flask
from flask.cli import FlaskGroup
from cheesecake.models import Team, Event, Match, District, Alliance
import os

CURRENT_YEAR = 2018

app = create_app()

def get_teams():
    pages = 20
    for i in range(20):
        teams = tba.teams(i)
        if len(teams) == 0:
            break
        for team in teams:
            db.session.merge(Team(**team))
    db.session.commit()

def get_events(min_year, max_year):
    for year in range(min_year, max_year + 1):
        events = tba.events(year)
        for i, event in enumerate(events):
            if event["event_type"] > 10:
                continue
            db.session.merge(Event(**event))
            db.session.commit()
    db.session.commit()

def get_districts():
    for year in range(2003, CURRENT_YEAR + 1):
        for district in tba.districts(year):
            db.session.merge(District(**district))
    db.session.commit()

def get_matches():
    events = Event.query.all()
    existing_matches = set(x[0] for x in Match.query.with_entities(Match.key).all())
    for i, event in enumerate(events):
        matches = tba.event_matches(event.key)
        for match in matches:
            if int(match["time"]) > 1521997200:
                continue
            if match["key"] not in existing_matches:
                match["alliances"]["red"]["color"] = "red"
                match["alliances"]["blue"]["color"] = "blue"
                match["alliances"]["red"]["match_key"] = match["key"]
                match["alliances"]["blue"]["match_key"] = match["key"]
                match["alliances"]["red"]["key"] = match["key"] + "_red"
                match["alliances"]["blue"]["key"] = match["key"] + "_blue"
                red_teams = match["alliances"]["red"]["team_keys"]
                blue_teams = match["alliances"]["blue"]["team_keys"]
                del match["alliances"]["red"]["team_keys"]
                del match["alliances"]["blue"]["team_keys"]
                red = Alliance(**match["alliances"]["red"])
                blue = Alliance(**match["alliances"]["blue"])
                for team in red_teams:
                    t = Team.query.get(team)
                    if t:
                        red.team_keys.append(t)
                for team in blue_teams:
                    t = Team.query.get(team)
                    if t:
                        blue.team_keys.append(t)
                del match["alliances"]
                db.session.add(Match(**match))
                db.session.add(red)
                db.session.add(blue)
        print(event.key)
        db.session.commit()

def get_district_teams():
    districts = District.query.with_entities(District).all()
    for district in districts:
        dteams = tba.district_teams(district.key, keys=True)
        if type(dteams) != list:
            continue
        for key in dteams:
            team = Team.query.get(key)
            if team not in district.teams:
                district.teams.append(team)
    db.session.commit()

def get_event_teams():
    events = Event.query.filter(
        Event.event_type < 10
    ).all()
    for event in events:
        event_teams = tba.event_teams(event.key, keys=True)
        if type(event_teams) != list:
            continue
        for key in event_teams:
            team = Team.query.get(key)
            if team is None:
                continue
            if team not in event.teams:
                event.teams.append(team)
        print(event.key)
        db.session.commit()

@app.cli.command()
@click.argument('table')
def sync(table):
    if "FLASK_ENV" in os.environ and os.environ["FLASK_ENV"] == "development":
        min_year = 2016
        max_year = 2018
    else:
        min_year = 2003
        max_year = CURRENT_YEAR + 1
    if table == "teams" or table is None:
        print("Downloading Teams")
        get_teams()
    if table == "districts" or table is None:
        print("Downloading Districts")
        get_districts()
    if table == "events" or table is None:
        print("Downloading Events")
        get_events(min_year, max_year)
    if table == "matches" or table is None:
        print("Downloading Matches")
        get_matches()
    if table == "district_teams" or table is None:
        print("Downloading district teams")
        get_district_teams()
    if table == "event_teams" or table is None:
        print("Downloading event teams")
        get_event_teams()


if __name__ == '__main__':
    app.run()
#    socketio.init_app(app)
#    socketio.run(app, host='0.0.0.0')
