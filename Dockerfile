FROM golang:1.13

MAINTAINER Chris Watson (chris@dreamcove.com)

USER root

COPY go.mod ./src
COPY main.go ./src

RUN apt-get update
RUN apt-get -y install unzip

RUN curl https://stockfishchess.org/files/stockfish-11-linux.zip -o stockfish-11-linux.zip
RUN unzip stockfish-11-linux.zip stock*64
RUN cp stockfish-11-linux/Linux/*_x64 ./stockfish_x64
RUN rm -Rf stockfish-11-linux*
RUN chmod a+x ./stockfish_x64

ENV STOCKFISH_PATH=./stockfish_x64

EXPOSE 8081

RUN cd src && go build
RUN mv src/stockfish-server .
RUN mkdir data

HEALTHCHECK --interval=1m --timeout=3s CMD curl -f "http://localhost:8081/move?game=test&fen=rnbqkbnr%2Fpppppppp%2F8%2F8%2F8%2F8%2FPPPPPPPP%2FRNBQKBNR%20w%20KQkq%20-%200%201" | grep "Results"

ENTRYPOINT ./stockfish-server
