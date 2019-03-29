package lwasm

type SimpleGasPolicy struct {
	GasPerInstruction int64
}

func (p *SimpleGasPolicy) GetCost(key string) int64 {
	return p.GasPerInstruction
}
