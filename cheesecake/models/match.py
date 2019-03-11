from .. import db
from sqlalchemy.dialects.postgresql import JSON

class Match(db.Model):
    key = db.Column(db.String(25), primary_key=True)
    comp_level = db.Column(db.String(2))
    set_number = db.Column(db.Integer)
    match_number = db.Column(db.Integer)
    alliances = db.relationship('Alliance', backref='match')
    predictions = db.relationship('PredictionHistory', backref='pmatch')
    winning_alliance = db.Column(db.String(5))
    event_key = db.Column(db.String(25), db.ForeignKey('event.key'))
    time = db.Column(db.Integer)
    actual_time = db.Column(db.Integer)
    predicted_time = db.Column(db.Integer)
    post_result_time = db.Column(db.Integer)
    # TODO
    score_breakdown = db.Column(JSON)
    videos = None

    def as_dict(self):
        d =  {c.name: getattr(self, c.name) for c in self.__table__.columns}
        d["alliances"] = [x.as_dict() for x in self.alliances]
        return d

    def get_prediction(self, model):
        for prediction in self.predictions:
            if prediction.model == model:
                return prediction
        return None

    def get_alliances(self):
        alliances = {}
        for alliance in self.alliances:
            alliances[alliance.color] = alliance
        return alliances

    @property
    def serialize(self):
        alliances = self.get_alliances()
        for key in alliances:
            alliances[key] = alliances[key].as_dict()
        preds = {}
        predictions = [x.serialize for x in self.predictions]
        for p in predictions:
            preds[p["model"]] = p["prediction"]
        breakdown = {
            "red": {},
            "blue": {}
        }
        for color, dic in breakdown.items():
            if not self.score_breakdown:
                continue
            for key, value in self.score_breakdown[color].items():
                if 'RankingPoint' in key:
                    dic[key] = value
        return {
            "key": self.key,
            "comp_level": self.comp_level,
            "match_number": self.match_number,
            "winning_alliance": self.winning_alliance,
            "alliances": alliances,
            "predictions": preds,
            "score_breakdown": breakdown
        }


    def result(self):
        """
        Who won the match? 1 for red, 0 for blue, 0.5 for a tie.
        None if something goes wrong or if the match hasn't been
        played yet.
        """
        alliances = self.get_alliances()
        if "red" not in alliances or "blue" not in alliances:
            return None
        # TODO: Should there be a function to check if a match has
        # been played yet?
        if alliances["red"].score == -1 and alliances["blue"].score == -1:
            return None
        if alliances["red"].score > alliances["blue"].score:
            return 1
        elif alliances["red"].score < alliances["blue"].score:
            return 0
        else:
            return 0.5

    def diff(self):
        alliances = self.get_alliances()
        if "red" not in alliances or "blue" not in alliances:
            return None
        return  alliances["red"].score -  alliances["blue"].score
