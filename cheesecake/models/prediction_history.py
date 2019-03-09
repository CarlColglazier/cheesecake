from .. import db

class PredictionHistory(db.Model):
    key = db.Column(db.Integer, primary_key=True, autoincrement=True)
    match = db.Column(db.String(25), db.ForeignKey('match.key'))
    prediction = db.Column(db.Float)
    model = db.Column(db.String(100))

    @property
    def serialize(self):
        return {
            "prediction": self.prediction,
            "model": self.model,
        }
