from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload
import pickle
import numpy as np
import pandas as pd

from ..models import *
from ..predictors import *
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

@cache.cached(timeout=MINUTE, key_prefix='all_matches')
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

@cache.cached(timeout=MINUTE, key_prefix='run_elo')
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
@cache.cached(timeout=HOUR)
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
@cache.cached(timeout=DAY)
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

@api.route('simulate/<string:event>', methods=['GET'])
@cache.cached(timeout=MINUTE)
def simulate_event(event):
    teams = [x.key for x in Event.query.get(event).teams]
    predictor = run_elo()
    np.random.seed(0)
    sample = np.random.choice(teams, size=(10000, 6))
    predictions = [predictor.predict_keys(x) for x in sample]
    reds = sample[:,0:3].flatten()
    blues = sample[:,3:6].flatten()
    pred_repeat = np.repeat(predictions, 3)
    df = pd.DataFrame({
        "teams": np.concatenate((reds, blues), axis=None),
        "predictions": np.concatenate((pred_repeat, np.subtract(1.0, pred_repeat)))
    })
    dic = df.groupby("teams").mean().sort_values(by='predictions', ascending=False)
    ## TODO: Surely there is a better way to do this.
    values = []
    for key, val in dic["predictions"].iteritems():
        values.append({
            'key': key,
            'mean': val
        })
    return jsonify(values)
