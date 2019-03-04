from .. import db
from .tables import district_teams

class District(db.Model):
    abbreviation = db.Column(db.String(10))
    display_name = db.Column(db.String(100))
    key = db.Column(db.String(10), primary_key=True)
    year = db.Column(db.Integer)
    teams = db.relationship(
        'Team',
        secondary=district_teams,
        lazy='subquery',
        order_by=district_teams.c.position,
        backref=db.backref('districts', lazy=True))

    def as_dict(self):
       return {c.name: getattr(self, c.name) for c in self.__table__.columns}
