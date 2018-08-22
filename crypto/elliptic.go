package crypto

import (
	"math/big"
	"crypto/rand"
	"bytes"
    "errors"
    "BlockChainTest/common"
)

var Zero = new(big.Int)
var MaxRandomRetry = 10
var ByteLen = 32

type Curve interface {
	Params() *CurveParams
	IsOnCurve(*common.Point) bool
	Add(*common.Point, *common.Point) *common.Point
	Double(*common.Point) *common.Point
	ScalarMultiply(*common.Point, []byte) *common.Point
	ScalarBaseMultiply([]byte) *common.Point
}

// Y^2 == X^3 + AX + B (mod p), with a == 0
type CurveParams struct {
	P *big.Int
	N *big.Int
	B *big.Int
	G *common.Point
	BitSize int
	Name string
}

type JacobiCoordinate struct {
	X *big.Int
	Y *big.Int
	Z *big.Int
}

func (curveParams *CurveParams) Params() *CurveParams {
	return curveParams
}

// Y^2 == X^3 + 7 (mod p)
func (curveParams *CurveParams) IsOnCurve(point *common.Point) bool {
	x, y := new(big.Int).SetBytes(point.X), new(big.Int).SetBytes(point.Y)
	P := curveParams.P
	B := curveParams.B
	ySquare := new(big.Int).Mul(y, y)
	ySquare.Mod(ySquare, P)
	xCube := new(big.Int).Mul(x, x)
	xCube.Mod(xCube, P)
	xCube.Mul(xCube, x)
	xPolynomial := new(big.Int).Add(xCube, B)
	xPolynomial.Mod(xPolynomial, P)
	return ySquare.Cmp(xPolynomial) == 0
}

func JacobiAffine(point *common.Point) *JacobiCoordinate {
	x, y, z := new(big.Int).SetBytes(point.X), new(big.Int).SetBytes(point.Y), new(big.Int)
	if x.Sign() != 0 || y.Sign() != 0 {
		z.SetInt64(1)
	}
	return &JacobiCoordinate{x, y, z}
}

func (curveParams *CurveParams) InverseJacobiAffine(jc *JacobiCoordinate) *common.Point {
	x, y, z := jc.X, jc.Y, jc.Z
	if z.Sign() == 0 {
		return &common.Point{X: new(big.Int).Bytes(), Y: new(big.Int).Bytes()}
	}
	P := curveParams.P
	zInv := new(big.Int).ModInverse(z, P)
	zInvSquare := new(big.Int).Mul(zInv, zInv)
	zInvSquare.Mod(zInvSquare, P)
	zInvCube := new(big.Int).Mul(zInvSquare, zInv)
	zInvCube.Mod(zInvCube, P)
	xOut := new(big.Int).Mul(x, zInvSquare)
	xOut.Mod(xOut, P)
	yOut := new(big.Int).Mul(y, zInvCube)
	yOut.Mod(yOut, P)
	return &common.Point{X: xOut.Bytes(), Y: yOut.Bytes()}
}

