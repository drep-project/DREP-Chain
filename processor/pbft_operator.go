package processor
//
//import (
//    "BlockChainTest/bean"
//    "math"
//    "math/big"
//    "BlockChainTest/crypto"
//)
//
//type ChallengeGenerator struct {
//    PrvKey *bean.PrivateKey
//    Object interface{}
//    ChallengeBytes []byte
//    MinorCommitmentMap map[string] *bean.Commitment
//}
//
//func (g *ChallengeGenerator) Generate() error {
//    obj, err := bean.Marshal(g.Object)
//    if err != nil {
//        return err
//    }
//    curve := crypto.GetCurve()
//    groupPubKey := &bean.Point{}
//    groupQ := &bean.Point{}
//    for _, commitment := range g.MinorCommitmentMap {
//        groupPubKey = curve.Add(groupPubKey, commitment.PubKey)
//        groupQ = curve.Add(groupQ, commitment.Q)
//    }
//    r := crypto.ConcatHash256(groupQ.Bytes(), groupPubKey.Bytes(), obj)
//    challenge := &bean.Challenge{GroupPubKey: groupPubKey, GroupQ: groupQ, Object: obj, R: r}
//    chaBytes, err := bean.Marshal(challenge)
//    if err != nil {
//       g.ChallengeBytes = chaBytes
//    }
//    return nil
//}
//
//type ResponseValidator struct {
//    R []byte
//    ChallengeBytes []byte
//    MinorNum int
//    MinorCommitmentMap map[string] *bean.Commitment
//    MinorResponseMap map[string] *bean.Response
//}
//
//func (v *ResponseValidator) Validate() bool {
//    if len(v.MinorResponseMap) < len(v.MinorCommitmentMap) {
//        return false
//    }
//    if float64(len(v.MinorResponseMap)) < math.Ceil(float64(v.MinorNum * 2.0 / 3.0) + 1) {
//        return false
//    }
//    curve := crypto.GetCurve()
//    groupPubKey := &bean.Point{}
//    groupS := new(big.Int)
//    for addr, response := range v.MinorResponseMap {
//        if _, b := v.MinorCommitmentMap[addr]; !b {
//            return false
//        }
//        groupPubKey = curve.Add(groupPubKey, response.PubKey)
//        groupS.Add(groupS, new(big.Int).SetBytes(response.S))
//    }
//    groupS.Mod(groupS, curve.N)
//    sig := &bean.Signature{R: v.R, S: groupS.Bytes()}
//    return crypto.Verify(sig, groupPubKey, v.ChallengeBytes)
//}
