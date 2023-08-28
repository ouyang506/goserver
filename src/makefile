objects = login gate mysqlproxy redisproxy

.PHONY : all
all: $(objects)

.PHONY : clean
clean: 
	$(foreach object, $(objects), rm -f ../bin/$(object);)

.PHONY : utility
utility:
	@echo "build module utility"
	@cd utility && \
	go mod tidy

.PHONY : common
common: utility
	@echo "build module common"
	@cd common && \
	go mod tidy

.PHONY : mockclient
mockclient: common
	@echo "build mockclient"
	@cd clients/mockclient && \
	go mod tidy	&& \
	go build -o ../../../bin/
	
.PHONY : login
login: common
	@echo "build server login"
	@cd servers/login && \
	go mod tidy	&& \
	go build -o ../../../bin/
	
.PHONY : gate
gate: common
	@echo "build server gate"
	@cd servers/gate && \
	go mod tidy	&& \
	go build -o ../../../bin/
	
.PHONY : mysqlproxy
mysqlproxy: common
	@echo "build server mysqlproxy"
	@cd servers/mysqlproxy && \
	go mod tidy	&& \
	go build -o ../../../bin/
	
.PHONY : redisproxy
redisproxy: common
	@echo "build server redisproxy"
	@cd servers/redisproxy && \
	go mod tidy	&& \
	go build -o ../../../bin/

