package lwasm

const (
	CALL_FEE = 10

)
type SimpleGasPolicy struct {
	GasPerInstruction int64
}

func (p *SimpleGasPolicy) GetCost(key string) int64 {
	/*if key.Op == "call" {

	} else{

	}*/
	return p.GasPerInstruction
}
