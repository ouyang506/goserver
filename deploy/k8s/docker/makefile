objects = login gate player mysqlproxy redisproxy

.PHONY : all
all: $(objects)

.PHONY : login
login:
	docker build -f login.dockerfile -t login ../../../

.PHONY : gate
gate:
	docker build -f gate.dockerfile -t gate ../../../

.PHONY : player
player:
	docker build -f player.dockerfile -t player ../../../

.PHONY : mysqlproxy
mysqlproxy:
	docker build -f mysqlproxy.dockerfile -t mysqlproxy ../../../

.PHONY : redisproxy
redisproxy:
	docker build -f redisproxy.dockerfile -t redisproxy ../../../

.PHONY : run
run:
	@for s in $(objects); do \
		echo run server $$s; \
		docker run -itd --name $$s $$s ;\
	done

.PHONY : restart
restart: stop
	@for s in $(objects); do \
		echo run server $$s; \
		docker run -itd --name $$s $$s ;\
	done
	
.PHONY : stop
stop:
	@for s in $(objects); do \
		echo stop server $$s; \
		docker rm -f $$s;\
	done