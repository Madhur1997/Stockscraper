FROM golang:1.14.3

#Install chrome headless shell
RUN apt-get update && apt-get install sudo && \
    wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add - && \
    echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | sudo tee /etc/apt/sources.list.d/google-chrome.list && \
    sudo apt-get update && \ 
    sudo apt-get install google-chrome-stable -y

RUN mkdir -p /src/stockscraper
WORKDIR /src/stockscraper

COPY go.mod go.sum ./

#Install go dependencies.
RUN go mod download

COPY . .

#Build application
RUN make
