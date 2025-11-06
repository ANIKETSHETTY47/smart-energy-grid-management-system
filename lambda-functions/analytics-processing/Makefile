.PHONY: build clean deploy

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	zip function.zip bootstrap

clean:
	rm -f bootstrap function.zip

deploy: clean build
	aws lambda update-function-code \
		--function-name analytics-processing \
		--zip-file fileb://function.zip