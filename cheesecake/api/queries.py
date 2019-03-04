from sqlalchemy.orm import joinedload

from .. import cache, db
from ..models import *
from .times import *

MATCH_ORDER = {
    "qm": 0,
    "ef": 10,
    "qf": 11,
    "sf": 12,
    "f": 13
}
sort_order = db.case(value=Match.comp_level, whens=MATCH_ORDER)

@cache.memoize(timeout=MINUTE)
def fetch_matches(year=None):
    matches =  Match.query.join(Event).filter(
        Event.event_type < 10
    )
    if year is not None:
        matches = matches.filter(
            Event.year == year
        )
    matches = matches.options(
        joinedload('alliances')
    ).order_by(
        Event.start_date,
        Match.time,
        sort_order,
        Match.match_number
    ).all()
    return matches
