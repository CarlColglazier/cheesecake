from flask import Blueprint, jsonify, request
from sqlalchemy.orm import joinedload
import json
import os
import datetime
import numpy as np

from ..models import *
from ..predictors import *
from ..utils import update_schedule, update_match
from .. import cache
from .queries import fetch_matches, sort_order
from .times import *

api = Blueprint('api', __name__)

elo_predictor = EloScorePredictor()

@cache.memoize(timeout=MINUTE)
def predict():
    global elo_predictor
    matches = fetch_matches(2019)
    filehandler = open("elo.json", 'r')
    elos = json.load(filehandler)
    predictor = EloScorePredictor()
    predictor.elos = elos
    for match in matches:
        p = predictor.predict(match)
        history = [x for x in match.predictions if x.model == type(predictor).__name__]
        if len(history) > 0:
            history = history[0]
        else:
            history = None
        if not history:
            history = PredictionHistory(match=match.key,
                                        prediction=p,
                                        model=type(predictor).__name__)
        history.prediction = p
        db.session.add(history)
        if sum([x.score for x in match.alliances]) != -2:
            predictor.add_result(match)
    db.session.commit()
    elo_predictor = predictor

@api.route('/', methods=['POST'])
def webhook():
    data = json.loads(request.data)
    if data["message_type"] == "schedule_updated":
        update_schedule(data["message_data"]["event_key"])
        predict()
    if data["message_type"] == "match_score":
        update_match(data["message_data"]["match"])
        predict()
    if data["message_type"] == "verification":
        print(data)
    return jsonify([])

@api.route('/', methods=['GET'])
def test():
    predict()
    return jsonify([])
    
@api.route('events/upcoming', methods=['GET'])
@cache.memoize(timeout=2 * HOUR)
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
    return jsonify([x.serialize for x in events])

@api.route('matches/<string:event>', methods=['GET'])
@cache.memoize(timeout=2 * MINUTE)
def get_matches(event):
    if Event.query.get(event) is None:
        resp = jsonify([])
        resp.status_code = 404
        return resp
    matches = Match.query.filter(
        Match.event_key == event
    ).options(
        joinedload(Match.alliances)
    ).options(
        joinedload(Match.predictions)
    ).order_by(
        Match.time,
        sort_order,
        Match.match_number
    ).all()
    series = [x.serialize for x in matches]
    return jsonify(series)

@api.route('verify/brier/<int:year>', methods=['GET'])
@cache.memoize(timeout=MINUTE)
def brier(year):
    matches = fetch_matches(year)
    completed = [x for x in matches if x.result() is not None]
    score = [(x.predictions[0].prediction - x.result()) ** 2 for x in completed]
    return jsonify({
        "brier": sum(score) / len(score)
    })

@api.route('verify/calibration/<int:year>', methods=['GET'])
def calibration(year):
    matches = fetch_matches(year)
    completed = [x for x in matches if x.result() is not None]
    m = []
    for match in completed:
        m.append({
            "winner": match.result(),
            "prediction": match.predictions[0].prediction
        })
    results = {}
    for i in np.arange(0, 1, .1):
        correct = 0
        total = 0
        for match in m:
            if match["prediction"] >= i and match["prediction"] < i + .1:
                total += 1
                if match["winner"] == 1:
                    correct += 1
        results[i] = {
            "correct": correct,
            "total": total,
            "fraction": float(correct) / total
        }
    return jsonify(results)
        
@api.route('/teams/rankings', methods=['GET'])
def rankings():
    if not elo_predictor:
        return jsonify([])
    return jsonify(
        sorted(
            elo_predictor.elos.items(),
            key=lambda x: x[1],
            reverse=True
        )[0:25]
    )
