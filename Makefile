GOCMD=go
GOBUILD=$(GOCMD) build
EXEC=cheesecake-backend
all: build
build:
	$(GOBUILD) -o $(EXEC) ./backend/
clean:
	$(GOCMD) clean
	rm -f $(EXEC)

dev:
	@docker-compose \
		-f docker-compose.yml \
		-f docker-compose.dev.yml \
		up -d --build

prod:
	@docker-compose down && \
		docker-compose \
			-f docker-compose.yml \
			-f docker-compose.prod.yml \
		up -d --build

test:
	docker-compose build && \
	docker-compose up -d db && \
	docker-compose \
		-f docker-compose.test.yml \
		up backend && \
	docker-compose \
		-f docker-compose.test.yml \
		up frontend && \
	docker-compose down
