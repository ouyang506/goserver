FROM debian:11

WORKDIR /data/gameserver

COPY --chmod=755 ./bin/login /data/gameserver/bin/
COPY ./conf/login.xml /data/gameserver/conf/

WORKDIR /data/gameserver/bin
ENTRYPOINT ["./login"]
