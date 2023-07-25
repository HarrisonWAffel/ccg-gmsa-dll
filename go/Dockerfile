FROM golang:1.20

# This file is just for local testing
# TODO: provide some args to disable windows specific things

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

EXPOSE 8080

# Run
CMD ["/main"]