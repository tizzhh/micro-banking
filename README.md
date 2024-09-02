# Micro-banking

A simple example of small subset of banking operations in a microservice environment. User management, balance management and currency acquirement.

## Requirements

- Docker & Docker compose
- config in the format of example.yaml in config/ directory. If env CONFIG_PATH is not set, the service expects config/local.yaml as config.
- .env in infra/ directory with the following fields: 
    - POSTGRES_USER=
    - POSTGRES_PASSWORD=
    - POSTGRES_DB=
- for goose migrations another .env is required:
    ```
    goose.env
    export GOOSE_DRIVER='postgres'
    export GOOSE_DBSTRING='postgres://user:password@host:port/db_name?sslmode=disable'
    export GOOSE_MIGRATION_DIR='./migrations'
    ```
    then `source goose.env` and goose up

## Used packages / tools / stack

- Docker, Docker compose.
- gRPC, REST.
- Postgresql, Kafka, Redis.
- Emailing service for User notification.
- Testing with [Mockery](https://github.com/vektra/mockery).
- Healthcheck and CRUD API implementations with OpenAPI specifications and JWT authentication.
- The usage of [Goose](https://github.com/pressly/goose) for the database migrations and [GORM](https://gorm.io/) as the database ORM.
- The usage of slog as the centralized logger.
- The usage of [Validator.v10](https://github.com/go-playground/validator) as the form validator.
- The usage of [Protovalidate](https://github.com/bufbuild/protovalidate) as gRPC message validator.
- Documentation with [Swaggo/swag](https://github.com/swaggo/swag).


## Endpoints

| Name        | HTTP Method | Route          |
|-------------|-------------|----------------|
| Health      | GET         | /v1/liveness   |
| Register User| POST | /v1/auth/register |
| Login User | POST | /v1/auth/login |
| Change Password | PUT | /v1/auth/change-password |
| Unregister | DELETE | /v1/auth/unregister |
| Get User | GET | /v1/auth/user |
| Check wallet | GET | /v1/bank/my-wallet |
| Deposit | POST | /v1/bank/deposit |
| Withdraw | POST | /v1/bank/withdraw |
| Buy currency | POST | /v1/currency/buy |
| Sell currency | POST | /v1/currency/sell |

Swag documentation included:

💡 [swaggo/swag](https://github.com/swaggo/swag) : `make swag`

##  Database design

#### users

| Column Name    | Datatype  | Not Null | Primary Key |
|----------------|-----------|----------|-------------|
| id             | BIGINT      | ✅        | ✅           |
| email          | VARCHAR      | ✅        |             |
| pass_hash         | VARCHAR      | ✅        |             |
| first_name      | VARCHAR      |  ✅       |             |
| last_name    | VARCHAR      |  ✅       |             |
| balance     | BIGINT | ✅        |             |
| age     | SMALLINT | ✅        |             |

#### currencies

| Column Name    | Datatype  | Not Null | Primary Key |
|----------------|-----------|----------|-------------|
| id             | BIGINT      | ✅        | ✅           |
| code          | CHAR      | ✅        |             |

#### user_wallets

| Column Name    | Datatype  | Not Null | Primary Key |
|----------------|-----------|----------|-------------|
| id             | BIGINT      | ✅        | ✅           |
| user_id          | Foreign key      | ✅        |             |
| currency_id         | Foreign key      | ✅        |             |
| balance | BIGINT      | ✅        |             |



## Usage 💡

- clone the repository
- `cd infra`
- `docker build -f prod.Dockerfile . -t myapp_app`

## 📁 Project structure

```shell
micro-bank
├── Makefile
├── README.md
├── buf.gen.yaml
├── buf.lock
├── buf.yaml
├── cmd
│   ├── auth
│   │   └── main.go
│   ├── bank
│   │   └── main.go
│   ├── currency
│   │   └── main.go
│   └── mail
│       └── main.go
├── config
│   ├── example.yaml
│   ├── local.yaml
│   └── prod.yaml
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── env
│   └── goose.env
├── gen
│   └── go
│       └── protos
│           └── proto
│               ├── auth
│               │   ├── auth.pb.go
│               │   └── auth_grpc.pb.go
│               └── currency
│                   ├── currency.pb.go
│                   └── currency_grpc.pb.go
├── go.mod
├── go.sum
├── infra
│   ├── Dockerfile-auth
│   ├── Dockerfile-bank
│   ├── Dockerfile-currency
│   ├── Dockerfile-mail
│   ├── docker-compose.yaml
│   └── entrypoint.sh
├── internal
│   ├── api
│   │   ├── permissions
│   │   │   └── permissions.go
│   │   ├── response
│   │   │   └── response.go
│   │   └── validate
│   │       └── validate.go
│   ├── app
│   │   ├── auth
│   │   │   ├── app.go
│   │   │   └── grpc
│   │   │       └── app.go
│   │   ├── bank
│   │   │   ├── app.go
│   │   │   └── http
│   │   │       └── app.go
│   │   └── currency
│   │       ├── app.go
│   │       └── grpc
│   │           └── app.go
│   ├── clients
│   │   ├── auth
│   │   │   └── grpc
│   │   │       ├── client.go
│   │   │       └── methods.go
│   │   ├── currency
│   │   │   └── grpc
│   │   │       ├── client.go
│   │   │       └── methods.go
│   │   └── kafka
│   │       ├── consumer
│   │       │   └── kafka.go
│   │       ├── model.go
│   │       └── producer
│   │           └── kafka.go
│   ├── config
│   │   └── config.go
│   ├── delivery
│   │   ├── grpc
│   │   │   ├── auth
│   │   │   │   └── server.go
│   │   │   └── currency
│   │   │       └── server.go
│   │   └── http
│   │       └── bank
│   │           ├── common
│   │           │   └── common.go
│   │           ├── resource
│   │           │   ├── auth
│   │           │   │   ├── handler.go
│   │           │   │   ├── mocks
│   │           │   │   │   └── AuthClient.go
│   │           │   │   └── resource.go
│   │           │   ├── bank
│   │           │   │   ├── handler.go
│   │           │   │   ├── mocks
│   │           │   │   │   └── Balancer.go
│   │           │   │   └── resource.go
│   │           │   └── currency
│   │           │       ├── handler.go
│   │           │       ├── mocks
│   │           │       │   └── CurrencyClient.go
│   │           │       └── resource.go
│   │           └── router
│   │               ├── middleware
│   │               │   ├── auth
│   │               │   │   └── auth.go
│   │               │   └── logger
│   │               │       └── logger.go
│   │               └── router.go
│   ├── domain
│   │   ├── auth
│   │   │   └── models
│   │   │       └── user.go
│   │   └── currency
│   │       └── models
│   │           └── currency.go
│   ├── services
│   │   ├── auth
│   │   │   ├── auth.go
│   │   │   └── errors
│   │   │       └── errors.go
│   │   ├── bank
│   │   │   ├── bank.go
│   │   │   └── errors
│   │   │       └── errors.go
│   │   └── currency
│   │       ├── currency.go
│   │       └── errors
│   │           └── errors.go
│   └── storage
│       ├── errors.go
│       ├── postgres
│       │   └── postgres.go
│       └── redis
│           └── redis.go
├── migrations
│   ├── 00001_create_users.sql
│   └── 00002_create_currency.sql
├── pkg
│   ├── currencyapi
│   │   ├── currencyapi.go
│   │   └── domain
│   │       └── http
│   │           └── currencyapi.go
│   ├── jwt
│   │   └── jwt.go
│   ├── logger
│   │   └── sl
│   │       └── sl.go
│   └── mail
│       └── app.go
├── protos
│   └── proto
│       ├── auth
│       │   └── auth.proto
│       └── currency
│           └── currency.proto
└── tests
    ├── auth_http_handlers_test.go
    ├── auth_service_test.go
    ├── bank_http_handlers_test.go
    ├── currency_http_handlers_test.go
    ├── currency_service_test.go
    ├── migrations
    │   └── 00003_insert_test_user.sql
    └── suite
        ├── auth
        │   └── suite.go
        └── currency
            └── suite.go
```
