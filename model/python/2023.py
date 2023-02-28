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

def piece_count(d, t):
	counts = {}
	for k, v in d.items():
		counts[k] = len([x for x in d[k] if x == t])
	return counts

def proc_match(m):
	g = []
	if m['score_breakdown'] == None:
		return g
	for a in ['red', 'blue']:
		b = m['score_breakdown'][a]
		auto_counts_cone = piece_count(b['autoCommunity'], 'Cone')
		teleop_counts_cone = piece_count(b['teleopCommunity'], 'Cone')
		auto_counts_cube = piece_count(b['autoCommunity'], 'Cube')
		teleop_counts_cube = piece_count(b['teleopCommunity'], 'Cube')
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
				b[f"mobilityRobot{i+1}"] == 'Yes',
				b[f"autoChargeStationRobot{i+1}"],
				b["autoChargeStationPoints"],
				auto_counts_cone['T'] + auto_counts_cube['T'],
				auto_counts_cone['M'] + auto_counts_cube['M'],
				auto_counts_cone['B'] + auto_counts_cube['B'],
				teleop_counts_cone['T'] + teleop_counts_cube['T'],
				teleop_counts_cone['M'] + teleop_counts_cube['M'],
				teleop_counts_cone['B'] + teleop_counts_cube['B'],
				b[f"endGameChargeStationRobot{i+1}"],
				b["endGameBridgeState"]
			]
			g.append(r)
	return g


#e = get_event("2023isde1")
#matches = get_matches(e)
#print(m)

matches = []
for event_key in ["2023week0", "2023isde1"]:#get_all_events_year(2023):
	event = get_event(event_key)
	matches += get_matches(event)

data = list(itertools.chain.from_iterable(map(proc_match, matches)))
df = pd.DataFrame(data, columns=[
	'event', 'week', 'event_type', 'key', 'alliance', 'comp_level',
	'match_number', 'time', 'team', 'score', 'winner',
	'mobility', 'auto_charge', 'auto_charge_points',
	'auto_countT', 'auto_countM', 'auto_countB',
	'teleop_countT', 'teleop_countM', 'teleop_countB',
	'endgame_charge', 'endGameBridgeState'
])
df.to_feather("../data/raw/frc2023.feather")
