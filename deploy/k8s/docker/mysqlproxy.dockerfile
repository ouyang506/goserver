FROM debian:11

WORKDIR /data/gameserver

COPY --chmod=755 ./bin/mysqlproxy /data/gameserver/bin/
COPY ./conf/mysqlproxy.xml /data/gameserver/conf/

WORKDIR /data/gameserver/bin
ENTRYPOINT ["./mysqlproxy"]
