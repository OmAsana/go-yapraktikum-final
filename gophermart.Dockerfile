FROM golang:1.18-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build  -o /gophermart cmd/gophermart/main.go

CMD /gophermart
