from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload
from functools import lru_cache
from operator import itemgetter
from scipy.stats import norm
from trueskill import Rating, BETA, global_env, rate
import itertools
import math

from ..models import Team, Event, District, Match, Alliance

import abc

api = Blueprint('api', __name__)

## TODO:
## This is very much a work in progress!
class Predictor(abc.ABC):
    def predict_match(self, match: Match) -> float:
        return 0.0

    def add_result(self, match: Match):
        return

    def current_values(self):
        return

class TeamNumberPredictor(Predictor):
    def predict_match(self, match: Match) -> float:
        alliances = match.get_alliances()
        blue = sum([x.team_number for x in alliances["blue"].team_keys])
        red = sum([x.team_number for x in alliances["red"].team_keys])
        return 1 - red / (red + blue)

class RedPredictor(Predictor):
    def predict_match(self, match: Match) -> float:
        if match.comp_level == "qm":
            return 0.5
        return 0.75

class EloPredictor(Predictor):
    def __init__(self):
        self.elos = {}
        self.k = 48

    def _get_elo(self, team) -> float:
        if team not in self.elos:
            self.elos[team] = 0.0
        return self.elos[team]

    def _alliance_elo(self, alliance: Alliance) -> float:
        return sum([self._get_elo(t.key) for t in alliance.team_keys])

    def predict_match(self, match: Match) -> float:
        alliances = match.get_alliances()
        red = self._alliance_elo(alliances["red"])
        blue = self._alliance_elo(alliances["blue"])
        return 1 / (1 + 10 ** ((blue - red) / 400.0))

    def add_result(self, match: Match):
        expected = self.predict_match(match)
        actual = match.result()
        change = self.k * (actual - expected)
        if match.comp_level != "qm":
            change /= 3
        alliances = match.get_alliances()
        for team in alliances["red"].team_keys:
            self.elos[team.key] += change
        for team in alliances["blue"].team_keys:
            self.elos[team.key] -= change

    def current_values(self):
        return self.elos

class TrueSkillPredictor(Predictor):
    def __init__(self):
        self.ratings = {}

    def _get_rating(self, team) -> Rating:
        if team not in self.ratings:
            self.ratings[team] = Rating()
        return self.ratings[team]

    def _alliance_ratings(self, alliance: Alliance):
        return [self._get_rating(t.key) for t in alliance.team_keys]

    def win_probability(self, team1, team2):
        beta = BETA
        delta_mu = sum(r.mu for r in team1) - sum(r.mu for r in team2)
        sum_sigma = sum(r.sigma ** 2 for r in itertools.chain(team1, team2))
        size = len(team1) + len(team2)
        denom = math.sqrt(size * (beta * beta) + sum_sigma)
        ts = global_env()
        return ts.cdf(delta_mu / denom)

    def predict_match(self, match: Match) -> float:
        alliances = match.get_alliances()
        red = self._alliance_ratings(alliances["red"])
        blue = self._alliance_ratings(alliances["blue"])
        return self.win_probability(red, blue)

    def add_result(self, match: Match):
        alliances = match.get_alliances()
        red = self._alliance_ratings(alliances["red"])
        blue = self._alliance_ratings(alliances["blue"])
        actual = match.diff()
        r_red, r_blue = rate([red, blue], ranks=[1-int(actual), int(actual)])
        for i, rating in enumerate(r_red):
            self.ratings[alliances["red"].team_keys[i].key] = rating
        for i, rating in enumerate(r_blue):
            self.ratings[alliances["blue"].team_keys[i].key] = rating

    def current_values(self):
        r = {}
        for key, value in self.ratings.items():
            r[key] = value.mu - 3 * value.sigma
        return r

class EloScorePredictor(EloPredictor):
    def add_result(self, match: Match):
        expected = self.predict_match(match)
        expected_score = norm.ppf(expected, loc=0, scale=225)
        actual = match.diff()
        change = self.k * (actual - expected_score) / 225
        if match.comp_level != "qm":
            change /= 3
        alliances = match.get_alliances()
        for team in alliances["red"].team_keys:
            self.elos[team.key] += change
        for team in alliances["blue"].team_keys:
            self.elos[team.key] -= change

@lru_cache(maxsize=1)
def fetch_all_matches():
    return Match.query.join(Event).filter(
        Event.event_type < 10
    ).options(
        joinedload('alliances')
    ).order_by(
        Match.time
    ).all()

@lru_cache()
def run_elo():
    matches = fetch_all_matches()
    predictor = TrueSkillPredictor()
    for match in matches:
        predictor.add_result(match)
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
    return jsonify(scores)
        

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
    #return jsonify(results)
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
        #if not(match.comp_level == "qm" and match.match_number < 30):
        if True == True:
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
        results.append({
            "predicted": predictor.predict_match(match),
            "actual": match.result()
        })
        predictor.add_result(match)
    return "{} {}".format(
        len([x for x in results if x["actual"] != 0.5 and abs(x["predicted"] - x["actual"]) < 0.5]) / len(results),
        sum((x["actual"] - x["predicted"]) ** 2 for x in results) / len(results)
    )

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
@lru_cache()
def get_all_matches():
    return jsonify(
        [x.as_dict() for x in Match.query.options(
            joinedload('alliances')
        ).all()]
    )

@api.route('matches/official', methods=['GET'])
def get_official_matches():
    matches = fetch_all_matches()
    return jsonify([x.as_dict() for x in matches])

@api.route('matches/<string:event>', methods=['GET'])
def get_matches(event):
    return jsonify([x.as_dict() for x in Match.query.filter(
        Match.event_key == event
    ).all()])