// add-2007-bl addition
// Cost: 11M + 5S + 9add + 4*2
// Cost: 10M + 4S + 9add + 4*2 dependent upon the first point
// Source: 2007 Bernstein–Lange; note that the improvement from 12M+4S to 11M+5S was already mentioned in 2001 Bernstein http://cr.yp.to/talks.html#2001.10.29
// Explicit formulas:
// Explicit formulas:
//      Z1Z1 = Z1^2
//      Z2Z2 = Z2^2
//      U1 = X1*Z2Z2
//      U2 = X2*Z1Z1
//      S1 = Y1*Z2*Z2Z2
//      S2 = Y2*Z1*Z1Z1
//      H = U2-U1
//      I = (2*H)^2
//      J = H*I
//      R = 2*(S2-S1)
//      V = U1*I
//      X3 = R^2-J-2*V
//      Y3 = R*(V-X3)-2*S1*J
//      Z3 = ((Z1+Z2)^2-Z1Z1-Z2Z2)*H
func (curveParams *CurveParams) JacobiAddition(jc1, jc2 *JacobiCoordinate) *JacobiCoordinate {
	x1, y1, z1 := jc1.X, jc1.Y, jc1.Z
	x2, y2, z2 := jc2.X, jc2.Y, jc2.Z
	x3, y3, z3 := new(big.Int), new(big.Int), new(big.Int)
	if z1.Sign() == 0 {
		return jc2
	}
	if z2.Sign() == 0 {
		return jc1
	}
	P := curveParams.P
	z1Square := new(big.Int).Mul(z1, z1)
	z1Square.Mod(z1Square, P)
	z1Cube := new(big.Int).Mul(z1Square, z1)
	z1Cube.Mod(z1Cube, P)
	z2Square := new(big.Int).Mul(z2, z2)
	z2Square.Mod(z2Square, P)
	z2Cube := new(big.Int).Mul(z2Square, z2)
	z2Cube.Mod(z2Cube, P)
	u1 := new(big.Int).Mul(x1, z2Square)
	u1.Mod(u1, P)
	u2 := new(big.Int).Mul(x2, z1Square)
	u2.Mod(u2, P)
	s1 := new(big.Int).Mul(y1, z2Cube)
	s1.Mod(s1, P)
	s2 := new(big.Int).Mul(y2, z1Cube)
	s2.Mod(s2, P)
	h := new(big.Int).Sub(u2, u1)
	h.Mod(h, P)
	i := new(big.Int).Lsh(h, 1)
	i.Mul(i, i)
	i.Mod(i, P)
	j := new(big.Int).Mul(h, i)
	j.Mod(j, P)
	r := new(big.Int).Sub(s2, s1)
	r.Lsh(r, 1)
	r.Mod(r, P)
	v := new(big.Int).Mul(u1, i)
	v.Mod(v, P)
	rSquare := new(big.Int).Mul(r, r)
	rSquare.Mod(rSquare, P)
	vDouble := new(big.Int).Lsh(v, 1)
	x3.Add(x3, rSquare)
	x3.Sub(x3, j)
	x3.Sub(x3, vDouble)
	x3.Mod(x3, P)
	y3Item1 := new(big.Int).Sub(v, x3)
	y3Item1.Mul(y3Item1, r)
	y3Item1.Mod(y3Item1, P)
	y3Item2 := new(big.Int).Lsh(s1, 1)
	y3Item2.Mul(y3Item2, j)
	y3Item2.Mod(y3Item2, P)
	y3.Sub(y3Item1, y3Item2)
	y3.Mod(y3, P)
	z3.Mul(z1, z2)
	z3.Mul(z3, h)
	z3.Lsh(z3, 1)
	z3.Mod(z3, P)
	return &JacobiCoordinate{x3, y3, z3}
}

// dbl-2007-bl
// Cost: 1M + 8S + 1*A + 10add + 2*2 + 1*3 + 1*8
// Source: 2007 Bernstein–Lange
// Explicit formulas:
//      XX = X1^2
//      YY = Y1^2
//      YYYY = YY^2
//      ZZ = Z1^2
//      S = 2*((X1+YY)^2-XX-YYYY)
//      M = 3*XX
//      T = M^2-2*S
//      X3 = T
//      Y3 = M*(S-T)-8*YYYY
//      Z3 = (Y1+Z1)^2-YY-ZZ
func (curveParams *CurveParams) JacobiDoubling(jc *JacobiCoordinate) *JacobiCoordinate {
	x1, y1, z1 := jc.X, jc.Y, jc.Z
	x2, y2, z2 := new(big.Int), new(big.Int), new(big.Int)
	if z1.Sign() == 0 {
		return jc
	}
	P := curveParams.P
	x1Square := new(big.Int).Mul(x1, x1)
	x1Square.Mod(x1Square, P)
	y1Square := new(big.Int).Mul(y1, y1)
	y1Square.Mod(y1Square, P)
	y1Biquadratic := new(big.Int).Mul(y1Square, y1Square)
	y1Biquadratic.Mod(y1Biquadratic, P)
	z1Square := new(big.Int).Mul(z1, z1)
	z1Square.Mod(z1Square, P)
	s := new(big.Int).Mul(x1, y1Square)
	s.Lsh(s, 2)
	s.Mod(s, P)
	m := new(big.Int).Lsh(x1Square, 1)
	m.Add(m, x1Square)
	m.Mod(m, P)
	t := new(big.Int).Mul(m, m)
	sDouble := new(big.Int).Lsh(s, 1)
	t.Sub(t, sDouble)
	t.Mod(t, P)
	x2.Set(t)
	y2Item1 := new(big.Int).Sub(s, t)
	y2Item1.Mul(y2Item1, m)
	y2Item1.Mod(y2Item1, P)
	y2Item2 := new(big.Int).Lsh(y1Biquadratic, 3)
	y2.Sub(y2Item1, y2Item2)
	y2.Mod(y2, P)
	z2.Mul(y1, z1)
	z2.Lsh(z2, 1)
	z2.Mod(z2, P)
	return &JacobiCoordinate{x2, y2, z2}
}

func (curveParams *CurveParams) Add(pt1, pt2 *common.Point) *common.Point {
	jc1 := JacobiAffine(pt1)
	jc2 := JacobiAffine(pt2)
	return curveParams.InverseJacobiAffine(curveParams.JacobiAddition(jc1, jc2))
}

func (curveParams *CurveParams) Double(point *common.Point) *common.Point {
	jc := JacobiAffine(point)
	return curveParams.InverseJacobiAffine(curveParams.JacobiDoubling(jc))
}

