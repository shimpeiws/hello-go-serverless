.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./hello-world/hello-world
	rm -rf ./s3-triggered/s3-triggered

build:
	GOOS=linux GOARCH=amd64 go build -o hello-world/hello-world ./hello-world
	GOOS=linux GOARCH=amd64 go build -o s3-triggered/s3-triggered ./s3-triggered
