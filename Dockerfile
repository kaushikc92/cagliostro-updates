FROM ubuntu:20.04

RUN apt-get update -y
RUN apt-get install -y vim wget git build-essential

COPY stockfish /usr/bin/
COPY cagliostro-updates /usr/bin/

#CMD ["tail", "-f", "/dev/null"]
CMD ["cagliostro-updates"]
