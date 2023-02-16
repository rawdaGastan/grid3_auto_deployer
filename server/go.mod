module github.com/rawdaGastan/grid3_auto_deployer

go 1.16

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/threefoldtech/grid3-go v0.0.0-20230214163319-124637fb2909
	golang.org/x/crypto v0.6.0
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gorm v1.24.5
)

replace github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.5 => github.com/threefoldtech/go-substrate-rpc-client/v4 v4.0.6-0.20230102154731-7c633b7d3c71
