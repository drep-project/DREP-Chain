package bft

import (
	"container/heap"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"math/big"
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

func GetCandidates(store store.StoreInterface, topN int) []*Producer {
	candidateAddrs, err := store.GetCandidateAddrs()
	if err != nil {
		log.Errorf("get candidates err:%v", err)
		return nil
	}

	csh := make(creditsHeap, 0)
	for _, addr := range candidateAddrs {
		addr := addr
		totalCredit := store.GetVoteCreditCount(&addr)
		csh = append(csh, &addrAndCredit{addr: &addr, value: totalCredit})
	}

	heap.Init(&csh)

	producerAddrs := []*Producer{}

	addNum := 0
	for csh.Len() > 0 {
		v := heap.Pop(&csh)
		ac := v.(*addrAndCredit)

		data, err := store.GetCandidateData(ac.addr)
		if err != nil {
			log.WithField("err", err).Info("get candidate data err")
			continue
		}

		cd := &types.CandidateData{}
		err = binary.Unmarshal(data, cd)
		if err != nil {
			log.WithField("err", err).Info("unmarshal data to candidateData err")
			continue
		}

		log.Trace("get candidates info:", cd.Node, string(cd.Pubkey.Serialize()), ac.addr.String())
		n := &enode.Node{}
		err = n.UnmarshalText([]byte(cd.Node))
		if err != nil {
			continue
		}
		producer := &Producer{
			Pubkey: cd.Pubkey,
			Node:   n,
		}
		producerAddrs = append(producerAddrs, producer)
		addNum++
		if addNum == topN {
			return producerAddrs
		}
	}
	return producerAddrs
}
