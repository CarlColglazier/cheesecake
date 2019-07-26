FROM golang

ADD . /cheesecake
WORKDIR /cheesecake
RUN go get -v ./...
RUN go build
ENTRYPOINT ["backend", "server"]
EXPOSE 8080
