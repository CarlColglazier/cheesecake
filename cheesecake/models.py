from . import db

class District(db.Model):
    abbreviation = db.Column(db.String(10))
    display_name = db.Column(db.String(100))
    key = db.Column(db.String(10), primary_key=True)
    year = db.Column(db.Integer)

    def as_dict(self):
       return {c.name: getattr(self, c.name) for c in self.__table__.columns}

class Event(db.Model):
    address = db.Column(db.String(1000))
    city = db.Column(db.String(100))
    country = db.Column(db.String(50))
    # TODO: Implement this?
    # It looks like it's a rather simple one-to-many relationship.
    division_keys = None
    # TOOD: Should this be a different type of object?
    end_date = db.Column(db.String(25))
    event_code = db.Column(db.String(10))
    event_type = db.Column(db.Integer)
    event_type_string = db.Column(db.String(50))
    first_event_code = db.Column(db.String(25))
    first_event_id = db.Column(db.String(50))
    gmaps_place_id = db.Column(db.String(100))
    gmaps_url = db.Column(db.String(100))
    key = db.Column(db.String(10), primary_key=True)
    lat = db.Column(db.Float)
    lng = db.Column(db.Float)
    location_name = db.Column(db.String(100))
    name = db.Column(db.String(250))
    # TODO: Implement this.
    district = None
    parent_event_key = None
    playoff_type = db.Column(db.Integer)
    playoff_type_string = db.Column(db.Integer)
    postal_code = db.Column(db.String(50))
    short_name = db.Column(db.String(250))
    start_date = db.Column(db.String(10))
    state_prov = db.Column(db.String(50))
    timezone = db.Column(db.String(50))
    webcasts = None
    website = db.Column(db.String(100))
    week = db.Column(db.Integer)
    year = db.Column(db.Integer)
    matches = db.relationship('Match', backref='matches', lazy=True)
    awards = db.relationship('Award', backref='awards', lazy=True)

    def as_dict(self):
       return {c.name: getattr(self, c.name) for c in self.__table__.columns}

class Match(db.Model):
    key = db.Column(db.String(25), primary_key=True)
    comp_level = db.Column(db.String(2))
    set_number = db.Column(db.Integer)
    match_number = db.Column(db.Integer)
    # TODO
    alliances = None
    winning_alliance = db.Column(db.String(3))
    event_key = db.Column(db.String(25), db.ForeignKey('event.key'))
    time = db.Column(db.Integer)
    actual_time = db.Column(db.Integer)
    predicted_time = db.Column(db.Integer)
    post_result_time = db.Column(db.Integer)
    # TODO
    score_breakdown = None
    videos = None

    def as_dict(self):
       return {c.name: getattr(self, c.name) for c in self.__table__.columns}

class Award(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100))
    award_type = db.Column(db.Integer)
    event_key = db.Column(db.String(25), db.ForeignKey('event.key'))
    # TODO: Implement this.
    recipient_list = None
    year = db.Column(db.Integer)

    def as_dict(self):
       return {c.name: getattr(self, c.name) for c in self.__table__.columns}

class Team(db.Model):
    key = db.Column(db.String(8), primary_key=True)
    team_number = db.Column(db.Integer)
    nickname = db.Column(db.String(100))
    name = db.Column(db.String(1_000))
    city = db.Column(db.String(100))
    state_prov = db.Column(db.String(100))
    country = db.Column(db.String(100))
    address = db.Column(db.String(1000))
    postal_code = db.Column(db.String(25))
    website = db.Column(db.String(250))
    rookie_year = db.Column(db.Integer)
    motto = db.Column(db.String(250))
    # These will likely need to be added in the future:
    gmaps_place_id = None
    gmaps_url = None
    lat = None
    lng = None
    location_name = None
    # This one still needs to be implemented.
    home_championship = None

    def as_dict(self):
       return {c.name: getattr(self, c.name) for c in self.__table__.columns}

    @property
    def serialize(self):
        return {
            "key": self.key,
            "nickname": self.nickname
        }
