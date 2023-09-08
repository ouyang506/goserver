objects = login gate mysqlproxy redisproxy

# make debug=true
ifeq ($(debug),true)
	build_flags = -gcflags "-l -N"
endif

.PHONY : all
all: $(objects)

.PHONY : clean
clean: 
	$(foreach object, $(objects), rm -f ../bin/$(object);)

.PHONY : proto
proto:
	@cd framework/proto/idl &&\
	./proto_gen.bat

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

.PHONY : framework
framework: utility  common
	@echo "build module framework"
	@cd framework && \
	go mod tidy

.PHONY : mockclient
mockclient: utility  common  framework
	@echo "build mockclient"
	@cd clients/mockclient && \
	go mod tidy	&& \
	go build $(build_flags) -o ../../../bin/
	
.PHONY : login
login: utility  common  framework
	@echo "build server login"
	@cd servers/login && \
	go mod tidy	&& \
	go build $(build_flags) -o ../../../bin/
	
.PHONY : gate
gate: utility  common  framework
	@echo "build server gate"
	@cd servers/gate && \
	go mod tidy	&& \
	go build $(build_flags) -o ../../../bin/
	
.PHONY : mysqlproxy
mysqlproxy: utility  common  framework
	@echo "build server mysqlproxy"
	@cd servers/mysqlproxy && \
	go mod tidy	&& \
	go build $(build_flags) -o ../../../bin/
	
.PHONY : redisproxy
redisproxy: utility  common  framework
	@echo "build server redisproxy"
	@cd servers/redisproxy && \
	go mod tidy	&& \
	go build $(build_flags) -o ../../../bin/

