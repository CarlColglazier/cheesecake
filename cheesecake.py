from flask import Flask, Blueprint, jsonify
from flask_cors import CORS
from flask_script import Manager
from flask_migrate import Migrate, MigrateCommand
from flask_sqlalchemy import SQLAlchemy
from flask_socketio import SocketIO, emit
from secret import TBA_KEY
import tbapy

CURRENT_YEAR = 2018

app = Flask(__name__)
cors = CORS(app, resources={r"/*":{"origins":"http://localhost:3000"}})
tba = tbapy.TBA(TBA_KEY)
#tasks = Blueprint('tasks', __name__)
api = Blueprint('api', __name__)

app.config["SQLALCHEMY_DATABASE_URI"] = "sqlite:///app.db"
# TODO: Change this.
app.config["SECRET_KEY"] = "secret!"

db = SQLAlchemy(app)
socketio = SocketIO(app)
migrate = Migrate(app, db)
manager = Manager(app)
manager.add_command('db', MigrateCommand)

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
    Name = db.Column(db.String(250))
    # TODO: Implement this.
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

class Match(db.Model):
    key = db.Column(db.String(25), primary_key=True)
    comp_level = db.Column(db.String(2))
    set_number = db.Column(db.Integer)
    match_number = db.Column(db.Integer)
    # TODO
    alliances = None
    winning_alliance = db.Column(db.String(3))
    event_key = None#db.relationship('Event', backref='matches', lazy=True)
    time = db.Column(db.Integer)
    actual_time = db.Column(db.Integer)
    predicted_time = db.Column(db.Integer)
    post_result_time = db.Column(db.Integer)
    # TODO
    score_breakdown = None

class Award(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100))
    award_type = db.Column(db.Integer)
    event_key = db.Column(db.String(25), db.ForeignKey('event.key'))
    recipient_list = None#db.relationship('Teams', backref='awards', lazy=True)
    year = db.Column(db.Integer)

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

    @property
    def serialize(self):
        return {
            "key": self.key,
            "nickname": self.nickname
        }

"""
@app.route('/')
def hello_world():
    return 'Hello, World!'


@tasks.route('/status')
def status():
    return jsonify(tba.status())

@tasks.route('/teams')
def scrape_teams():
    teams = []
    # TODO: This probably shouldn't be hard coded.
    for i in range(20):
        teams += tba.teams(i)
    for team in teams:
        db.session.merge(Team(**team))
    db.session.commit()
    return jsonify(teams)
"""

@socketio.on('teams')
def get_teams():
    for i in range(20):
        teams = tba.teams(i)
        if len(teams) == 0:
            break
        for team in teams:
            db.session.merge(Team(**team))
        db.session.commit()
        emit('teams', teams)

"""
@tasks.route('events/<int:year>')
def scrape_events(year):
    return jsonify(tba.events(year))
"""

@socketio.on('events')
def get_events():
    for event in tba.events(CURRENT_YEAR):
        db.session.merge(Event(**event))
        emit('events', event)
    db.session.commit()

"""
@tasks.route('districts/<int:year>')
def scrape_districts(year):
    districts = tba.districts(year)
    for district in districts:
        db.session.merge(District(**district))
    db.session.commit()
    return jsonify(districts)
"""

@socketio.on('districts')
def get_districts():
    for district in tba.districts(CURRENT_YEAR):
        db.session.merge(District(**district))
    db.session.commit()

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

#app.register_blueprint(tasks, url_prefix='/tasks')
app.register_blueprint(api, url_prefix='/api')

if __name__ == '__main__':
    manager.run()
    socketio.run()
