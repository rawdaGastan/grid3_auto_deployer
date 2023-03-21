module github.com/rawdaGastan/cloud4students

go 1.18

require (
	github.com/caitlin615/nist-password-validator v0.0.0-20190321104149-45ab5d3140de
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/magiconair/properties v1.8.4
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.28.0
	github.com/spf13/cobra v1.6.1
	github.com/threefoldtech/grid3-go v0.0.0-20230313121415-1da999636079
	github.com/threefoldtech/grid_proxy_server v1.6.12
	github.com/threefoldtech/zos v0.5.6-0.20230224113017-e887a6ca3fc5
	golang.org/x/crypto v0.7.0
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gorm v1.24.6
)

require (
	github.com/ChainSafe/go-schnorrkel v1.0.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.5 // indirect
	github.com/cosmos/go-bip39 v1.0.0 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/base58 v1.0.3 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/ethereum/go-ethereum v1.10.17 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/gtank/merlin v0.1.1 // indirect
	github.com/gtank/ristretto255 v0.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jbenet/go-base58 v0.0.0-20150317085156-6237cf65f3a6 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20210601165009-122bf33a46e0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pierrec/xxHash v0.1.5 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/threefoldtech/rmb-sdk-go v1.0.1-0.20230308130815-83a645307186 // indirect
	github.com/threefoldtech/substrate-client v0.1.2 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/vedhavyas/go-subkey v1.0.3 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200609130330-bd2cb7843e1b // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

replace github.com/centrifuge/go-substrate-rpc-client/v4 v4.0.5 => github.com/threefoldtech/go-substrate-rpc-client/v4 v4.0.6-0.20230102154731-7c633b7d3c71
