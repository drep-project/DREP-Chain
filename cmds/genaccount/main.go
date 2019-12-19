package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/bft"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/solo"
	"github.com/drep-project/DREP-Chain/pkgs/log"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"net"
	"os"
	path2 "path"
	"path/filepath"
	"time"

	p2pTypes "github.com/drep-project/DREP-Chain/network/types"
	accountComponent "github.com/drep-project/DREP-Chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/DREP-Chain/pkgs/accounts/types"
	chainIndexerTypes "github.com/drep-project/DREP-Chain/pkgs/chain_indexer"
	filterTypes "github.com/drep-project/DREP-Chain/pkgs/filter"
	"github.com/drep-project/DREP-Chain/types"
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
	cfg, err := parserConfig(cfgPath)
	if err != nil {
		return err
	}
	path := ""
	if ctx.GlobalIsSet(pathFlag.Name) {
		path = ctx.GlobalString(pathFlag.Name)
	} else {
		path = appPath
	}
	nodeItems := cfg.Miners
	bootsNodes := []*enode.Node{}
	standbyKey := []*secp256k1.PrivateKey{}
	nodes := []*types.Node{}
	produces := make([]types.CandidateData, 0)
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
		produces = append(produces, types.CandidateData{
			Node:   node.String(),
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

	consensusConfig := &service.ConsensusConfig{}

	if len(nodeItems) == 1 {
		consensusConfig.ConsensusMode = "solo"
		consensusConfig.Solo = &solo.SoloConfig{
			MyPk:          nil,
			StartMiner:    true,
			BlockInterval: 5,
		}
	} else {
		consensusConfig.ConsensusMode = "bft"
		consensusConfig.Bft = &bft.BftConfig{
			MyPk:           nil,
			StartMiner:     true,
			BlockInterval:  5,
			ProducerNum:    len(nodeItems),
			ChangeInterval: 100,
		}
	}

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

	if len(nodeItems) == 1 {
		consensusConfig.Bft.MyPk = (*secp256k1.PublicKey)(&standbyKey[0].PublicKey)
		userDir := path2.Join(path, nodeItems[0].Name)
		os.MkdirAll(userDir, os.ModeDir|os.ModePerm)
		keyStorePath := path2.Join(userDir, "keystore")
		password := "123"
		if nodeItems[0].Password != "" {
			password = nodeItems[0].Password
		}

		store := accountComponent.NewFileStore(keyStorePath, nil)
		cryptoPassowrd := string(sha3.Keccak256([]byte(password)))
		store.StoreKey(nodes[0], cryptoPassowrd)

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
		offset = writePhase(fs, "genesis",
			struct {
				Preminer []*chain.Preminer
				Miners   []types.CandidateData
			}{
				Preminer: cfg.Preminer,
				Miners:   produces,
			}, offset)
		fs.Truncate(offset - 2)
		fs.WriteAt([]byte("\n}"), offset-2)
	} else {
		for i := 0; i < len(nodeItems); i++ {
			consensusConfig.Bft.MyPk = (*secp256k1.PublicKey)(&standbyKey[i].PublicKey)
			userDir := path2.Join(path, nodeItems[i].Name)
			os.MkdirAll(userDir, os.ModeDir|os.ModePerm)
			keyStorePath := path2.Join(userDir, "keystore")
			password := "123"
			if nodeItems[i].Password != "" {
				password = nodeItems[i].Password
			}

			store := accountComponent.NewFileStore(keyStorePath, nil)
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

			offset = writePhase(fs, "genesis",
				struct {
					Preminer []*chain.Preminer
					Miners   []types.CandidateData
				}{
					Preminer: cfg.Preminer,
					Miners:   produces,
				}, offset)
			fs.Truncate(offset - 2)
			fs.WriteAt([]byte("\n}"), offset-2)
		}
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

	prvKey, _ = crypto.GenerateKey(rand.Reader)
	chainCode = append(seed, []byte(types.DrepMark)...)
	chainCode = common.HmAC(chainCode, prvKey.PubKey().Serialize())

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

func parserConfig(cfgPath string) (*GenesisConfig, error) {
	content, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg := &GenesisConfig{}
	err = json.Unmarshal([]byte(content), &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type GenesisConfig struct {
	Preminer []*chain.Preminer
	Miners   []*NodeItem
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
