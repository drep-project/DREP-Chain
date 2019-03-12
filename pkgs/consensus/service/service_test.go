package service

import (
	"testing"
)

func TestAccumulateRewards(t *testing.T) {
	//b := common.Bytes("0x03177b8e4ef31f4f801ce00260db1b04cc501287e828692a404fdbc46c7ad6ff26")
	//err := b.UnmarshalText(b)
	//signPubkey, err := secp256k1.ParsePubKey(b)
	//
	//if err != nil {
	//	panic(err)
	//}
	//ce := ConsensusService{
<<<<<<< HEAD
	//	leader: NewLeader(signPubkey, nil, nil),
	//	member: &Member{members: []*secp256k1.PublicKey{signPubkey, signPubkey, signPubkey}},
=======
	//	leader: NewLeader(pubkey, nil, nil),
	//	member: &Member{producers: []*secp256k1.PublicKey{pubkey, pubkey, pubkey}},
>>>>>>> fix member find
	//	DatabaseService: &database.DatabaseService{},
	//}
	//
	//ctx := &app.ExecuteContext{
	//	CommonConfig:&app.CommonConfig{HomeDir:"./"},
	//	//CliContext: cli.NewContext(nil,nil,nil)
	//}
	//
	//ctx.AddService( &database.DatabaseService{})
	//
	//ce.DatabaseService.Init(ctx)
	//
	//ty := common.ChainIdType{1, 2, 3, 4, 5, 6}
	//ce.accumulateRewards(ty)
}
