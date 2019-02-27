package service

type ConsensusApi struct {
	consensusService *ConsensusService
}

//TODO mock a rpc to provent rpc error
func (consensusApi *ConsensusApi) Mock(){

}

func (consensusApi *ConsensusApi) Minning() bool {
	switch consensusApi.consensusService.Config.ConsensusMode {
	case "solo":
		return true
	case "bft":
		return true
	default:
		return false
	}
}

