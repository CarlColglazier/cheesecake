from cheesecake import create_app, db, socketio

if __name__ == '__main__':
    app = create_app()
    socketio.init_app(app)
    socketio.run(app)
