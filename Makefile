.PHONY: clean test build synth deploy cfn

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
	cdk deploy --require-approval never --method direct

cfn:
	aws cloudformation deploy --template-file setup.yml --stack-name "OSMPTSetup" --capabilities "CAPABILITY_NAMED_IAM" --region=eu-west-1
