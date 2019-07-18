FROM golang

ADD . /go/src/github.com/carlcolglazier/cheesecake
WORKDIR /go/src/github.com/carlcolglazier/cheesecake/
RUN cd backend && go get ./...
RUN cd backend && go install ./...
CMD /go/bin/backend reset
