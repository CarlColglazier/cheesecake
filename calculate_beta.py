from cheesecake.models import *
import itertools
from scipy.stats import beta
from sqlalchemy.orm import joinedload
import pandas as pd

#events = Event.query.filter(Event.year == 2019).filter(Event.week == 0).all()
matches = Match.query.options(joinload(Match.event_key)).filter(
    Event.year == 2019
).filter(
    Event.week == 0
).all()
#matches = list(itertools.chain.from_iterable([x.matches for x in events]))
matches = [x for x in matches if x.winning_alliance != "" and x.comp_level == "qm"]
colors = ['red', 'blue']
teams = {}
blank = {"sucess": 0, "attempts": 0}
for match in matches:
    alliances = match.get_alliances()
    for color in colors:
        #sucess = match.score_breakdown[color]["habDockingRankingPoint"]
        sucess = match.score_breakdown[color]["completeRocketRankingPoint"]
        for team in alliances[color].team_keys:
            key = team.key
            if key not in teams:
                teams[key] = blank.copy()
            teams[key]["attempts"] += 1
            if sucess:
                teams[key]["sucess"] += 1


data = pd.DataFrame.from_dict(teams, orient='index')
print(beta.fit(data["sucess"]/data["attempts"], loc=0, scale=1))

# a=0.7229 b=2.4517
# a=0.2749 b=62.2755
