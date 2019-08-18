.DEFAULT_GOAL := easy
.PHONY: install clean all easy

bin/eureka-lookup:
	go build -o bin/eureka-lookup cmd/eureka-lookup/lookup.go

bin/generate-tls-cert:
	go build -o bin/generate-tls-cert cmd/generate-tls-cert/generate.go

all: bin/eureka-lookup bin/generate-tls-cert

install: bin/eureka-lookup bin/generate-tls-cert
	cp bin/* ~/bin

clean:
	rm -f bin/*

easy: install clean