all: clean auth_service currency_service bank_service

generate: generate_auth

generate_auth:
	cd protos && protoc -I proto proto/auth/auth.proto --go_out=./gen/go --go_opt=paths=source_relative --go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative

auth_service: clean
	go build -o $@ cmd/auth/main.go

currency_service: clean
	go build -o $@ cmd/currency/main.go 

mail_service: clean
	go build -o $@ cmd/mail/main.go 

bank_service: clean
	go build -o $@ cmd/bank/main.go 

swag:
	swag init -d internal/delivery/http/bank/resource/auth,internal/delivery/http/bank/resource/bank,internal/delivery/http/bank/resource/currency,internal/api/response -g ../../../../../../cmd/bank/main.go -o docs

clean:
	rm -rf auth_service currency_service bank_service mail_service