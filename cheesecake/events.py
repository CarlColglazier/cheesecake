from . import db, socketio, tba
from .models import Team, Event, Match, District

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
    events = tba.events(CURRENT_YEAR)
    for i, event in enumerate(events):
        db.session.merge(Event(**event))
        socketio.emit('events', float(i) / len(events))
        socketio.sleep(0)
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
    for i, event in enumerate(events):
        matches = tba.event_matches(event.key)
        for match in matches:
            db.session.merge(Match(**match))
        print(event.key)
        socketio.emit('matches', float(i) / len(events))
        socketio.sleep(0)
    socketio.emit('matches', 1)
    db.session.commit()
