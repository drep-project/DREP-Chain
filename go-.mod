module github.com/drep-project/drep-chain

go 1.13

replace (
	github.com/drep-project/binary => /root/go/src/github.com/drep-project/binary
	github.com/drep-project/dlog => /root/go/src/github.com/drep-project/dlog
	github.com/drep-project/rpc => /root/go/src/github.com/drep-project/rpc

)

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20190914173115-096a252f7b2f
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/allegro/bigcache v1.2.1
	github.com/aristanetworks/goarista v0.0.0-20191023202215-f096da5361bb
	github.com/asaskevich/EventBus v0.0.0-20180315140547-d46933a94f05
	github.com/astaxie/beego v1.12.0 // indirect
	github.com/btcsuite/btcd v0.20.0-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cosmos/cosmos-sdk v0.37.3
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/docker/docker v1.13.1
	github.com/drep-project/binary v0.0.0-00010101000000-000000000000
	github.com/drep-project/dlog v0.0.0-00010101000000-000000000000
	github.com/drep-project/rpc v0.0.0-00010101000000-000000000000
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/elastic/gosigar v0.10.5 // indirect
	github.com/ethereum/go-ethereum v1.9.6
	github.com/fatih/color v1.7.0
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20190607065134-2772fd86a8ff // indirect
	github.com/golang/snappy v0.0.1
	github.com/huin/goupnp v1.0.0
	github.com/jackpal/go-nat-pmp v1.0.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/karalabe/usb v0.0.0-20190919080040-51dc0efba356 // indirect
	github.com/mattn/go-colorable v0.1.4
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/olekukonko/tablewriter v0.0.1 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterh/liner v1.1.0
	github.com/pingcap/errors v0.11.4
	github.com/pkg/errors v0.8.1
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/rubblelabs/ripple v0.0.0-20190714134121-6dd7d15dd060
	github.com/sasaxie/go-client-api v0.0.0-20190820063117-f0587df4b72e
	github.com/shengdoushi/base58 v1.0.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644
	github.com/sirupsen/logrus v1.4.2
	github.com/status-im/keycard-go v0.0.0-20190424133014-d95853db0f48 // indirect
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tidwall/pretty v1.0.0 // indirect
	github.com/tyler-smith/go-bip39 v1.0.2 // indirect
	github.com/vishalkuo/bimap v0.0.0-20180703190407-09cff2814645
	github.com/wsddn/go-ecdh v0.0.0-20161211032359-48726bab9208 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	go.mongodb.org/mongo-driver v1.1.2
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/net v0.0.0-20191028085509-fe3aa8a45271
	golang.org/x/sys v0.0.0-20191028164358-195ce5e7f934
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
)
