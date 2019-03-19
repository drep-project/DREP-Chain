package types

type ConsensusConfig struct {
	ConsensusMode 	string          	`json:"consensusMode"`
	Me 	 			string 				`json:"me"`
	EnableConsensus bool				`json:"enableConsensus"`
}