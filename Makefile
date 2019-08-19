.DEFAULT_GOAL := easy
.PHONY: install clean all easy

bin/eureka-lookup:
	go build -o bin/eureka-lookup cmd/eureka-lookup/lookup.go

bin/generate-tls-cert:
	go build -o bin/generate-tls-cert cmd/generate-tls-cert/generate.go

bin/github-make-token:
    go build -o bin/github-make-token cmd/github-make-token/github-make-token.go

bin/wti:
	go build -o bin/wti cmd/wti/wti.go

all: bin/eureka-lookup bin/generate-tls-cert bin/github-make-token bin/wti

install: bin/eureka-lookup bin/generate-tls-cert bin/github-make-token bin/wti
	cp bin/* ~/bin

clean:
	rm -f bin/*

easy: install clean