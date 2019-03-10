from .. import db
from .tables import alliance_teams

class Alliance(db.Model):
    key = db.Column(db.String(25), primary_key=True)
    score = db.Column(db.Integer)
    color = db.Column(db.String(10))
    team_keys = db.relationship(
        'Team',
        secondary=alliance_teams,
        lazy='subquery',
        order_by=alliance_teams.c.position,
        backref=db.backref('alliances', lazy=True))
    match_key = db.Column(db.String(25), db.ForeignKey('match.key'))
    surrogate = None
    surrogate_team_keys = None
    dq_team_keys = None

    def as_dict(self):
        d = {}
        d["score"] = self.score
        d["color"] = self.color
        d["team_keys"] = [x.key for x in self.team_keys]
        return d

    def team_number_sum(self):
        return sum([x.team_number for x in self.team_keys])
