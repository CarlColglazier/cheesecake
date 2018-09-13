from . import db, socketio, tba
from .models import Team, Event, Match, District, Alliance

CURRENT_YEAR = 2018

@socketio.on('teams')
def get_teams():
    pages = 20
    for i in range(20):
        teams = tba.teams(i)
        if len(teams) == 0:
            break
        for team in teams:
            db.session.merge(Team(**team))
        socketio.emit('teams', float(i) / pages)
        # Releases the thread.
        socketio.sleep(0)
    db.session.commit()
    socketio.emit('teams', 1)

@socketio.on('events')
def get_events():
    socketio.emit('events', 0.001)
    socketio.sleep(0)
    for year in range(2003, CURRENT_YEAR):
        events = tba.events(year)
        for i, event in enumerate(events):
            if event["event_type"] > 10:
                continue
            db.session.merge(Event(**event))
            db.session.commit()
            socketio.sleep(0)
        socketio.emit('events', (year - 2003) / (CURRENT_YEAR - 2003))
    socketio.emit('events', 1)
    db.session.commit()

@socketio.on('districts')
def get_districts():
    for district in tba.districts(CURRENT_YEAR):
        db.session.merge(District(**district))
    db.session.commit()

@socketio.on('matches')
def get_matches():
    events = Event.query.all()
    existing_matches = set(x[0] for x in Match.query.with_entities(Match.key).all())
    for i, event in enumerate(events):
        matches = tba.event_matches(event.key)
        for match in matches:
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
        socketio.emit('matches', float(i) / len(events))
        db.session.commit()
        socketio.sleep(0)
    socketio.emit('matches', 1)
    #db.session.commit()
