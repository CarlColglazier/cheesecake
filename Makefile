GOCMD=go
GOBUILD=$(GOCMD) build
EXEC=cheesecake-backend
all: build
build:
	$(GOBUILD) -o $(EXEC) ./backend/
clean:
	$(GOCMD) clean
	rm -f $(EXEC)
