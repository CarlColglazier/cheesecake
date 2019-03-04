from .. import db

district_teams = db.Table('district_teams',
                          db.Column('position', db.Integer,
                                    primary_key=True),
                          db.Column('district_key', db.String(10),
                                    db.ForeignKey('district.key')),
                          db.Column('team_key', db.String(8),
                                    db.ForeignKey('team.key'))
)

event_teams = db.Table('event_teams',
                       db.Column('position', db.Integer,
                                 primary_key=True),
                       db.Column('event_key', db.String(10),
                                 db.ForeignKey('event.key')),
                       db.Column('team_key', db.String(8),
                                 db.ForeignKey('team.key'))
)

alliance_teams = db.Table('alliance_teams',
                          db.Column('position', db.Integer,
                                    primary_key=True,
                                    autoincrement=True),
                          db.Column('alliance_id', db.String(25),
                                    db.ForeignKey('alliance.key'),
                                    primary_key=True),
                          db.Column('team_key', db.String(8),
                                    db.ForeignKey('team.key'),
                                    primary_key=True)
)
