from enum import Enum

class EventState(Enum):
    NO_SCHEDULE = 0
    NOT_STARTED = 1
    QUALIFICATIONS = 2
    ALLIANCE_SELECTION = 3
    PLAYOFFS = 4
    AWARDS = 5
    COMPLETE = 6
