FROM debian:11

WORKDIR /data/gameserver

COPY --chmod=755 ./bin/player /data/gameserver/bin/
COPY ./conf/player.xml /data/gameserver/conf/

WORKDIR /data/gameserver/bin
ENTRYPOINT ["./player"]
