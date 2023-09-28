FROM debian:11

WORKDIR /data/gameserver

COPY --chmod=755 ./bin/gate /data/gameserver/bin/
COPY ./conf/gate.xml /data/gameserver/conf/

WORKDIR /data/gameserver/bin
ENTRYPOINT ["./gate"]
