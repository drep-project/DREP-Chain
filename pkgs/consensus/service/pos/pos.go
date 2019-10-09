package pos

import (
	"container/heap"
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/crypto"
	dlog "github.com/drep-project/drep-chain/pkgs/log"
	"math/big"
)

const (
	MODULENAME = "pos"
	BftBackboneNum = 7
)

var (
	log = dlog.EnsureLogger(MODULENAME)
)

type addrAndCredit struct {
	addr  *crypto.CommonAddress
	value *big.Int
}

type creditsHeap []*addrAndCredit

func (h creditsHeap) Len() int           { return len(h) }
func (h creditsHeap) Less(i, j int) bool { return h[i].value.Cmp(h[j].value) > 0 }
func (h creditsHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *creditsHeap) Push(x interface{}) {
	*h = append(*h, x.(*addrAndCredit))
}

func (h *creditsHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func GetCandidates(store store.StoreInterface) []*crypto.CommonAddress {
	voteAddrs, err := store.GetCandidateAddrs()
	if err != nil {
		log.Errorf("get candidates err:%v", err)
		return nil
	}

	csh := make(creditsHeap, 0)
	for addr, _ := range voteAddrs {
		addr := addr
		totalCredit := store.GetVoteCredit(&addr)
		csh = append(csh, &addrAndCredit{addr: &addr, value: totalCredit})
	}

	heap.Init(&csh)

	candidateAddrs := make([]*crypto.CommonAddress, 0)
	for i := 1; i <= BftBackboneNum; i++ {
		v := heap.Pop(&csh)
		ac := v.(*addrAndCredit)
		candidateAddrs = append(candidateAddrs, ac.addr)
	}

	return candidateAddrs
}
