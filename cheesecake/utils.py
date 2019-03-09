from . import db, tba
from .models import Team, Event, Match, District, Alliance

# This function is huge and needs to be broken up.
def update_match(match):
    match["alliances"]["red"]["color"] = "red"
    match["alliances"]["blue"]["color"] = "blue"
    match["alliances"]["red"]["match_key"] = match["key"]
    match["alliances"]["blue"]["match_key"] = match["key"]
    match["alliances"]["red"]["key"] = match["key"] + "_red"
    match["alliances"]["blue"]["key"] = match["key"] + "_blue"
    if "team_keys" in match["alliances"]["red"]:
        red_teams = match["alliances"]["red"]["team_keys"]
        blue_teams = match["alliances"]["blue"]["team_keys"]
        del match["alliances"]["red"]["team_keys"]
        del match["alliances"]["blue"]["team_keys"]
    else:
        # Webhook
        red_teams = match["alliances"]["red"]["teams"]
        blue_teams = match["alliances"]["blue"]["teams"]
        del match["alliances"]["red"]["teams"]
        del match["alliances"]["blue"]["teams"]
        if match["alliances"]["red"]["score"] > match["alliances"]["blue"]["score"]:
            match["winning_alliance"] = "red"
        elif match["alliances"]["red"]["score"] < match["alliances"]["blue"]["score"]:
            match["winning_alliance"] = "blue"
        else:
            match["winning_alliance"] = ""
    red = Alliance.query.get(match["alliances"]["red"]["key"])
    blue = Alliance.query.get(match["alliances"]["blue"]["key"])
    if red is None:
        red = Alliance(**match["alliances"]["red"])
    else:
        red.score = match["alliances"]["red"]["score"]
    if blue is None:
        blue = Alliance(**match["alliances"]["blue"])
    else:
        blue.score = match["alliances"]["blue"]["score"]
    for team in red_teams:
        t = Team.query.get(team)
        if t and t not in red.team_keys:
            red.team_keys.append(t)
    for team in blue_teams:
        t = Team.query.get(team)
        if t and t not in blue.team_keys:
            blue.team_keys.append(t)
    del match["alliances"]
    m = Match.query.get(match["key"])
    if m is None:
        m = Match(**match)
    for c in m.__table__.columns:
        cname = c.name
        if "time" not in cname:
            setattr(m, cname, match[cname])
    db.session.merge(m)    
    db.session.commit()
    return m

def update_schedule(event):
    matches = tba.event_matches(event)
    for match in matches:
        update_match(match)
    db.session.commit()

