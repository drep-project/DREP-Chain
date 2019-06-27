package service

import (
	"encoding/hex"
	"fmt"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/secp256k1/schnorr"
	"github.com/drep-project/drep-chain/crypto/sha3"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"log"
	"testing"
	"time"
)

func TestAddAffine2sds(t *testing.T) {
	//for i:= 0;i<100;i++{
	//0xbf32cb47d9b800820447a06840e5c25819b2cfa3ae45c9650d4bb94e3c34a045 176
	//0x28773104ef5a1964e3bb3aa47213027575ec6c16e42821a21a28a5d55ed3bce3 198
	//0xfdb91d0ddd535fce8ca2522605c49cc560b62e75f9a29fab28fd7c3507cf42a3 163

	bytes, _ := hex.DecodeString("bf32cb47d9b800820447a06840e5c25819b2cfa3ae45c9650d4bb94e3c34a045")
	priv1, _ := secp256k1.PrivKeyFromBytes(bytes)
	bytes, _ = hex.DecodeString("28773104ef5a1964e3bb3aa47213027575ec6c16e42821a21a28a5d55ed3bce3")
	priv2, _ := secp256k1.PrivKeyFromBytes(bytes)

	bytes, _ = hex.DecodeString("fdb91d0ddd535fce8ca2522605c49cc560b62e75f9a29fab28fd7c3507cf42a3")
	priv3, _ := secp256k1.PrivKeyFromBytes(bytes)

	//	priv1,_ := secp256k1.GeneratePrivateKey(nil)
	//	priv2,_ := secp256k1.GeneratePrivateKey(nil)
	//	priv3,_ := secp256k1.GeneratePrivateKey(nil)

	//setup
	msg := sha3.Keccak256([]byte{1, 2, 3})
	fmt.Println(msg)
	//commit
	randomPrivakey1, pubNonce1, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, priv1, nil, schnorr.Sha256VersionStringRFC6979)
	fmt.Println("176")
	fmt.Println(randomPrivakey1)
	fmt.Println(pubNonce1)

	time.Sleep(time.Millisecond * time.Duration(10))
	randomPrivakey2, pubNonce2, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, priv2, nil, schnorr.Sha256VersionStringRFC6979)
	fmt.Println("198")
	fmt.Println(randomPrivakey2)
	fmt.Println(pubNonce2)

	time.Sleep(time.Millisecond * time.Duration(10))
	randomPrivakey3, pubNonce3, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, priv3, nil, schnorr.Sha256VersionStringRFC6979)
	fmt.Println("163")
	fmt.Println(randomPrivakey3)
	fmt.Println(pubNonce3)

	//sig1
	sigma1 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pubNonce3, pubNonce2})
	sig1, _ := schnorr.PartialSign(secp256k1.S256(), msg, priv1, randomPrivakey1, sigma1)
	fmt.Println("-----176")
	fmt.Println(msg)
	fmt.Println(priv1)
	fmt.Println(randomPrivakey1)
	fmt.Println(sigma1)

	fmt.Println(sig1)
	fmt.Println("-----END-----")
	//sig2
	sigma2 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pubNonce3, pubNonce1})
	sig2, _ := schnorr.PartialSign(secp256k1.S256(), msg, priv2, randomPrivakey2, sigma2)

	//sig3
	sigma3 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pubNonce1, pubNonce2})
	sig3, _ := schnorr.PartialSign(secp256k1.S256(), msg, priv3, randomPrivakey3, sigma3)

	fmt.Println("-----176")
	fmt.Println(msg)
	fmt.Println(priv1)
	fmt.Println(randomPrivakey1)
	fmt.Println(sigma1)

	fmt.Println(sig3)
	fmt.Println("-----END-----")

	cSig, _ := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{sig1})
	//	fmt.Println(cSig)
	cSig, _ = schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{cSig, sig2})
	//	fmt.Println(cSig)
	cSig, _ = schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{cSig, sig3})
	//	fmt.Println(cSig)

	finalsigmaPubKey := schnorr.CombinePubkeys([]*secp256k1.PublicKey{priv1.PubKey()})
	finalsigmaPubKey = schnorr.CombinePubkeys([]*secp256k1.PublicKey{finalsigmaPubKey, priv2.PubKey()})
	finalsigmaPubKey = schnorr.CombinePubkeys([]*secp256k1.PublicKey{finalsigmaPubKey, priv3.PubKey()})
	fmt.Println("******")
	fmt.Println(msg)
	fmt.Println(finalsigmaPubKey)
	fmt.Println(cSig.R)
	fmt.Println(cSig.S)
	OK := schnorr.Verify(finalsigmaPubKey, msg, cSig.R, cSig.S)
	if !OK {
		log.Fatal("xxxx")
	}
	fmt.Println(OK)
	//	}
}

