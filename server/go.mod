module github.com/rawdaGastan/grid3_auto_deployer

go 1.16

require (
	github.com/caitlin615/nist-password-validator v0.0.0-20190321104149-45ab5d3140de
	github.com/golang-jwt/jwt/v4 v4.3.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-envparse v0.1.0
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.29.0 // indirect
	github.com/threefoldtech/grid3-go v0.0.0-20230214163319-124637fb2909
	golang.org/x/crypto v0.6.0
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gorm v1.24.5
)

replace github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.5 => github.com/threefoldtech/go-substrate-rpc-client/v4 v4.0.6-0.20230102154731-7c633b7d3c71
