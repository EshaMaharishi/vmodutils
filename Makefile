
bin/vmodutils: bin *.go cmd/module/*.go *.mod touch/*.go
	go build -o bin/vmodutils cmd/module/cmd.go

test:
	go test

lint:
	gofmt -w -s .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

module: bin/vmodutils
	tar czf module.tar.gz bin/vmodutils meta.json

bin:
	-mkdir bin
