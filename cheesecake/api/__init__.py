from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload
import pickle
import numpy as np
import pandas as pd
import os
import datetime

from ..models import *
from ..predictors import *
from ..states import EventState
from ..simulation import *
from .. import cache

MINUTE = 60
HOUR = 3600
DAY = 86400

api = Blueprint('api', __name__)

MATCH_ORDER = {
    "qm": 0,
    "ef": 10,
    "qf": 11,
    "sf": 12,
    "f": 13
}
sort_order = db.case(value=Match.comp_level, whens=MATCH_ORDER)

@cache.memoize(timeout=DAY)
def fetch_all_matches():
    matches =  Match.query.join(Event).filter(
        Event.event_type < 10
    ).options(
        joinedload('alliances')
    ).order_by(
        Event.start_date,
        Match.time,
        sort_order,
        Match.match_number
    ).all()
    return matches

@cache.memoize(timeout=DAY)
def run_elo():
    # This is kind of a hack, but I really don't want to keep
    # having to run this over and over again on each refresh,
    # so I'm going to just load it from a file.
    try:
        filehandler = open("elo.pickle", 'rb')
        predictor = pickle.load(filehandler)
        return predictor
    except:
        matches = fetch_all_matches()
        predictor = EloScorePredictor()
        for match in matches:
            p = predictor.predict(match)
            history = PredictionHistory(match=match.key,
                                    prediction=p,
                                    model=type(predictor).__name__)
            db.session.merge(history)
            predictor.add_result(match)
        db.session.commit()
        filehandler = open("elo.pickle", 'wb')
        pickle.dump(predictor, filehandler)
        return predictor

@api.route('teams/<int:page>', methods=['GET'])
@cache.memoize(timeout=HOUR)
def get_teams(page=1):
    per_page = 250
    teams = Team.query.order_by(
        Team.team_number.desc()
    ).paginate(
        page,
        per_page,
        error_out=False)
    return jsonify([x.serialize for x in teams.items])

@api.route('events/upcoming', methods=['GET'])
def get_official_events_upcoming():
    d = datetime.date.today()
    t = datetime.date.today()
    while d.weekday() != 6:
        d += datetime.timedelta(1)
    events = Event.query.filter(
        Event.first_event_code != None
    ).filter(
        Event.event_type < 99
    ).filter(
        Event.end_date >= str(t)
    ).filter(
        Event.end_date <= str(d)
    ).order_by(
        Event.start_date,
        Event.name
    ).all()
    print(str(t), str(d))
    return jsonify([x.serialize for x in events])


@api.route('events/<int:year>', methods=['GET'])
@cache.memoize(timeout=DAY)
def get_official_events_year(year):
    events = Event.query.filter(
        Event.first_event_code != None
    ).filter(
        Event.year == year
    ).filter(
        Event.event_type < 99
    ).order_by(
        Event.start_date,
        Event.name
    ).all()
    return jsonify([x.serialize for x in events])

@api.route('matches/<string:event>', methods=['GET'])
def get_matches(event):
    if Event.query.get(event) is None:
        resp = jsonify([])
        resp.status_code = 404
        return resp
    matches = Match.query.filter(
        Match.event_key == event
    ).options(
        joinedload('alliances')
    ).options(
        joinedload('predictions')
    ).order_by(
        Match.time,
        sort_order,
        Match.match_number
    ).all()
    series = [x.serialize for x in matches]
    return jsonify(series)

def simulate_event(event):
    event = Event.query.get(event)
    state = event.state()
    predictor = run_elo()
    if state == EventState.NO_SCHEDULE:
        simulator = PreEventSimulator(event, predictor)
    else:
        simulator = QualificationEventSimulator(event, predictor)
    return simulator

@api.route('simulate/<string:event>/matches', methods=['GET'])
@cache.memoize(timeout=MINUTE)
def simulate_event_endpoint(event):
    simulator = simulate_event(event)
    predictions = simulator.matches()
    return jsonify([x[0].serialize for x in predictions])

@api.route('event/<string:key>', methods=['GET'])
@cache.memoize(timeout=MINUTE)
def event(key):
    event_data = Event.query.get(key)
    simulator = simulate_event(key)
    predictions = simulator.matches()
    return jsonify({
        "event": event_data.as_dict(),
        "simulate": [x[0].serialize for x in predictions]
    })
