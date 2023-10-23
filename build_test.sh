go build -ldflags '-s -w -linkmode "external" -extldflags "-static"' -o redis_back_test *.go
