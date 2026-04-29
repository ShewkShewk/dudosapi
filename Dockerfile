FROM golang:1.25

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /dudosapi

EXPOSE 8080

CMD ["/dudosapi"]