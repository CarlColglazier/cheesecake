import abc
from math import sqrt
from .models import *
from scipy.stats import norm, beta
from trueskill import Rating, BETA, global_env, rate

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
        self.k = 16

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
        denom = sqrt(size * (beta * beta) + sum_sigma)
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
    def __init__(self):
        super().__init__()
        self.stds = {
            "2003": 50.9,
            "2004": 45.6,
            "2005": 24.6,
            "2006": 28.4,
            "2007": 46.2,
            "2008": 24.4,
            "2009": 21.0,
            "2010": 2.7,
            "2011": 28.4,
            "2012": 15.5,
            "2013": 31.1,
            "2014": 49.3,
            "2015": 33.2,
            "2016": 47.0,
            "2017": 95.0,
            "2018": 225.0
        }
        self.last_year = "2002"

    def _dampen(self):
        for key, value in self.elos.items():
            self.elos[key] = 0.9 * value + 0.1 * 150
        
    def add_result(self, match: Match):
        expected = self.predict_match(match)
        year = match.key[0:4]
        if self.last_year != year:
            self.last_year = year
            self._dampen()
        scale = self.stds[year]
        expected_score = norm.ppf(expected, loc=0, scale=scale)
        if match.comp_level != "qm":
            expected_score = norm.ppf(expected, loc=0, scale=scale/2)
        actual = match.diff()
        change = self.k * (actual - expected_score) / scale
        #if match.comp_level != "qm":
        #    change /= 2
        alliances = match.get_alliances()
        for team in alliances["red"].team_keys:
            self.elos[team.key] += change
        for team in alliances["blue"].team_keys:
            self.elos[team.key] -= change