func (curveParams *CurveParams) ScalarMultiply(point *common.Point, k []byte) *common.Point {
	jc0 := JacobiAffine(point)
	jc := &JacobiCoordinate{new(big.Int), new(big.Int), new(big.Int)}
	for _, byt := range k {
		for bitNum := 0; bitNum < 8; bitNum++ {
			jc = curveParams.JacobiDoubling(jc)
			if byt & 0x80 == 0x80 {
				jc = curveParams.JacobiAddition(jc, jc0)
			}
			byt <<= 1
		}
	}
	return curveParams.InverseJacobiAffine(jc)
}

func (curveParams *CurveParams) ScalarBaseMultiply(k []byte) *common.Point {
	return curveParams.ScalarMultiply(curveParams.G, k)
}

func (curveParams *CurveParams) ScalarBaseMultiplyByFormula(k int) *common.Point {
	Gx, Gy := new(big.Int).SetBytes(curveParams.G.X), new(big.Int).SetBytes(curveParams.G.Y)
	Fx, Fy := new(big.Int).Set(Gx), new(big.Int).Set(Gy)
	P := curveParams.P
	for i := 1; i < k; i ++ {
		x1 := new(big.Int).Mul(Fx, Fx)
		x1.Mod(x1, P)
		x2 := new(big.Int).Mul(Gx, Gx)
		x2.Mod(x2, P)
		x3 := new(big.Int).Mul(Fx, Gx)
		x3.Mod(x3, P)
		x4 := new(big.Int).Add(Fy, Gy)
		x4.ModInverse(x4, P)
		t := new(big.Int).Set(x1)
		t.Add(t, x2)
		t.Add(t, x3)
		t.Mul(t, x4)
		t.Mod(t, P)
		tSquare := new(big.Int).Mul(t, t)
		tSquare.Mod(tSquare, P)
		xSum := new(big.Int).Add(Fx, Gx)
		u := new(big.Int).Sub(tSquare, xSum)
		u.Mod(u, P)
		v := new(big.Int).Sub(Fx, u)
		v.Mul(v, t)
		v.Sub(v, Fy)
		v.Mod(v, P)
		Fx.Set(u)
		Fy.Set(v)
	}
	return &common.Point{X: Fx.Bytes(), Y:Fy.Bytes()}
}

func GetRandomKQ(curve Curve) ([]byte, *common.Point, error) {
	mask := []byte{0xff, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f}
	N := curve.Params().N
	BitSize := curve.Params().BitSize
	byteLen := (BitSize + 7) >> 3
	ok := false
	try := 0
	var k []byte
	for !ok {
	    if try > MaxRandomRetry {
	        break
       }
		k = make([]byte, byteLen)
		if _, err := rand.Read(k); err != nil {
		    try += 1
			continue
		}
		k[0] &= mask[BitSize % 8]
		kInt := new(big.Int).SetBytes(k)
		if kInt.Cmp(Zero) > 0 || kInt.Cmp(N) < 0 {
			ok = true
		} else {
		    try += 1
       }
	}
	if ok {
	    return k, curve.ScalarBaseMultiply(k), nil
   } else {
       return nil, nil, errors.New("random fail")
   }
}

func ConcatHash(Q, pubKey *common.Point, msg []byte) []byte {
    concat := make([]byte, 4 * ByteLen + len(msg))
    copy(concat[:2 * ByteLen], Q.Bytes())
    copy(concat[2 * ByteLen:], pubKey.Bytes())
    copy(concat[4 * ByteLen:], msg)
	hash := Hash256(concat)
	return hash
}

func GenerateKey(curve Curve) (*common.PrivateKey, error) {
	prv, pubKey, err := GetRandomKQ(curve)
	if err != nil {
		return nil, err
	}
	prvKey := &common.PrivateKey{Prv: prv, PubKey: pubKey}
	return prvKey, nil
}

func Sign(curve Curve, prvKey *common.PrivateKey, msg []byte) (*common.Signature, error) {
	r, s := new(big.Int), new(big.Int)
	prvInt := new(big.Int).SetBytes(prvKey.Prv)
	for r.Cmp(Zero) == 0 || s.Cmp(Zero) == 0 {
		k, Q, err := GetRandomKQ(curve)
		if err != nil {
			return nil, err
		}
		N := curve.Params().N
		r = new(big.Int).SetBytes(ConcatHash(Q, prvKey.PubKey, msg))
		r.Mod(r, N)
		s = new(big.Int).Mul(r, prvInt)
		s.Mod(s, N)
		s.Sub(new(big.Int).SetBytes(k), s)
		s.Mod(s, N)
	}
	sig := &common.Signature{}
	sig.R = r.Bytes()
	sig.S = s.Bytes()
	return sig, nil
}

