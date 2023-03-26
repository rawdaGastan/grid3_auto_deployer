module github.com/rawdaGastan/cloud4students

go 1.16

require (
	github.com/caitlin615/nist-password-validator v0.0.0-20190321104149-45ab5d3140de
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.28.0
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.12.0+incompatible
	github.com/spf13/cobra v1.6.1
	github.com/threefoldtech/grid3-go v0.0.0-20230313121415-1da999636079
	github.com/threefoldtech/grid_proxy_server v1.6.12
	github.com/threefoldtech/zos v0.5.6-0.20230224113017-e887a6ca3fc5
	golang.org/x/crypto v0.7.0
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gorm v1.24.6
)

replace github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.5 => github.com/threefoldtech/go-substrate-rpc-client/v4 v4.0.6-0.20230102154731-7c633b7d3c71
