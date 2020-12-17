FROM golang:1.14.3

RUN mkdir -p /src/stockscraper
WORKDIR /src/stockscraper

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make
