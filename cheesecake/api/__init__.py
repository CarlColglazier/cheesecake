from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload, contains_eager
from functools import lru_cache
from operator import itemgetter
import pickle
import numpy as np
import pandas as pd
from scipy.stats import norm

import itertools
import math

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

@api.route('elo')
def elo():
    predictor = run_elo()
    scores = []
    for key, value in sorted(predictor.current_values().items(), key=itemgetter(1), reverse=True):
        scores.append({
            "key": key,
            "score": value
        })
    return jsonify(scores[0:100])

@api.route('elo/loc/<string:key>')
def get_loc_ranks(key):
    predictor = run_elo()
    scores = []
    team_keys = [x.key for x in Team.query.filter(Team.state_prov == key).all()]
    for key, value in sorted(predictor.current_values().items(), key=itemgetter(1), reverse=True):
        if key in team_keys:
            scores.append({
                "key": key,
                "score": value
            })
    return jsonify(scores)

@api.route('elo/<string:key>')
def get_team_elo(key):
    predictor = run_elo()
    match_keys = [x.match.key for x in Team.query.get(key).alliances]
    predictions = {}
    for key in match_keys:
        if key in predictor.prediction_history:
            """
            predictions.append({
                "key": key,
                "prediction": predictor.prediction_history[key]
            })
            """
            predictions[key] = predictor.prediction_history[key]
    return jsonify(predictions)
        
@api.route('predict/red')
def predict_red():
    matches = fetch_all_matches()
    predictor = RedPredictor()
    results = []
    for match in matches:
        if match.key[0:4] == "2018" and "_qm" in match.key:
            results.append({
                "predicted": predictor.predict_match(match),
                "actual": match.result()
            })
    return "{}\n{}".format(
        len([x for x in results if x["actual"] == 1]) / len(results),
        sum([(x["actual"] - x["predicted"]) ** 2 for x in results]) / len(results)
    )

@api.route('predict/numbers')
def predict_team_numbers():
    matches = fetch_all_matches()
    results = []
    predictor = TeamNumberPredictor()
    for match in matches:
        if match.key[0:4] == "2018":
            results.append({
                "predicted": predictor.predict_match(match),
                "actual": match.result()
            })
    return "{} {}".format(
        len([x for x in results if x["actual"] != 0.5 and abs(x["predicted"] - x["actual"]) < 0.5]) / len(results),
        sum((x["actual"] - x["predicted"]) ** 2 for x in results) / len(results)
    )

@api.route('predict/hybrid')
def predict_hybrid():
    matches = fetch_all_matches()
    results = []
    predictor1 = EloPredictor()
    predictor2 = EloScorePredictor()
    for match in matches:
        if match.key[0:4] == "2018":
            p2 = predictor2.predict_match(match)
            p1 = predictor1.predict_match(match)
            results.append({
                "predicted": 0.25 * p1 + 0.75 * p2,
                "actual": match.result()
            })
        predictor1.add_result(match)
        predictor2.add_result(match)
    return "{} {}".format(
        len([x for x in results if x["actual"] != 0.5 and abs(x["predicted"] - x["actual"]) < 0.5]) / len(results),
        sum((x["actual"] - x["predicted"]) ** 2 for x in results) / len(results)
    )
    
        
@api.route('predict/elo')
def predict_elo():
    matches = fetch_all_matches()
    results = []
    predictor = EloPredictor()
    for match in matches:
        if match.key[0:4] == "2018":
            results.append({
                "predicted": predictor.predict_match(match),
                "actual": match.result()
            })
        predictor.add_result(match)
    return "{} {}".format(
        len([x for x in results if x["actual"] != 0.5 and abs(x["predicted"] - x["actual"]) < 0.5]) / len(results),
        sum((x["actual"] - x["predicted"]) ** 2 for x in results) / len(results)
    )

