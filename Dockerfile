FROM golang:1.16

MAINTAINER Chris Watson (chris@dreamcove.com)

USER root

COPY go.mod ./src
COPY main.go ./src

RUN apt-get update
RUN apt-get -y install unzip

RUN curl https://stockfishchess.org/files/stockfish_14_linux_x64.zip -o stockfish_14_linux_x64.zip
RUN unzip stockfish_14_linux_x64.zip stock*64
RUN cp stockfish_14_linux_x64/*_x64 ./stockfish_14_x64
RUN rm -Rf stockfish_14_linux_x64*
RUN chmod a+x ./stockfish_14_x64

ENV STOCKFISH_PATH=./stockfish_14_x64

EXPOSE 8081

RUN cd src && go mod download github.com/freeeve/uci && go build
RUN mv src/stockfish-server .
RUN mkdir data

HEALTHCHECK --interval=1m --timeout=3s CMD curl -f "http://localhost:8081/move?game=test&fen=rnbqkbnr%2Fpppppppp%2F8%2F8%2F8%2F8%2FPPPPPPPP%2FRNBQKBNR%20w%20KQkq%20-%200%201" | grep "Results"

ENTRYPOINT ./stockfish-server
