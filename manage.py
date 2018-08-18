from cheesecake import create_app, db, socketio

if __name__ == '__main__':
    app = create_app()
    socketio.init_app(app)
    #migrate = Migrate(app, db)
    #app.run()
    socketio.run(app)
