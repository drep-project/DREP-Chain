package common

type SetUp1Message struct {
    BlockHeight int
    Block *Block
    PubKey []byte
    Sig []byte
}

type Block1CommitMessage struct {
    Q []byte
    PubKey []byte
}

type Block1ChallengeMessage struct {
    R []byte
    Q []byte
    PubKey []byte
}

type Block1ResponseMessage struct {
    S []byte
    PubKey []byte
}

type SetUp2Message struct {
    PubKey []byte
    Sig []byte
}

type Block2CommitMessage struct {
    // TODO Message part should be added
    Q []byte
    PubKey []byte
}

type Block2ChallengeMessage struct {
    R []byte
    Q []byte
    PubKey []byte
}

type Block2ResponseMessage struct {
    S []byte
    PubKey []byte
}
