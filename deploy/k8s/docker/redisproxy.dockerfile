FROM debian:11

WORKDIR /data/gameserver

COPY --chmod=755 ./bin/redisproxy /data/gameserver/bin/
COPY ./conf/redisproxy.xml /data/gameserver/conf/

WORKDIR /data/gameserver/bin
ENTRYPOINT ["./redisproxy"]
