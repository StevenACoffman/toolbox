.DEFAULT_GOAL := easy
.PHONY: install clean all easy

bin/eureka-lookup:
	go build -o bin/eureka-lookup cmd/eureka-lookup/lookup.go

bin/generate-tls-cert:
	go build -o bin/generate-tls-cert cmd/generate-tls-cert/generate.go

bin/gh-make-pull:
	go build -o bin/gh-make-pull cmd/gh-make-pull/gh-make-pull.go

bin/gh-make-token:
	go build -o bin/gh-make-token cmd/gh-make-token/gh-make-token.go

bin/j2m:
	go build -o bin/j2m cmd/j2m/j2m.go

bin/jp:
	go build -o bin/jp cmd/jira-pull/jp.go

bin/wti:
	go build -o bin/wti cmd/wti/wti.go

all: bin/eureka-lookup bin/generate-tls-cert bin/gh-make-pull bin/gh-make-token bin/j2m bin/jp bin/wti

install: bin/eureka-lookup bin/generate-tls-cert bin/gh-make-pull bin/gh-make-token bin/j2m bin/jp bin/wti
	cp bin/* ~/bin

clean:
	rm -f bin/*

easy: install clean