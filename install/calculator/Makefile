GOROOT := /usr/local/go
GOPATH := $(shell cd ../../;pwd)

build:
	mkdir -p build/calculator/bin
	GOPATH=$(GOPATH) GOARCH="amd64" GOOS="linux" go build -o build/calculator/bin/calculator $(GOPATH)/src/calculator/calculator.go

rpm:
	mkdir -p build/calculator/conf
	mkdir -p build/calculator/log
	mkdir -p build/calculator/run
	cp calculatord build/calculator/run
	cp ../../conf/calculator.yaml build/calculator/conf
	tar czvf calculator.tar.gz -C build calculator
	./buildrpm.sh -n . -s calculator.spec
	rm -f calculator.tar.gz

upload:
	curl -F $$(ls -Art RPMS/| tail -n 1)=@RPMS/$$(ls -Art RPMS/| tail -n 1) download.hy01.wandoujia.com/upload

clean:
	rm -rf build/ calculator.tar.gz
