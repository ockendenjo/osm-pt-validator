.PHONY: clean test build synth deploy

clean:
	rm -rf build
	rm -rf cdk.out

test:
	bash -c 'diff -u <(echo -n) <(go fmt $(go list ./...))'
	go vet ./...
	go test ./... -v && echo "\nResult=OK" || (echo "\nResult=FAIL" && exit 1)

build:
	go run scripts/build/main.go

synth: build
	cdk synth

deploy: build
	cdk deploy --require-approval never
