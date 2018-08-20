package network

import (
	"math/big"
	"crypto/rand"
)

var zero = new(big.Int)

type Curve interface {
	Params() *CurveParams
	IsOnCurve(*Point) bool
	Add(*Point, *Point) *Point
	Double(*Point) *Point
	ScalarMultiply(*Point, []byte) *Point
	ScalarBaseMultiply([]byte) *Point
}

// Y^2 == X^3 + AX + B (mod p), with a == 0
type CurveParams struct {
	P *big.Int
	N *big.Int
	B *big.Int
	G *Point
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
func (curveParams *CurveParams) IsOnCurve(point *Point) bool {
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

func (curveParams *CurveParams) Negative(point *Point) *Point {
    y := new(big.Int).SetBytes(point.Y)
    y.Sub(curveParams.P, y)
    return &Point{X: point.X, Y: y.Bytes()}
}

func JacobiAffine(point *Point) *JacobiCoordinate {
	x, y, z := new(big.Int).SetBytes(point.X), new(big.Int).SetBytes(point.Y), new(big.Int)
	if x.Sign() != 0 || y.Sign() != 0 {
		z.SetInt64(1)
	}
	return &JacobiCoordinate{x, y, z}
}

func (curveParams *CurveParams) InverseJacobiAffine(jc *JacobiCoordinate) *Point {
	x, y, z := jc.X, jc.Y, jc.Z
	if z.Sign() == 0 {
		return &Point{X: new(big.Int).Bytes(), Y: new(big.Int).Bytes()}
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
	return &Point{X: xOut.Bytes(), Y: yOut.Bytes()}
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

func (curveParams *CurveParams) Add(pt1, pt2 *Point) *Point {
	jc1 := JacobiAffine(pt1)
	jc2 := JacobiAffine(pt2)
	return curveParams.InverseJacobiAffine(curveParams.JacobiAddition(jc1, jc2))
}

func (curveParams *CurveParams) Double(point *Point) *Point {
	jc := JacobiAffine(point)
	return curveParams.InverseJacobiAffine(curveParams.JacobiDoubling(jc))
}

func (curveParams *CurveParams) ScalarMultiply(point *Point, k []byte) *Point {
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

func (curveParams *CurveParams) ScalarBaseMultiply(k []byte) *Point {
	return curveParams.ScalarMultiply(curveParams.G, k)
}

func (curveParams *CurveParams) ScalarBaseMultiplyByFormula(k int) *Point {
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
	return &Point{X: Fx.Bytes(), Y:Fy.Bytes()}
}

func RandomSample(curve Curve) (k []byte, err error) {
	mask := []byte{0xff, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f}
	N := curve.Params().N
	BitSize := curve.Params().BitSize
	byteLen := (BitSize + 7) >> 3
	ok := false
	for !ok {
		k = make([]byte, byteLen)
		if _, err = rand.Read(k); err != nil {
			return
		}
		k[0] &= mask[BitSize % 8]
		kInt := new(big.Int).SetBytes(k)
		if kInt.Cmp(zero) > 0 || kInt.Cmp(N) < 0 {
			ok = true
		}
	}
	return
}

func ConcatEncode(Q, pubKey *Point, msg []byte) []byte {
	QxBytes := Q.X;
	plainText := make([]byte, len(QxBytes))
	plainText = append(plainText, Q.Y...)
	plainText = append(plainText, pubKey.X...)
	plainText = append(plainText, pubKey.Y...)
	plainText = append(plainText, msg...)
	cipherText := HashEnc(plainText)
	return cipherText
}

func GenerateKey(curve Curve) (prvKey *PrivateKey, err error) {
	prv, err := RandomSample(curve)
	if err != nil {
		return
	}
	pubKey := curve.ScalarBaseMultiply(prv)
	prvKey = &PrivateKey{Prv: prv, PubKey: pubKey}
	return
}

func Sign(curve Curve, prvKey *PrivateKey, msg []byte) (*Signature, error) {
	r, s := new(big.Int), new(big.Int)
	prvInt := new(big.Int).SetBytes(prvKey.Prv)
	for r.Cmp(zero) == 0 || s.Cmp(zero) == 0 {
		k, err := RandomSample(curve)
		if err != nil {
			return nil, err
		}
		N := curve.Params().N
		Q := curve.ScalarBaseMultiply(k)
		r = new(big.Int).SetBytes(ConcatEncode(Q, prvKey.PubKey, msg))
		r.Mod(r, N)
		s = new(big.Int).Mul(r, prvInt)
		s.Mod(s, N)
		s.Sub(new(big.Int).SetBytes(k), s)
		s.Mod(s, N)
	}
	sig := &Signature{}
	sig.R = r.Bytes()
	sig.S = s.Bytes()
	return sig, nil
}

func Verify(curve Curve, sig *Signature, pubKey *Point, msg []byte) bool {
	r, s := new(big.Int).SetBytes(sig.R), new(big.Int).SetBytes(sig.S)
	if r.Cmp(zero) <= 0 || r.Cmp(curve.Params().N) >= 0 || s.Cmp(zero) <=0 || s.Cmp(curve.Params().N) >=0 {
		return false
	}
	N := curve.Params().N
	sG := curve.ScalarBaseMultiply(sig.S)
	rP := curve.ScalarMultiply(pubKey, sig.R)
	Q:= curve.Add(sG, rP)
	Qx, Qy := new(big.Int).SetBytes(Q.X), new(big.Int).SetBytes(Q.Y)
	if Qx.Cmp(zero) == 0 && Qy.Cmp(zero) == 0 {
		return false
	}
	v := new(big.Int).SetBytes(ConcatEncode(Q, pubKey, msg))
	v.Mod(v, N)
	if v.Cmp(r) == 0{
		return true
	} else {
		return false
	}
}

//func BytesToPoint(curve Curve, msg []byte) *Point {
//
//}
//
//func Encrypt(curve Curve, plaintext []byte, pubKey *Point) ([]byte, error) {
//	k, err := RandomSample(curve)
//	if err != nil {
//		return nil, err
//	}
//}

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
	curveParams.G = &Point{X: Gx, Y: Gy}
	curveParams.BitSize = 256
	curveParams.Name = "Secp256-k1"
	return
}

func InitCurveByString() (curveParams *CurveParams) {
	curveParams = &CurveParams{}
	curveParams.P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	curveParams.N, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
	curveParams.B = new(big.Int).SetBytes([]byte{0x07})
	Gx, _ := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	Gy, _ := new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
	curveParams.G = &Point{X: Gx.Bytes(), Y: Gy.Bytes()}
	curveParams.BitSize = 256
	curveParams.Name = "Secp256-k1"
	return
}

func InitMiniCurve() (curveParams *CurveParams) {
	curveParams = &CurveParams{}
	curveParams.P = new(big.Int).SetInt64(int64(71))
	curveParams.N = new(big.Int).SetInt64(int64(70))
	curveParams.B = new(big.Int).SetInt64(int64(7))
	Gx := new(big.Int).SetInt64(int64(2))
	Gy := new(big.Int).SetInt64(int64(21))
	curveParams.G = &Point{X: Gx.Bytes(), Y: Gy.Bytes()}
	curveParams.BitSize = 7
	curveParams.Name = "MiniCurve71"
	return
}

func PointEqual(p0, p1 *Point) bool {
	x0, y0 := new(big.Int).SetBytes(p0.X), new(big.Int).SetBytes(p0.Y)
	x1, y1 := new(big.Int).SetBytes(p1.X), new(big.Int).SetBytes(p1.Y)
	if x0.Cmp(x1) != 0 || y0.Cmp(y1) != 0 {
		return false
	}
	return true
}
