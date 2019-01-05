from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload
from functools import lru_cache
import pickle

from ..models import *
from ..predictors import *

api = Blueprint('api', __name__)

MATCH_ORDER = {
    "qm": 0,
    "ef": 10,
    "qf": 11,
    "sf": 12,
    "f": 13
}
sort_order = db.case(value=Match.comp_level, whens=MATCH_ORDER)

@lru_cache()
def fetch_all_matches():
    return Match.query.join(Event).filter(
        Event.event_type < 10
    ).options(
        joinedload('alliances')
    ).order_by(
        Event.start_date,
        Match.time,
        sort_order,
        Match.match_number
    ).all()

@lru_cache()
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
            predictor.predict(match)
            predictor.add_result(match)
        filehandler = open("elo.pickle", 'wb')
        pickle.dump(predictor, filehandler)
        return predictor

@api.route('teams/<int:page>', methods=['GET'])
def get_teams(page=1):
    per_page = 250
    teams = Team.query.order_by(
        Team.team_number.desc()
    ).paginate(
        page,
        per_page,
        error_out=False)
    return jsonify([x.serialize for x in teams.items])

@api.route('events/<int:year>', methods=['GET'])
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
    return jsonify([x.as_dict() for x in events])

@api.route('matches/<string:event>', methods=['GET'])
def get_matches(event):
    matches = Match.query.filter(
        Match.event_key == event
    ).options(
        joinedload('alliances')
    ).order_by(
        Match.time,
        sort_order,
        Match.match_number
    ).all()
    series = [x.serialize for x in matches]
    predictor = run_elo()
    for s in series:
        s["prediction"] = predictor.prediction_history[s["key"]]
    return jsonify(series)