@api.route('predict/eloscore')
def predict_elo_score():
    matches = fetch_all_matches()
    results = {}
    predictor = EloScorePredictor()
    for match in matches:
        if match.key[0:4] == "2018":
            if match.result() is not None:
                p = predictor.predict_match(match)
                results[match.key] = {
                    "predicted": p,
                    "actual": match.result(),
                    "diff": match.diff(),
                    "pdiff": norm.ppf(p, scale=predictor.stds["playoffs"]["2018"])
                }
        predictor.add_result(match)
    match_df = pd.DataFrame.from_dict(results, orient="index")
    groups = match_df.groupby(match_df["predicted"].apply(lambda x: round(x, 1)))
    table = groups.sum() / groups.count()
    ws = match_df[match_df["actual"] != 0.5]
    return "{}\n{}\n{}\n{}".format(
        table.to_html(),
        table.count(axis=1),
        ((ws["predicted"] - ws["actual"]) ** 2).sum() / len(ws),
        ((match_df["predicted"] - match_df["actual"]) ** 2).sum() / len(match_df)
    )

"""
@api.route('predict/eloscore/team/<string:team>')
def predict_elo_score_team(team):
    matches = fetch_all_matches()
    results = {}
    predictor = EloScorePredictor()
    team_matches = [x.match.as_dict() for x in Team.query.get(team).alliances]
    keys = [x.match.key for x in Team.query.get(team).alliances]
    for match in matches:
        prediction = predictor.predict_match(match),
        predictor.add_result(match)
        if match.key in keys:
            results[match.key] = prediction
    for match in team_matches:
        match["prediction"] = results[match["key"]]
    return jsonify(sorted(team_matches, key=lambda x: x["actual_time"]))

# TODO: This is broken.
@api.route('predict/ts')
def predict_trueskill_score():
    matches = fetch_all_matches()
    results = []
    predictor = TrueSkillPredictor()
    for match in matches:
        if match.key[0:4] == "2018":
            results.append({
                "predicted": predictor.predict_match(match),
                "actual": match.result()
            })
        predictor.add_result(match)
    return "{} {}".format(
        len([x for x in results if x["actual"] != 0.5 and abs(x["predicted"] - x["actual"]) < 0.5]) / len(results),
        sum((x["actual"] - x["predicted"]) ** 2 for x in results) / len(results)
    )
"""

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

@api.route('districts', methods=['GET'])
def get_districts():
    return jsonify([x.as_dict() for x in District.query.all()])

@api.route('events', methods=['GET'])
def get_events():
    return jsonify([x.as_dict() for x in Event.query.all()])

@api.route('events/official', methods=['GET'])
def get_official_events():
    events = Event.query.filter(Event.first_event_code != None).all()
    return jsonify([x.as_dict() for x in events])

@api.route('matches', methods=['GET'])
def get_official_matches():
    matches = fetch_all_matches()
    return jsonify([x.as_dict() for x in matches])

@api.route('matches/<string:event>', methods=['GET'])
def get_matches(event):
    return jsonify([x.as_dict() for x in Match.query.filter(
        Match.event_key == event
    ).all()])

@api.route('team/<string:key>/matches', methods=['GET'])
def get_team_matches(key):
    m = Team.query.join(Team.alliances).options(
        contains_eager('alliances')
    ).filter(
        Team.key == key
    ).filter(
        Alliance.key.like("2018%")
    ).all()
    if len(m) == 0:
        return jsonify([])
    team = m[0]
    matches = [x.match for x in team.alliances]
    matches = sorted(matches, key=lambda x: x.actual_time or -1)
    keys = [x.key for x in matches]
    alliances = Alliance.query.filter(
        Alliance.match_key.in_(keys)
    ).all()
    alliance_table = {}
    for alliance in alliances:
        if alliance.match_key not in alliance_table:
            alliance_table[alliance.match_key] = {}
        alliance_table[alliance.match_key][alliance.color] = alliance.as_dict()
    matches = [x.serialize for x in matches]
    for match in matches:
        match["alliances"] = alliance_table[match["key"]]
    return jsonify(matches)
