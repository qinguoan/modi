
GOROOT := /usr/local/go
GOPATH := $(shell cd ../../; pwd)
GOARCH="amd64"
GOOS="linux"

build:
	go build -o build/tail.linux $(GOPATH)/src/taild/taild.go

rpm:
	mkdir -p build/taild
	cp taild build/tail.linux build/taild/
	tar czvf taild.tar.gz -C build taild
	cp taild.spec build/taild.spec
	./buildrpm.sh -n ./build -s build/taild.spec
	rm -f taild.tar.gz

upload:
	curl -F $$(ls -Art build/RPMS/| tail -n 1)=@build/RPMS/$$(ls -Art build/RPMS/| tail -n 1) download.hy01.wandoujia.com/upload

clean:
	rm -rf build/ taild.tar.gz
