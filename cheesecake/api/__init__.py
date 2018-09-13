from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload, contains_eager
from functools import lru_cache
from operator import itemgetter
import pickle
import numpy as np

import itertools
import math

from ..models import *
from ..predictors import *

api = Blueprint('api', __name__)

_whens = {
    "qm": 0,
    "ef": 10,
    "qf": 11,
    "sf": 12,
    "f": 13
}
sort_order = db.case(value=Match.comp_level, whens=_whens)

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
        results.append({
            "predicted": predictor.predict_match(match),
            "actual": match.result()
        })
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
    results = []
    predictor = EloScorePredictor()
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

@api.route('predict/ts')
def predict_trueskill_score():
    matches = fetch_all_matches()
    results = []
    predictor = TrueSkillPredictor()
    for match in matches:
        results.append({
            "predicted": predictor.predict_match(match),
            "actual": match.result()
        })
        predictor.add_result(match)
    return "{} {}".format(
        len([x for x in results if x["actual"] != 0.5 and abs(x["predicted"] - x["actual"]) < 0.5]) / len(results),
        sum((x["actual"] - x["predicted"]) ** 2 for x in results) / len(results)
    )

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
        

@lru_cache()
def get_18_auto():
    results = {}
    keys = Team.query.with_entities(Team.key).all()
    keys = [x[0] for x in keys]
    for key in keys:
        results[key] = {
            "move": 0,
            "switch": 0,
            "moveswitch": 0,
            "attempts": 0
        }
    matches = fetch_all_matches()
    for match in matches:
        for alliance in match.alliances:
            color = alliance.color
            for i, team in enumerate(alliance.team_keys):
                key = team.key
                auto_string = "autoRobot{}".format(i + 1)
                results[key]["attempts"] += 1
                if match.score_breakdown[color][auto_string] == "AutoRun":
                    results[key]["move"] += 1
                if match.score_breakdown[color]["autoSwitchAtZero"]:
                    results[key]["switch"] += 1
                results[key]["moveswitch"] += (
                    match.score_breakdown[color][auto_string] and
                    match.score_breakdown[color]["autoSwitchAtZero"]
                )
    return results

@api.route('2018/auto')
def api_get_18_auto():
    results = get_18_auto()
    l = []
    for key, result in results.items():
        result["team"] = key
        l.append(result)
    return jsonify(l)

@api.route('event/<string:key>/predict/auto')
def predict_event_auto(key):
    alpha = 4.68
    beta = 0.84
    num_trials = 10_000
    matches = Event.query.get(key).matches
    results = get_18_auto()
    predictions = {}
    for match in matches:
        alliances = match.get_alliances()
        predictions[match.key] = {}
        for color, alliance in alliances.items():
            prob_move = np.ones(num_trials)
            prob_switch = np.zeros(num_trials)
            for team in alliance.team_keys:
                auto = results[team.key]
                a_move = 4.68 + auto["move"]
                b_move = 0.84 + auto["attempts"] - auto["move"]
                a_switch = 4.22 + auto["switch"]
                b_switch = 3.04 + auto["attempts"] - auto["switch"]
                prob_move *= np.random.beta(a_move, b_move, size=num_trials)
                prob_switch = np.maximum(
                    prob_switch,
                    np.random.beta(a_switch, b_switch, size=num_trials)
                )
            predictions[match.key][color] = np.median(prob_move * prob_switch)
    return jsonify(predictions)

    
@api.route('2018/endgame')
def get_18_endgame():
    results = {}
    keys = Team.query.with_entities(Team.key).all()
    keys = [x[0] for x in keys]
    for key in keys:
        results[key] = {
            "runs": 0,
            "attempts": 0
        }
    matches = fetch_all_matches()
    for match in matches:
        for alliance in match.alliances:
            color = alliance.color
            for i, team in enumerate(alliance.team_keys):
                key = team.key
                auto_string = "endgameRobot{}".format(i + 1)
                results[key]["attempts"] += 1
                if match.score_breakdown[color][auto_string] == "Climbing":
                    results[key]["runs"] += 1
    l = []
    for key, result in results.items():
        result["team"] = key
        l.append(result)
    return jsonify(l)
