FROM golang

ADD . /go/src/github.com/carlcolglazier/cheesecake
WORKDIR /go/src/github.com/carlcolglazier/cheesecake/
RUN cd backend && go get -v ./...
RUN cd backend && go install ./...
ENTRYPOINT ["/go/bin/backend", "server"]
EXPOSE 8080
