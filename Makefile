.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./hello-world/hello-world
	rm -rf ./s3-triggered/s3-triggered
	rm -rf ./sqs-triggered/sqs-triggered

build:
	GOOS=linux GOARCH=amd64 go build -o hello-world/hello-world ./hello-world
	GOOS=linux GOARCH=amd64 go build -o s3-triggered/s3-triggered ./s3-triggered
	GOOS=linux GOARCH=amd64 go build -o sqs-triggered/sqs-triggered ./sqs-triggered

deploy:
	sam package --profile go-serverless\
		--output-template-file packaged.yaml \
		--s3-bucket hello-go-serverless
	sam deploy --profile go-serverless\
		--template-file packaged.yaml \
		--stack-name hello-go-serverless \
		--capabilities CAPABILITY_IAM
