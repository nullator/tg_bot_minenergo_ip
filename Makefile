BINARY_NAME=minenergo

build:
	GOARCH=amd64 GOOS=darwin go build -o ./output/${BINARY_NAME}-mac cmd/main.go
	GOARCH=amd64 GOOS=linux go build -o ./output/${BINARY_NAME}-linux cmd/main.go
	go build -o ./output/${BINARY_NAME} cmd/main.go

run: build
	./output/${BINARY_NAME}

clean:
	go clean
	rm ./output/${BINARY_NAME}-mac
	rm ./output/${BINARY_NAME}-linux
	rm ./output/${BINARY_NAME}

mock:
	mockery --all

deploy: build
	./deploy.sh
