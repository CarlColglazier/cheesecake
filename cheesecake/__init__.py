from flask import Flask
from flask_caching import Cache
from flask_cors import CORS
from flask_sqlalchemy import SQLAlchemy
from flask_socketio import SocketIO, emit
from flask_migrate import Migrate
from secret import *
import tbapy
import os

from config import DevelopmentConfig

# Flask extensions
db = SQLAlchemy(session_options={
    "autoflush": False,
    "autocommit": False,
    "expire_on_commit": False
})
socketio = SocketIO()
cache = Cache(config={'CACHE_TYPE': 'simple'})

# Other
tba = tbapy.TBA(TBA_KEY)

# Import models
from . import models

# Import socket events.
#from . import events

def create_app():
    app = Flask(__name__)
    
    app.config.from_object(DevelopmentConfig)

    db.init_app(app)
    cors = CORS(app, resources={r"/*":{"origins":"*"}})
    cache.init_app(app)
    Migrate(app, db, compare_type=True)

    from .api import api as api_blueprint
    app.register_blueprint(api_blueprint, url_prefix='/api')
    return app
