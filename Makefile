.PHONY: fmt tidy test cover

fmt:
	goimports -w .

tidy:
	go mod tidy

test:
	go test -v -run "$(RUN)" -count 1 ./...

cover:
	go test -v -run "$(RUN)" -count 1 -coverprofile cp.out ./...
	go tool cover -html=cp.out
	rm cp.out
