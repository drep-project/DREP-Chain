package addrgenerator

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	rippleCrypto "github.com/rubblelabs/ripple/crypto"
	"github.com/sasaxie/go-client-api/common/base58"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type AddrGenerate struct {
	PrivateKey *secp256k1.PrivateKey
}

func (addrGenerate *AddrGenerate) ToEth() string {
	pk := addrGenerate.PrivateKey.PubKey()
	ecdsaPk := (*ecdsa.PublicKey)(pk)
	return ethcrypto.PubkeyToAddress(*ecdsaPk).String()
}
func (addrGenerate *AddrGenerate) ToRipple() string {
	bytes := rippleCrypto.Sha256RipeMD160(addrGenerate.PrivateKey.PubKey().SerializeCompressed())
	hash, _ := rippleCrypto.NewAccountId(bytes)
	return hash.String()
}
func (addrGenerate *AddrGenerate) ToNeo() string {
	pub_bytes := addrGenerate.PrivateKey.PubKey().Serialize()

	pub_bytes = append([]byte{0x21}, pub_bytes...)
	pub_bytes = append(pub_bytes, 0xAC)

	/* SHA256 Hash */
	sha256_h := sha256.New()
	sha256_h.Reset()
	sha256_h.Write(pub_bytes)
	pub_hash_1 := sha256_h.Sum(nil)

	/* RIPEMD-160 Hash */
	ripemd160_h := ripemd160.New()
	ripemd160_h.Reset()
	ripemd160_h.Write(pub_hash_1)
	pub_hash_2 := ripemd160_h.Sum(nil)

	program_hash := pub_hash_2
	return addrGenerate.b58checkencodeNEO(0x17, program_hash)
}

func (addrGenerate *AddrGenerate) ToLiteCoin() string {
	coin := getCoin("Litecoin")
	return genCoin(addrGenerate.PrivateKey, coin.PubKeyHashAddrID, coin.PrivateKeyID, coin.Name)
}

func (addrGenerate *AddrGenerate) ToDogecoin() string {
	coin := getCoin("Dogecoin")
	return genCoin(addrGenerate.PrivateKey, coin.PubKeyHashAddrID, coin.PrivateKeyID, coin.Name)
}

func (addrGenerate *AddrGenerate) ToDash() string {
	coin := getCoin("Dash")
	return genCoin(addrGenerate.PrivateKey, coin.PubKeyHashAddrID, coin.PrivateKeyID, coin.Name)
}

func (addrGenerate *AddrGenerate) ToAtom() string {
	pubKey := addrGenerate.PrivateKey.PubKey()
	addr := sdk.AccAddress(pubKey.Serialize())
	return addr.String()
}

func (addrGenerate *AddrGenerate) ToTron() string {
	return addressFromKey(addrGenerate.PrivateKey)
}

func (addrGenerate *AddrGenerate) ToBtc() string {
	coin := getCoin("Bitcoin")
	return genCoin(addrGenerate.PrivateKey, coin.PubKeyHashAddrID, coin.PrivateKeyID, coin.Name)
}

func (addrGenerate *AddrGenerate) b58checkencodeNEO(ver uint8, b []byte) (s string) {
	/* Prepend version */
	bcpy := append([]byte{ver}, b...)

	/* Create a new SHA256 context */
	sha256_h := sha256.New()

	/* SHA256 Hash #1 */
	sha256_h.Reset()
	sha256_h.Write(bcpy)
	hash1 := sha256_h.Sum(nil)

	/* SHA256 Hash #2 */
	sha256_h.Reset()
	sha256_h.Write(hash1)
	hash2 := sha256_h.Sum(nil)

	/* Append first four bytes of hash */
	bcpy = append(bcpy, hash2[0:4]...)

	/* Encode base58 string */
	s = b58encode(bcpy)

	// /* For number of leading 0's in bytes, prepend 1 */
	// for _, v := range bcpy {
	// 	if v != 0 {
	// 		break
	// 	}
	// 	s = "1" + s
	// }

	return s
}

// b58encode encodes a byte slice b into a base-58 encoded string.
func b58encode(b []byte) (s string) {
	/* See https://en.bitcoin.it/wiki/Base58Check_encoding */

	const BITCOIN_BASE58_TABLE = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	/* Convert big endian bytes to big int */
	x := new(big.Int).SetBytes(b)

	/* Initialize */
	r := new(big.Int)
	m := big.NewInt(58)
	zero := big.NewInt(0)
	s = ""

	/* Convert big int to string */
	for x.Cmp(zero) > 0 {
		/* x, r = (x / 58, x % 58) */
		x.QuoRem(x, m, r)
		/* Prepend ASCII character */
		s = string(BITCOIN_BASE58_TABLE[r.Int64()]) + s
	}

	return s
}

func genCoin(pk *secp256k1.PrivateKey, PubKeyHashAddrID, PrivateKeyID byte, name string) string {
	net := &chaincfg.MainNetParams
	net.PubKeyHashAddrID = PubKeyHashAddrID
	net.PrivateKeyID = PrivateKeyID
	edsaPriv := (*ecdsa.PrivateKey)(pk)
	btcPriv := (*btcec.PrivateKey)(edsaPriv)
	wif, _ := btcutil.NewWIF(btcPriv, net, true)
	addr, _ := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), net)
	return addr.EncodeAddress()
}

func addressFromKey(secpKey *secp256k1.PrivateKey) string {
	// Build the Private Key and extract the Public Key
	key := (*ecdsa.PrivateKey)(secpKey)
	// #1
	pub := append(key.X.Bytes(), key.Y.Bytes()...)

	// #2
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pub)
	hashed := hash.Sum(nil)
	last20 := hashed[len(hashed)-20:]

	// #3
	addr41 := append([]byte{0x41}, last20...)

	// #4
	hash2561 := sha256.Sum256(addr41)
	hash2562 := sha256.Sum256(hash2561[:])
	checksum := hash2562[:4]

	// #5/#6
	rawAddr := append(addr41, checksum...)
	tronAddr := base58.Encode(rawAddr)

	return tronAddr
}
