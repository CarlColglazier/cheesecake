import requests
import itertools
import pandas as pd
from dotenv import load_dotenv
import os

load_dotenv()

API = "https://www.thebluealliance.com/api/v3"
headers = {"X-TBA-Auth-Key":os.environ.get("KEY")}
"""
Get keys for events in the year
"""
def get_all_events_year(year):
	url = f"{API}/events/{year}"
	r = requests.get(url, headers=headers)
	j = r.json()
	j = filter(lambda x: x['event_type'] <= 6, j)
	return list(j)


def get_event(key):
	url = f"{API}/event/{key}"
	r = requests.get(url, headers=headers)
	return r.json()

def get_matches(event):
	event_key = event['key']
	url = f"{API}/event/{event_key}/matches"
	r = requests.get(url, headers=headers)
	j = r.json()
	for m in j:
		if event['week'] is None:
			m['week'] = 10
		else:
			m['week'] = event['week']
		m['event_type'] = event['event_type']
	return j

def proc_match(m):
	g = []
	if m['score_breakdown'] == None:
		return g
	for a in ['red', 'blue']:
		b = m['score_breakdown'][a]
		bd_autocargou =  sum([b[y] for y in ['autoCargoUpperBlue', 'autoCargoUpperFar', 'autoCargoUpperNear', 'autoCargoUpperRed']])
		bd_telecargou =  sum([b[y] for y in ['teleopCargoUpperBlue', 'teleopCargoUpperFar', 'teleopCargoUpperNear', 'teleopCargoUpperRed']])
		bd_autocargol =  sum([b[y] for y in ['autoCargoLowerBlue', 'autoCargoLowerFar', 'autoCargoLowerNear', 'autoCargoLowerRed']])
		bd_telecargol =  sum([b[y] for y in ['teleopCargoLowerBlue', 'teleopCargoLowerFar', 'teleopCargoLowerNear', 'teleopCargoLowerRed']])
		for i, t in enumerate(m['alliances'][a]['team_keys']):
			r = [
				m['event_key'],
				m['week'],
				m['event_type'],
				m['key'],
				a,
				m['comp_level'],
				m['match_number'],
				m['time'],
				int(t[3:]),
				m['alliances'][a]['score'],
				m['winning_alliance'],
				bd_autocargou,
				bd_telecargou,
				bd_autocargol,
				bd_telecargol,
				b[f"taxiRobot{i+1}"] == 'Yes',
				b[f"endgameRobot{i+1}"],
				b['foulCount'],
				b['techFoulCount']
			]
			g.append(r)
	return g

matches = []
for event in get_all_events_year(2022):
	matches += get_matches(event)

data = list(itertools.chain.from_iterable(map(proc_match, matches)))
df = pd.DataFrame(data, columns=[
	'event', 'week', 'event_type', 'key', 'alliance', 'comp_level',
	'match_number', 'time', 'team', 'score', 'winner',
	'autoCargoUpper', 'teleopCargoUpper', 'autoCargoLower', 'teleopCargoLower', 'taxi', 'endgame',
	'foulCount', 'techFoulcount'
])
df.to_feather("../data/frc2022.feather")
