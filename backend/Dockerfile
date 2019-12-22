FROM golang

ADD . /cheesecake
WORKDIR /cheesecake
RUN go get -v ./...
RUN go build
EXPOSE 8080
