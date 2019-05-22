all: test
	go build -buildmode=c-shared -o out_cloudwatch_logs.so .

fast:
	go build out_cloudwatch_logs.go cloudwatch_logs.go

test:
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic

dep:
	dep ensure

clean:
	rm -rf *.so *.h
