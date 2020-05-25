module github.com/drep-project/DREP-Chain

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

go 1.13

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20200317173033-c483abfa40e2
	github.com/allegro/bigcache v1.2.1
	github.com/aristanetworks/goarista v0.0.0-20200513152637-638451432ae4
	github.com/asaskevich/EventBus v0.0.0-20200428142821-4fc0642a29f3
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.38.3
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v1.13.1
	github.com/drep-project/binary v0.0.0-20190919035907-78d88687b9c1
	github.com/drep-project/dlog v0.0.0-20200514080736-e9b04787eae9
	github.com/drep-project/drep-chain v0.0.0-20200513090939-9faecf3157e0
	github.com/drep-project/rpc v0.0.0-20200514081031-c317a9a1ce9d
	github.com/ethereum/go-ethereum v1.9.14
	github.com/fatih/color v1.9.0
	github.com/golang/snappy v0.0.1
	github.com/huin/goupnp v1.0.0
	github.com/jackpal/go-nat-pmp v1.0.2
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mattn/go-colorable v0.1.6
	github.com/meling/urs v0.0.0-20140826003057-dfe7ae28e94c
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/peterh/liner v1.2.0
	github.com/pingcap/errors v0.11.4
	github.com/pkg/errors v0.9.1
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff
	github.com/rubblelabs/ripple v0.0.0-20200504074039-c8844ec07a94
	github.com/sasaxie/go-client-api v0.0.0-20190820063117-f0587df4b72e
	github.com/shengdoushi/base58 v1.0.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/vishalkuo/bimap v0.0.0-20180703190407-09cff2814645
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	go.mongodb.org/mongo-driver v1.3.3
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200513185701-a91f0712d120
	golang.org/x/sys v0.0.0-20200513112337-417ce2331b5c
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
)
