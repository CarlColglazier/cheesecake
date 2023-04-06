import requests
import itertools
import pandas as pd
from dotenv import load_dotenv
import os
from pathlib import Path
import json
import argparse

load_dotenv()

API = "https://www.thebluealliance.com/api/v3"
header_file_loc = "../data/raw/headers.json"

for directory in ["../data/", "../data/raw/", "../data/cache/", "../data/schedules/", "../data/breakdowns/"]:
	if not os.path.exists(directory):
		os.makedirs(directory)

def get_headers():
	p = Path(header_file_loc)
	if p.is_file():
		with open(p, "r") as f:
			return json.load(f)
	return {}

header_cache = get_headers()

def run_query(u):
	header = {
		"X-TBA-Auth-Key":os.environ.get("KEY"),
	}
	uri = f"{API}/{u}"
	if uri in header_cache:
		header["If-None-Match"] = header_cache[uri]
	r = requests.get(uri, headers=header)
	p = Path(f"../data/cache/{u.replace('/', '_')}.json")
	if r.status_code == 200:
		header_cache[uri] = r.headers["ETag"]
		j = r.json()
		if p.is_file():
			with open(p, "r") as f:
				fj = json.load(f)
				if len(json.dumps(fj, sort_keys=True)) == len(json.dumps(j, sort_keys=True)):
					#print(f"Found equal match for {u}")
					return (j, False)
		print(f"UPDATING {uri}")
		with open(p, "w") as f:
			json.dump(j, f, sort_keys=True)
		return (j, True)
	if p.is_file():
		js = json.load(open(p, "r"))
		return (js, False)
	else:
		return (None, False)

"""
Get keys for events in the year
"""
def get_all_events_year(year):
	url = f"events/{year}"
	j, _ = run_query(url)
	j = filter(lambda x: x['event_type'] <= 6, j)
	return list(j)
	#return list(map(lambda x: x['key'], j))

def to_event_keys(j):
	return list(map(lambda x: x['key'], j))

def get_event(key):
	url = f"event/{key}"
	return run_query(url)[0]

def get_matches(event):
	event_key = event['key']
	url = f"event/{event_key}/matches"
	j, should_run = run_query(url)
	if j is None or should_run is False:
		return (None, False)
	for m in j:
		if event['week'] is None:
			m['week'] = 10
		else:
			m['week'] = event['week']
		m['event_type'] = event['event_type']
	return j, should_run

def piece_count(d, t):
	counts = {}
	for k, v in d.items():
		counts[k] = len([x for x in d[k] if x == t])
	return counts

def proc_schedule(m):
	return [
		m['key'],
		m['comp_level'],
		m['match_number'],
		m['time'],
		list(map(lambda t: int(t[3:]), m['alliances']['red']['team_keys'])),
		list(map(lambda t: int(t[3:]), m['alliances']['blue']['team_keys'])),
		m['alliances']['red']['score'],
		m['alliances']['blue']['score']
	]

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
				b["endGameBridgeState"],
				b["activationBonusAchieved"],
				b["sustainabilityBonusAchieved"]
			]
			g.append(r)
	return g


def run(week=None):
	events = get_all_events_year(2023)
	with open("../../files/api/events.json", "w") as f:
		json.dump(events, f, sort_keys=True)
	if week is not None:
		events = list(filter(lambda x: x['week'] == week, events))
	event_keys = to_event_keys(events)
	for event_key in event_keys:
		event = get_event(event_key)
		matches, should_run = get_matches(event)
		if should_run is False:
			continue
		if matches is None:
			continue
		if len(matches) == 0:
			continue
		m = list(map(proc_schedule, matches))
		df = pd.DataFrame(m, columns=[
			'key', 'comp_level', 'match_number', 'time', 'red_teams', 'blue_teams', 'red_score', 'blue_score'
		])
		if df.shape[0] > 0:
			df.to_feather(f"../data/schedules/{event_key}.feather")
		data = list(itertools.chain.from_iterable(map(proc_match, matches)))
		df = pd.DataFrame(data, columns=[
			'event', 'week', 'event_type', 'key', 'alliance', 'comp_level',
			'match_number', 'time', 'team', 'score', 'winner',
			'mobility', 'auto_charge', 'auto_charge_points',
			'auto_countT', 'auto_countM', 'auto_countB',
			'teleop_countT', 'teleop_countM', 'teleop_countB',
			'endgame_charge', 'endGameBridgeState',
			'activation', 'sustainability'
		])
		if df.shape[0] > 0:
			df.to_feather(f"../data/breakdowns/{event_key}.feather")

	with open(Path(header_file_loc), "w") as f:
		json.dump(header_cache, f)

def get_args():
	parser = argparse.ArgumentParser()
	parser.add_argument("--week", type=int, default=None)
	return parser.parse_args()

if __name__ == "__main__":
	args = get_args()
	run(week=args.week)