func Verify(curve Curve, sig *common.Signature, pubKey *common.Point, msg []byte) bool {
	r, s := new(big.Int).SetBytes(sig.R), new(big.Int).SetBytes(sig.S)
	if r.Cmp(Zero) <= 0 || r.Cmp(curve.Params().N) >= 0 || s.Cmp(Zero) <=0 || s.Cmp(curve.Params().N) >=0 {
		return false
	}
	N := curve.Params().N
	sG := curve.ScalarBaseMultiply(sig.S)
	rP := curve.ScalarMultiply(pubKey, sig.R)
	Q:= curve.Add(sG, rP)
	Qx, Qy := new(big.Int).SetBytes(Q.X), new(big.Int).SetBytes(Q.Y)
	if Qx.Cmp(Zero) == 0 && Qy.Cmp(Zero) == 0 {
		return false
	}
	v := new(big.Int).SetBytes(ConcatHash(Q, pubKey, msg))
	v.Mod(v, N)
	if v.Cmp(r) == 0{
		return true
	} else {
		return false
	}
}

func Encrypt(curve Curve, pubKey *common.Point, msg []byte) ([]byte, error) {
	k, p1, err := GetRandomKQ(curve)
	if err != nil {
	    return nil, err
    }
	c1 := p1.Bytes()
	p2 := curve.ScalarMultiply(pubKey, k)
	j2 := p2.Bytes()
	t := new(big.Int).SetBytes(KDF(j2))
	m := new(big.Int).SetBytes(msg)
	c2 := new(big.Int).Xor(m, t).Bytes()
	b := make([]byte, len(j2) + len(msg))
	copy(b[:len(j2)], j2)
	copy(b[len(j2):], msg)
	c3 := Hash256(b)
	cipher := make([]byte, 3 * ByteLen + len(c2))
	copy(cipher[: 2 * ByteLen], c1)
	copy(cipher[2 * ByteLen: 3 * ByteLen], c3)
	copy(cipher[3 * ByteLen:], c2)
	return cipher, nil
}

func Decrypt(curve Curve, prvKey *common.PrivateKey, cipher []byte) ([]byte, error) {
	p1 := &common.Point{X: cipher[:32], Y: cipher[32: 64]}
	if !curve.IsOnCurve(p1) {
		return nil, errors.New("point not on curve")
	}
	p2 := curve.ScalarMultiply(p1, prvKey.Prv)
	j2 := p2.Bytes()
	t := new(big.Int).SetBytes(KDF(j2))
	c2 := cipher[3 * ByteLen:]
	c := new(big.Int).SetBytes(c2)
	m := new(big.Int).Xor(c, t)
	msg := m.Bytes()
	b := make([]byte, len(j2) + len(msg))
	copy(b[:len(j2)], j2)
	copy(b[len(j2):], msg)
	u := Hash256(b)
	c3 := cipher[2 * ByteLen: 3 * ByteLen]
	if bytes.Equal(u, c3) {
		return msg, nil
	} else {
		return nil, errors.New("cipher wrong")
	}
}

func InitCurve() (curveParams *CurveParams) {
	curveParams = &CurveParams{}
	curveParams.P = new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFC, 0x2F})
	curveParams.N = new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFE, 0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B, 0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x41})
	curveParams.B = new(big.Int).SetBytes([]byte{0x07})
	Gx := []byte{0x79, 0xBE, 0x66, 0x7E, 0xF9, 0xDC, 0xBB, 0xAC, 0x55, 0xA0, 0x62, 0x95, 0xCE, 0x87, 0x0B, 0x07, 0x02, 0x9B, 0xFC,
					0xDB, 0x2D, 0xCE, 0x28, 0xD9, 0x59, 0xF2, 0x81, 0x5B, 0x16, 0xF8, 0x17, 0x98}
	Gy := []byte{0x48, 0x3A, 0xDA, 0x77, 0x26, 0xA3, 0xC4, 0x65, 0x5D, 0xA4, 0xFB, 0xFC, 0x0E, 0x11, 0x08, 0xA8, 0xFD, 0x17, 0xB4,
					0x48, 0xA6, 0x85, 0x54, 0x19, 0x9C, 0x47, 0xD0, 0x8F, 0xFB, 0x10, 0xD4, 0xB8}
	curveParams.G = &common.Point{X: Gx, Y: Gy}
	curveParams.BitSize = 256
	curveParams.Name = "Secp256-k1"
	return
}

func (p *common.Point) Bytes() []byte {
    j := make([]byte, 2 * ByteLen)
    copy(j[ByteLen - len(p.X): ByteLen], p.X)
    copy(j[2 * ByteLen - len(p.Y):], p.Y)
    return j
}

func (p *common.Point) Equal(q *common.Point) bool {
    if !bytes.Equal(p.X, q.X) {
        return false
    }
    if !bytes.Equal(p.Y, q.Y) {
        return false
    }
    return true
}