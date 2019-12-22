FROM golang

RUN mkdir /cheesecake
WORKDIR /cheesecake
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build
EXPOSE 8080