func TestAddAffine2(t *testing.T) {

	//pk1,_ := secp256k1.GeneratePrivateKey(nil)
	pk2, _ := secp256k1.GeneratePrivateKey(nil)
	pk3, _ := secp256k1.GeneratePrivateKey(nil)

	//setup
	msg := sha3.Keccak256([]byte{1, 2, 3})

	//commit
	randomPrivakey2, _, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, pk2, nil, schnorr.Sha256VersionStringRFC6979)
	randomPrivakey3, _, _ := schnorr.GenerateNoncePair(secp256k1.S256(), msg, pk3, nil, schnorr.Sha256VersionStringRFC6979)

	//oncommit
	sigmaPubKey := []*secp256k1.PublicKey{pk3.PubKey()}
	sigmaPubKey = append(sigmaPubKey, pk2.PubKey())

	sigmaCommitPubkey := []*secp256k1.PublicKey{randomPrivakey3.PubKey()}
	sigmaCommitPubkey = append(sigmaCommitPubkey, randomPrivakey2.PubKey())

	//challenge
	memIndex := 0
	sigmaPubKeys := []*secp256k1.PublicKey{}
	for index, pubkey := range sigmaPubKey {
		if !pubkey.IsEqual(pk2.PubKey()) {
			sigmaPubKeys = append(sigmaPubKeys, pubkey)
		} else {
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
	for index, pubkey := range sigmaPubKey {
		if !pubkey.IsEqual(pk3.PubKey()) {
			sigmaPubKeys = append(sigmaPubKeys, pubkey)
		} else {
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

	cSig, _ := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{sig2, sig3})
	//verify
	finalsigmaPubKey := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pk3.PubKey(), pk2.PubKey()})
	OK := schnorr.Verify(finalsigmaPubKey, msg, cSig.R, cSig.S)
	fmt.Println(OK)
}

func TestAddAffine(t *testing.T) {
	msg := []byte{1, 5, 12, 12, 5, 7}
	hash := sha3.Keccak256(msg)

	pk1, _ := secp256k1.GeneratePrivateKey(nil)
	pk2, _ := secp256k1.GeneratePrivateKey(nil)

	np1, nounce1, _ := schnorr.GenerateNoncePair(secp256k1.S256(), hash, pk1, nil, schnorr.Sha256VersionStringRFC6979)
	np2, nounce2, _ := schnorr.GenerateNoncePair(secp256k1.S256(), hash, pk2, nil, schnorr.Sha256VersionStringRFC6979)

	allPkSum := schnorr.CombinePubkeys([]*secp256k1.PublicKey{nounce1, nounce2})
	sig1, _ := schnorr.PartialSign(secp256k1.S256(), hash, pk1, np1, allPkSum)

	allPkSum2 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{nounce1, nounce2})
	sig2, _ := schnorr.PartialSign(secp256k1.S256(), hash, pk2, np2, allPkSum2)

	combineSig, _ := schnorr.CombineSigs(secp256k1.S256(), []*schnorr.Signature{sig1, sig2})

	fmt.Println(combineSig)

	allPkSum3 := schnorr.CombinePubkeys([]*secp256k1.PublicKey{pk1.PubKey(), pk2.PubKey()})
	isOk := schnorr.Verify(allPkSum3, hash, combineSig.R, combineSig.S)

	fmt.Print(isOk)
}
