FROM gcc:latest AS builder

RUN apt-get update && apt-get install -y build-essential && apt-get install -y libstdc++6

COPY master.cpp /app/master.cpp
COPY server.cpp /app/server.cpp

WORKDIR /app
RUN g++ -static-libstdc++ master.cpp -o master -lpthread
RUN g++ -static-libstdc++ server.cpp -o server -lpthread

FROM debian:latest

COPY --from=builder /app/master /app/master
COPY --from=builder /app/server /app/server

EXPOSE 9000

CMD ["/bin/bash"]
