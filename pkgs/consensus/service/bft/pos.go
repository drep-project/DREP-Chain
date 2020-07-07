package bft

import (
	"container/heap"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"math/big"
	"sort"
)

type addrsAndCredit struct {
	addrs []string //The addresses with the same value are put into a slice
	value *big.Int //The credit value corresponding to the address
}

type creditsHeap []*addrsAndCredit

func (h creditsHeap) Len() int           { return len(h) }
func (h creditsHeap) Less(i, j int) bool { return h[i].value.Cmp(h[j].value) > 0 }
func (h creditsHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *creditsHeap) Push(x interface{}) {
	*h = append(*h, x.(*addrsAndCredit))
}

func (h *creditsHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func GetCandidates(store store.StoreInterface, topN int) []types.Producer {
	candidateAddrs, err := store.GetCandidateAddrs()
	if err != nil || topN <= 0 {
		log.Errorf("topN:%d, get candidates err:%v", topN, err)
		return nil
	}

	//key for credit; value for addrs slice
	mapAddrs := make(map[string][]string)
	for _, addr := range candidateAddrs {
		addrStr := addr.String()
		totalCredit := store.GetVoteCreditCount(&addr).String()
		if addrs, ok := mapAddrs[totalCredit]; ok {
			addrs = append(addrs, addrStr)
			mapAddrs[totalCredit] = addrs
		} else {
			addrs := make([]string, 0)
			addrs = append(addrs, addrStr)
			mapAddrs[totalCredit] = addrs
		}

		log.WithField("addrs", addr.String()).WithField("credit", totalCredit).Trace("getCandidates")
	}

	csh := make(creditsHeap, 0)
	for k, v := range mapAddrs {
		ac := addrsAndCredit{}
		//排序addr
		sort.Strings(v)
		ac.addrs = v
		ac.value, _ = new(big.Int).SetString(k, 10)
		csh.Push(&ac)
	}
	heap.Init(&csh)

	producerAddrs := []types.Producer{}
	addNum := 0
	for csh.Len() > 0 {
		ac := heap.Pop(&csh).(*addrsAndCredit)
		for _, strAddr := range ac.addrs {
			addr := crypto.HexToAddress(strAddr)
			data, err := store.GetCandidateData(&addr)
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

			n := &enode.Node{}
			err = n.UnmarshalText([]byte(cd.Node))
			if err != nil {
				log.Errorf("get candidates err:%s,%v", err.Error(), cd.Node)
				continue
			}
			log.Trace("get candidates info:", cd.Node)
			producer := types.Producer{
				Pubkey: cd.Pubkey,
				Node:   n,
			}
			producerAddrs = append(producerAddrs, producer)
			addNum++
			if addNum >= topN {
				return producerAddrs
			}
		}
	}
	return producerAddrs
}
