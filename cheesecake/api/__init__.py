from flask import Blueprint, jsonify
from sqlalchemy.orm import joinedload
from functools import lru_cache

from ..models import Team, Event, District, Match

api = Blueprint('api', __name__)

## TODO:
## This is very much a work in progress!

@api.route('elo')
def elo():
    matches = Match.query.all()
    return '{}'.format(len(matches))

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

@api.route('matches', methods=['GET'])
@lru_cache()
def get_all_matches():
    return jsonify(
        [x.as_dict() for x in Match.query.options(
            joinedload('alliances')
        ).all()]
    )


@api.route('matches/<string:event>', methods=['GET'])
def get_matches(event):
    return jsonify([x.as_dict() for x in Match.query.filter(
        Match.event_key == event
    ).all()])
