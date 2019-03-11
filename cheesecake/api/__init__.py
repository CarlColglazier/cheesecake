from flask import Blueprint, jsonify, request
from sqlalchemy.orm import joinedload
import json
import os
import datetime
import numpy as np
from itertools import chain

from ..models import *
from ..predictors import *
from ..utils import update_schedule, update_match
from .. import cache
from .queries import fetch_matches, sort_order
from .times import *

api = Blueprint('api', __name__)

elo_predictor = EloScorePredictor()
hab_predictor = BetaPredictor(0.7229, 2.4517, "habDockingRankingPoint")
rocket_predictor = BetaPredictor(0.5, 12.0, "completeRocketRankingPoint")

@cache.memoize(timeout=MINUTE)
def predict():
    global elo_predictor, hab_predictor, rocket_predictor
    matches = chain.from_iterable(fetch_matches(2019))
    filehandler = open("elo.json", 'r')
    elos = json.load(filehandler)
    predictor = EloScorePredictor()
    h_predictor = BetaPredictor(0.7229, 2.4517, "habDockingRankingPoint")
    r_predictor = BetaPredictor(0.5, 12.0, "completeRocketRankingPoint")
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
        p = h_predictor.predict(match)
        history = [x for x in match.predictions if h_predictor.feature in x.model]
        if len(history) > 0:
            for h in history:
                h.prediction = p[h.model]
                db.session.add(h)
        else:
            for alliance in match.alliances:
                key = h_predictor.feature + alliance.color
                h = PredictionHistory(match=match.key,
                                      prediction=p[key],
                                      model=key)
                db.session.add(h)
        p = r_predictor.predict(match)
        history = [x for x in match.predictions if r_predictor.feature in x.model]
        if len(history) > 0:
            for h in history:
                h.prediction = p[h.model]
                db.session.add(h)
        else:
            for alliance in match.alliances:
                key = r_predictor.feature + alliance.color
                h = PredictionHistory(match=match.key,
                                      prediction=p[key],
                                      model=key)
                db.session.add(h)
        if sum([x.score for x in match.alliances]) != -2:
            predictor.add_result(match)
            h_predictor.add_result(match)
            r_predictor.add_result(match)
    db.session.commit()
    elo_predictor = predictor
    hab_predictor = h_predictor

@api.route('/', methods=['POST'])
def webhook():
    data = json.loads(request.data)
    if data["message_type"] == "schedule_updated":
        update_schedule(data["message_data"]["event_key"])
        predict()
    if data["message_type"] == "match_score":
        m = update_match(data["message_data"]["match"])
        global elo_predictor, hab_predictor, rocket_predictor
        elo_predictor.add_result(m)
        hab_predictor.add_result(m)
        rocket_predictor.add_result(m)
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
@cache.memoize(timeout=MINUTE)
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

@api.route('rankings/<string:event>', methods=['GET'])
@cache.memoize(timeout=MINUTE)
def get_rankings(event):
    if Event.query.get(event) is None:
        resp = jsonify([])
        resp.status_code = 404
        return resp
    matches = Match.query.filter(
        Match.event_key == event
    ).filter(
        Match.comp_level == "qm"
    ).options(
        joinedload(Match.alliances)
    ).order_by(
        Match.time,
        sort_order,
        Match.match_number
    ).all()
    team_scores = {}
    for match in matches:
        for alliance in match.alliances:
            for team in alliance.team_keys:
                if team.key not in team_scores:
                    team_scores[team.key] = 0
    for match in filter(lambda x: x.result() is not None, matches):
        winner = match.winning_alliance
        points = 2
        alliances = match.get_alliances()
        if winner == "":
            points = 1
            teams = list(chain.from_iterable([x.team_keys for x in match.alliances]))
        else:
            teams = alliances[winner].team_keys
        for team in teams:
            team_scores[team.key] += points
        colors = ['red', 'blue']
        for color in colors:
            # TODO
            if not match.score_breakdown:
                continue
            if match.score_breakdown[color]["habDockingRankingPoint"]:
                for team in alliances[color].team_keys:
                    team_scores[team.key] += 1
            if match.score_breakdown[color]["completeRocketRankingPoint"]:
                for team in alliances[color].team_keys:
                    team_scores[team.key] += 1
    for match in filter(lambda x: x.result() is None, matches):
        alliances = match.get_alliances()
        p = match.get_prediction("EloScorePredictor").prediction
        hr = match.get_prediction("habDockingRankingPointred").prediction
        hb = match.get_prediction("habDockingRankingPointblue").prediction
        rr = match.get_prediction("completeRocketRankingPointred").prediction
        rb = match.get_prediction("completeRocketRankingPointblue").prediction
        for team in alliances["red"].team_keys:
            team_scores[team.key] += 2 * p
            team_scores[team.key] += hr
            team_scores[team.key] += rr
        for team in alliances["blue"].team_keys:
            team_scores[team.key] += 2 * (1 - p)
            team_scores[team.key] += hb
            team_scores[team.key] += rb
    #for match in matches:
    return jsonify(team_scores)

@api.route('verify/brier/<int:year>', methods=['GET'])
@cache.memoize(timeout=MINUTE)
def brier(year):
    matches = chain.from_iterable(fetch_matches(2019))
    completed = [x for x in matches if x.result() is not None]
    score = [(x.get_prediction("EloScorePredictor").prediction - x.result()) ** 2 for x in completed]
    return jsonify({
        "brier": sum(score) / len(score)
    })

@api.route('verify/calibration/<int:year>', methods=['GET'])
def calibration(year):
    matches = chain.from_iterable(fetch_matches(2019))
    completed = [x for x in matches if x.result() is not None]
    m = []
    for match in completed:
        m.append({
            "winner": match.result(),
            "prediction": match.get_prediction("EloScorePredictor").prediction
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
    global elo_predcitor
    return jsonify(
        sorted(
            elo_predictor.elos.items(),
            key=lambda x: x[1],
            reverse=True
        )[0:25]
    )
