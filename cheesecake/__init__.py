from flask import Flask
from flask_caching import Cache
from flask_cors import CORS
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate
from secret import *
import tbapy
import os

from config import DevelopmentConfig, ProductionConfig

# Flask extensions
db = SQLAlchemy(session_options={
    "autoflush": False,
    "autocommit": False,
    "expire_on_commit": False
})

cache = Cache(config={'CACHE_TYPE': 'simple'})

# Other
tba = tbapy.TBA(TBA_KEY)

# Import models
from . import models

def create_app():
    print("Creating app...")
    app = Flask(__name__)
    if "FLASK_ENV" in os.environ and os.environ["FLASK_ENV"] == "development":
        print("Running in development mode.")
        app.config.from_object(DevelopmentConfig)
    else:
        app.config.from_object(ProductionConfig)
    print("Initializing database.")
    db.init_app(app)
    cors = CORS(app, resources={r"/*":{"origins":"*"}})
    print("Initializing cache.")
    cache.init_app(app)
    Migrate(app, db, compare_type=True)

    from .api import api as api_blueprint
    app.register_blueprint(api_blueprint, url_prefix='/api')
    return app
