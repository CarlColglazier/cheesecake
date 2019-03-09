import abc
from .models import *
from scipy.stats import norm

## TODO:
## This is very much a work in progress!
class Predictor(abc.ABC):
    def __init__(self):
        pass

    def predict_match(self, match: Match):
        return 0.0

    def predict(self, match: Match):
        p = self.predict_match(match)
        return p

    def predict_keys(self, keys: list) -> float:
        return 0.0

    def add_result(self, match: Match):
        return

    def current_values(self):
        return

class BetaPredictor(Predictor):
    def __init__(self, a, b, feature):
        self.a = a
        self.b = b
        self.feature = feature
        self.teams = {}

    def predict_match(self, match: Match):
        predictions = {}
        for alliance in match.alliances:
            current = 0
            for team in alliance.team_keys:
                if team.key not in self.teams:
                    self.teams[team.key] = {
                        "attempts": 0,
                        "completions": 0
                    }
                score = self.teams[team.key]
                calc = (self.a + score["completions"]) / (self.b + self.a + score["attempts"])
                current = max(current, calc)
            predictions[self.feature + alliance.color] = current
        return predictions

    def add_result(self, match: Match):
        if match.comp_level != "qm":
            return
        for alliance in match.alliances:
            for team in alliance.team_keys:
                if team.key not in self.teams:
                    self.teams[team.key] = {
                        "attempts": 0,
                        "completions": 0
                    }
                self.teams[team.key]["attempts"] += 1
                if match.score_breakdown[alliance.color][self.feature]:
                    self.teams[team.key]["completions"] += 1

class EloPredictor(Predictor):
    def __init__(self):
        super().__init__()
        self.elos = {}
        self.k = 12

    def _get_elo(self, team) -> float:
        if team not in self.elos:
            self.elos[team] = 0.0
        return self.elos[team]

    def _alliance_elo(self, alliance: Alliance) -> float:
        return sum([self._get_elo(t.key) for t in alliance.team_keys])

    def predict_keys(self, keys: list) -> float:
        #TODO: This assumes that it is used correctly.
        red = sum([self._get_elo(x) for x in keys[0:3]])
        blue = sum([self._get_elo(x) for x in keys[3:6]])
        return 1 / (1 + 10 ** ((blue - red) / 400.0))

    def predict_match(self, match: Match) -> float:
        alliances = match.get_alliances()
        red = self._alliance_elo(alliances["red"])
        blue = self._alliance_elo(alliances["blue"])
        return 1 / (1 + 10 ** ((blue - red) / 400.0))

    def add_result(self, match: Match):
        expected = self.predict_match(match)
        actual = match.result()
        if actual is None or expected is None:
            return
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

class EloScorePredictor(EloPredictor):
    def __init__(self):
        super().__init__()
        self.stds = {
            "quals": {
                "2003": 51.7,
                "2004": 44.3,
                "2005": 28.6,
                "2006": 32.8,
                "2007": 47.1,
                "2008": 30.7,
                "2009": 31.7,
                "2010": 4.0,#
                "2011": 36.0,
                "2012": 20.7,
                "2013": 42.7,
                "2014": 75.6,
                "2015": 36.2,
                "2016": 29.5,
                "2017": 91.4,
                "2018": 184.7,
                # TODO: This is temporary.
                "2019": 21.1,
            }, "playoffs": {
                "2003": 50.9,
                "2004": 45.6,
                "2005": 15.5,
                "2006": 20.5,
                "2007": 32.9,
                "2008": 24.4,
                "2009": 21.0,
                "2010": 2.7,
                "2011": 28.4,
                "2012": 15.5,
                "2013": 31.1,
                "2014": 49.3,
                "2015": 33.2,
                "2016": 27.5,
                "2017": 70.6,
                "2018": 106.9,
                # TODO: This is temporary.
                "2019": 18.8,
            }
        }
        self.last_year = "2002"

    def _dampen(self):
        # TODO: Consider using different values for different years.
        for key, value in self.elos.items():
            self.elos[key] = 0.5 * value + 0.1 * 150

    def _elo_diff(self, match: Match):
        alliances = match.get_alliances()
        red = self._alliance_elo(alliances["red"])
        blue = self._alliance_elo(alliances["blue"])
        return red - blue
            
    def predict_match(self, match: Match) -> float:
        return 1 / (1 + 10 ** ((-self._elo_diff(match)) / 400.0))
        
    def add_result(self, match: Match):
        year = match.key[0:4]
        odds = self.predict_match(match)
        # should this happen here?
        if self.last_year != year:
            self.last_year = year
            self._dampen()
        scale = self.stds["playoffs"][year]
        actual = match.diff()
        expected_score = norm.ppf(odds, scale=scale)
        change = self.k * (actual - expected_score) / scale
        alliances = match.get_alliances()
        for team in alliances["red"].team_keys:
            self.elos[team.key] += change
        for team in alliances["blue"].team_keys:
            self.elos[team.key] -= change

