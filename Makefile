.PHONY: test build synth

clean:
	rm -rf build
	rm -rf cdk.out

test:
	bash -c 'diff -u <(echo -n) <(go fmt $(go list ./...))'
	go vet ./...
	go test ./... -v && echo "\nResult=OK" || echo "\nResult=FAIL"

build:
	go run scripts/build/main.go

synth: build
	cdk synth
