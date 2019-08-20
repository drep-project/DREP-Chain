package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/network/p2p/enode"
	"github.com/drep-project/drep-chain/params"
	"github.com/drep-project/drep-chain/pkgs/log"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"net"
	"os"
	path2 "path"
	"path/filepath"
	"time"

	p2pTypes "github.com/drep-project/drep-chain/network/types"
	accountComponent "github.com/drep-project/drep-chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/pkgs/accounts/types"
	chainIndexerTypes "github.com/drep-project/drep-chain/pkgs/chain_indexer"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	filterTypes "github.com/drep-project/drep-chain/pkgs/filter"
	"github.com/drep-project/drep-chain/types"
	"github.com/drep-project/rpc"
)

var (
	parentNode = types.NewNode(nil, 0)
	pathFlag   = common.DirectoryFlag{
		Name:  "path",
		Usage: "keystore save to",
	}
)

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		pathFlag,
	}
	app.Action = gen
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func gen(ctx *cli.Context) error {
	appPath := getCurPath()
	cfgPath := path2.Join(appPath, "config.json")
	nodeItems, err := parserConfig(cfgPath)
	if err != nil {
		return err
	}
	path := ""
	if ctx.GlobalIsSet(pathFlag.Name) {
		path = ctx.GlobalString(pathFlag.Name)
	} else {
		path = appPath
	}
	bootsNodes := []*enode.Node{}
	standbyKey := []*secp256k1.PrivateKey{}
	nodes := []*types.Node{}
	produces := make([]consensusTypes.Producer, 0)
	for i := 0; i < len(nodeItems); i++ {
		aNode := getAccount(nodeItems[i].Name)
		nodes = append(nodes, aNode)
		ip := net.IP{}
		err := ip.UnmarshalText([]byte(nodeItems[i].Ip))
		if err != nil {
			return err
		}
		instanceDir := filepath.Join(path, nodeItems[i].Name, "drepnode")
		nodePrivateKey := GeneratePrivateKey(instanceDir)
		fmt.Println(crypto.PubkeyToAddress(nodePrivateKey.PubKey()).String(), hex.EncodeToString(nodePrivateKey.Serialize()))
		node := enode.NewV4(nodePrivateKey.PubKey(), ip, nodeItems[i].Port, nodeItems[i].Port)
		bootsNodes = append(bootsNodes, node)

		standbyKey = append(standbyKey, aNode.PrivateKey)
		produces = append(produces, consensusTypes.Producer{
			IP:     nodeItems[i].Ip,
			Pubkey: aNode.PrivateKey.PubKey(),
		})
	}

	logConfig := log.LogConfig{}
	logConfig.LogLevel = 4

	rpcConfig := rpc.RpcConfig{}
	rpcConfig.IPCEnabled = true
	rpcConfig.HTTPEnabled = true
	p2pConfig := p2pTypes.P2pConfig{}
	p2pConfig.MaxPeers = 20
	p2pConfig.NoDiscovery = false
	p2pConfig.DiscoveryV5 = true
	p2pConfig.Name = "drepnode"
	p2pConfig.ProduceNodes = bootsNodes
	p2pConfig.StaticNodes = bootsNodes
	p2pConfig.ListenAddr = "0.0.0.0:55555"

	consensusConfig := consensusTypes.ConsensusConfig{}
	consensusConfig.Enable = true
	consensusConfig.ConsensusMode = "bft"
	consensusConfig.Producers = produces
	//consensusConfig.Producers = produces

	chainConfig := chain.ChainConfig{}
	chainConfig.RemotePort = 55556
	chainConfig.ChainId = 0
	chainConfig.GenesisAddr = params.HoleAddress

	chainIndexerConfig := chainIndexerTypes.ChainIndexerConfig{}
	chainIndexerConfig.Enable = true
	chainIndexerConfig.SectionSize = 4096
	chainIndexerConfig.ConfirmsReq = 256
	chainIndexerConfig.Throttling = 100 * time.Millisecond

	filterConfig := filterTypes.FilterConfig{}
	filterConfig.Enable = true

	for i := 0; i < len(nodeItems); i++ {
		consensusConfig.MyPk = (*secp256k1.PublicKey)(&standbyKey[i].PublicKey)
		userDir := path2.Join(path, nodeItems[i].Name)
		os.MkdirAll(userDir, os.ModeDir|os.ModePerm)
		keyStorePath := path2.Join(userDir, "keystore")
		password := "123"
		if nodeItems[i].Password != "" {
			password = nodeItems[i].Password
		}

		store := accountComponent.NewFileStore(keyStorePath)
		cryptoPassowrd := string(sha3.Keccak256([]byte(password)))
		store.StoreKey(nodes[i], cryptoPassowrd)

		walletConfig := accountTypes.Config{}
		walletConfig.Enable = true
		walletConfig.Password = password

		cfgPath := path2.Join(userDir, "config.json")
		fs, _ := os.Create(cfgPath)
		offset := int64(0)
		fs.WriteAt([]byte("{\n"), offset)
		offset = int64(2)

		offset = writePhase(fs, "log", logConfig, offset)
		offset = writePhase(fs, "rpc", rpcConfig, offset)
		offset = writePhase(fs, "consensus", consensusConfig, offset)
		offset = writePhase(fs, "p2p", p2pConfig, offset)
		offset = writePhase(fs, "chain", chainConfig, offset)
		offset = writePhase(fs, "accounts", walletConfig, offset)
		offset = writePhase(fs, "chain_indexer", chainIndexerConfig, offset)
		offset = writePhase(fs, "filter", filterConfig, offset)

		fs.Truncate(offset - 2)
		fs.WriteAt([]byte("\n}"), offset-2)
	}
	return nil
}

func writePhase(fs *os.File, name string, config interface{}, offset int64) int64 {
	bytes, _ := json.MarshalIndent(config, "	", "      ")
	bytes = append([]byte("	\""+name+"\" : "), bytes...)
	fs.WriteAt(bytes, offset)
	offset += int64(len(bytes))

	fs.WriteAt([]byte(",\n"), offset)
	offset += 2
	return offset
}

func getAccount(name string) *types.Node {
	node := RandomNode([]byte(name))
	return node
}

func RandomNode(seed []byte) *types.Node {
	var (
		prvKey    *secp256k1.PrivateKey
		chainCode []byte
	)

	h := hmAC(seed, types.DrepMark)
	prvKey, _ = crypto.ToPrivateKey(h[:types.KeyBitSize])
	chainCode = h[types.KeyBitSize:]
	addr := crypto.PubkeyToAddress(prvKey.PubKey())
	return &types.Node{
		PrivateKey: prvKey,
		Address:    &addr,
		ChainId:    0,
		ChainCode:  chainCode,
	}
}

func hmAC(message, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func getCurPath() string {
	dir, _ := os.Getwd()
	return dir
}

func parserConfig(cfgPath string) ([]*NodeItem, error) {
	content, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg := []*NodeItem{}
	err = json.Unmarshal([]byte(content), &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type NodeItem struct {
	Name     string
	Ip       string
	Port     int
	Password string
}

func GeneratePrivateKey(instanceDir string) *secp256k1.PrivateKey {
	key, err := crypto.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Println("Failed to generate node key: %v", err)
	}

	if err := os.MkdirAll(instanceDir, 0700); err != nil {
		fmt.Println("Failed to persist node key: %v", err)
		return key
	}

	keyfile := filepath.Join(instanceDir, "nodekey")
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		fmt.Println("Failed to persist node key: %v", err)
	}
	return key
}
