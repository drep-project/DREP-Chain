module github.com/drep-project/drep-chain

replace (
	cloud.google.com/go => github.com/GoogleCloudPlatform/google-cloud-go v0.47.0
	cloud.google.com/go/bigquery => github.com/googleapis/google-cloud-go/bigquery v1.2.0

	cloud.google.com/go/datastore => github.com/googleapis/google-cloud-go/datastore v1.0.0

	cloud.google.com/go/pubsub => github.com/googleapis/google-cloud-go/pubsub v1.0.1
	cloud.google.com/go/storage => github.com/googleapis/google-cloud-go/storage v1.3.0
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20191111213947-16651526fdb4
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190731235908-ec7cb31e5a56
	golang.org/x/image => github.com/golang/image v0.0.0-20191009234506-e7c1f5e7dbb8
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190930215403-16217165b5de
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20191031020345-0945064e013a
	golang.org/x/mod => github.com/golang/mod v0.1.0
	golang.org/x/net => github.com/golang/net v0.0.0-20191109021931-daa7c04131f5
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys => github.com/golang/sys v0.0.0-20191110163157-d32e6e3b99c4
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/time => github.com/golang/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools => github.com/golang/tools v0.0.0-20191112005509-a3f652f18032
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20191011141410-1b5146add898
	google.golang.org/api => github.com/googleapis/google-api-go-client v0.8.0
	google.golang.org/appengine => github.com/golang/appengine v1.6.5
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20191108220845-16a3f7862a1a
	google.golang.org/grpc => github.com/grpc/grpc-go v1.25.1
	gopkg.in/urfave/cli.v1 => github.com/urfave/cli v1.21.0
)

go 1.12

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20191102041813-8372a58e6b17
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/allegro/bigcache v1.2.1
	github.com/aristanetworks/goarista v0.0.0-20191023202215-f096da5361bb
	github.com/asaskevich/EventBus v0.0.0-20180315140547-d46933a94f05
	github.com/astaxie/beego v1.12.0 // indirect
	github.com/btcsuite/btcd v0.20.0-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cosmos/cosmos-sdk v0.37.4
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/docker/docker v1.13.1
	github.com/drep-project/binary v0.0.0-20190919035907-78d88687b9c1
	github.com/drep-project/dlog v0.0.0-20190227085123-d6565cdad12a
	github.com/drep-project/rpc v0.0.0-20190627033216-170a674fe35f
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/elastic/gosigar v0.10.5 // indirect
	github.com/ethereum/go-ethereum v1.9.7
	github.com/fatih/color v1.7.0
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/golang/snappy v0.0.1
	github.com/huin/goupnp v1.0.0
	github.com/jackpal/go-nat-pmp v1.0.1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/karalabe/usb v0.0.0-20191104083709-911d15fe12a9 // indirect
	github.com/mattn/go-colorable v0.1.4
	github.com/meling/urs v0.0.0-20140826003057-dfe7ae28e94c
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterh/liner v1.1.0
	github.com/pingcap/errors v0.11.4
	github.com/pkg/errors v0.8.1
	github.com/prometheus/tsdb v0.10.0 // indirect
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
	go.mongodb.org/mongo-driver v1.1.3
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	golang.org/x/net v0.0.0-20190912160710-24e19bdeb0f2
	golang.org/x/sys v0.0.0-20190912141932-bc967efca4b8
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/urfave/cli.v1 v1.0.0-00010101000000-000000000000
)
