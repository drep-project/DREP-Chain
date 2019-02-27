package service

import (
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"fmt"
	"math/big"
	"testing"
)


func TestAddAffine2(t *testing.T) {

	//pk1,_ := secp256k1.GeneratePrivateKey(nil)
	pk2,_ := secp256k1.GeneratePrivateKey(nil)
	pk3,_ := secp256k1.GeneratePrivateKey(nil)

	//setup
	msg := sha3.Hash256([]byte{1,2,3})

	//commit
	randomPrivakey2, _, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, pk2,nil, schnorr.Sha256VersionStringRFC6979)
	randomPrivakey3, _, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, pk3,nil, schnorr.Sha256VersionStringRFC6979)

	//oncommit
	sigmaPubKey := []*secp256k1.PublicKey{ pk3.PubKey() }
	sigmaPubKey = append(sigmaPubKey,  pk2.PubKey())

	sigmaCommitPubkey := []*secp256k1.PublicKey{ randomPrivakey3.PubKey() }
	sigmaCommitPubkey = append(sigmaCommitPubkey,  randomPrivakey2.PubKey())

	//challenge
	memIndex := 0
	sigmaPubKeys := []*secp256k1.PublicKey{}
	for index, pubkey := range  sigmaPubKey {
		if !pubkey.IsEqual(pk2.PubKey()) {
			sigmaPubKeys = append(sigmaPubKeys, pubkey)
		}else{
			memIndex = index
		}
	}
	sigmaPubKeyss2 := schnorr.CombinePubkeys(sigmaPubKeys)

	commitPubkeys := []*secp256k1.PublicKey{}
	for index, pubkey := range sigmaCommitPubkey {
		if memIndex != index {
			commitPubkeys = append(commitPubkeys, pubkey)
		}
	}
	commitPubkey2 := schnorr.CombinePubkeys(commitPubkeys)
	challengeMsg := &consensusTypes.Challenge{SigmaPubKey: sigmaPubKeyss2, SigmaQ: commitPubkey2}
	//onchallenge
	sig2, _ := schnorr.PartialSign(secp256k1.S256(), msg, pk2, randomPrivakey2, challengeMsg.SigmaQ)

	memIndex = 0
	sigmaPubKeys = []*secp256k1.PublicKey{}
	for index, pubkey := range  sigmaPubKey {
		if !pubkey.IsEqual(pk3.PubKey()) {
			sigmaPubKeys = append(sigmaPubKeys, pubkey)
		}else{
			memIndex = index
		}
	}
	sigmaPubKeyss3 := schnorr.CombinePubkeys(sigmaPubKeys)

	commitPubkeys = []*secp256k1.PublicKey{}
	for index, pubkey := range sigmaCommitPubkey {
		if memIndex != index {
			commitPubkeys = append(commitPubkeys, pubkey)
		}
	}
	commitPubkey3 := schnorr.CombinePubkeys(commitPubkeys)
	challengeMsg = &consensusTypes.Challenge{SigmaPubKey: sigmaPubKeyss3, SigmaQ: commitPubkey3}
	//onchallenge
	sig3, _ := schnorr.PartialSign(secp256k1.S256(), msg, pk3, randomPrivakey3, challengeMsg.SigmaQ)

	//response
	sigmaS := sig3
	cSig, _  := schnorr.CombineSigs(secp256k1.S256(),[]*schnorr.Signature{sigmaS, sig2 })

	//verify
	finalsigmaPubKey := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pk3.PubKey(),pk2.PubKey()})
	OK :=  schnorr.Verify(finalsigmaPubKey, msg, cSig.R, cSig.S)
	fmt.Println(OK)
}


func TestAddAffine(t *testing.T) {
	msg := []byte{1,5,12,12,5,7,}
	hash :=sha3.Hash256(msg)

	pk1,_ := secp256k1.GeneratePrivateKey(nil)
	pk2,_ := secp256k1.GeneratePrivateKey(nil)


	np1, nounce1,_ := schnorr.GenerateNoncePair(secp256k1.S256(),hash, pk1,nil, schnorr.Sha256VersionStringRFC6979)
	np2, nounce2,_ := schnorr.GenerateNoncePair(secp256k1.S256(),hash, pk2,nil, schnorr.Sha256VersionStringRFC6979)

	allPkSum := schnorr.CombinePubkeys([]*secp256k1.PublicKey{nounce1,nounce2})
	allPkSum.Add(allPkSum.X, allPkSum.Y, new (big.Int).Neg(nounce1.X),  new (big.Int).Neg(nounce1.Y))
 	sig1, _ := schnorr.PartialSign(secp256k1.S256(), hash, pk1, np1, allPkSum)

	allPkSum2 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{nounce1,nounce2})
	allPkSum2.Add(allPkSum2.X,allPkSum2.Y, new (big.Int).Neg(nounce2.X),  new (big.Int).Neg(nounce2.Y))
	sig2, _ := schnorr.PartialSign(secp256k1.S256(), hash,pk2, np2, allPkSum2)


	combineSig, _ := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{sig1,sig2})

	fmt.Println(combineSig)

	allPkSum3 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pk1.PubKey(),pk2.PubKey()})
	isOk := schnorr.Verify(allPkSum3, hash, combineSig.R,combineSig.S)

	fmt.Print(isOk)
}