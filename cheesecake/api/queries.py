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

def fetch_matches(year=None):
    matches =  Match.query.join(Event).filter(
        Event.event_type < 10
    )
    if year is not None:
        matches = matches.filter(
            Event.year == year
        )
    matches = matches.options(
        joinedload('alliances'),
        joinedload('predictions')
    )
    matches = matches.order_by(
        Event.start_date,
        Match.time,
        sort_order,
        Match.match_number
    )
    count = matches.count()
    page_size = 500
    for i in range(int(count / page_size)):
        matches = matches.limit(page_size)
        matches = matches.offset(page_size * i)
        print(i)
        yield matches.all()
