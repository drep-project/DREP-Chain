package store

var (
    forwardedTrans = make(map[string]bool)
    forwardedBlocks = make(map[string]bool)
)

func ForwardTransaction(id string) {
    forwardedTrans[id] = true
}

func ForwardedTransaction(id string) bool {
    v, exists := forwardedTrans[id]
    //TODO
    //fatal error: concurrent map read and map write
    //
    //goroutine 37 [running]:
    //runtime.throw(0x4923a26, 0x21)
    ///Users/mac/go/src/runtime/panic.go:608 +0x72 fp=0xc000066de0 sp=0xc000066db0 pc=0x402e6c2
    //runtime.mapaccess2_faststr(0x4815640, 0xc00012f830, 0xc00b020240, 0x40, 0x0, 0x0)
    ///Users/mac/go/src/runtime/map_faststr.go:110 +0x458 fp=0xc000066e50 sp=0xc000066de0 pc=0x4014c08
    //BlockChainTest/store.ForwardedTransaction(...)
    ///Users/mac/Documents/go/src/BlockChainTest/store/record.go:13
    //BlockChainTest/processor.(*transactionProcessor).process(0x5134090, 0xc0010b77c0, 0x486cd60, 0xc000b59e60)
    ///Users/mac/Documents/go/src/BlockChainTest/processor/chain_processor.go:22 +0xe0 fp=0xc000066f40 sp=0xc000066e50 pc=0x474e5e0
    //BlockChainTest/processor.(*Processor).dispatch(0xc000132de0, 0xc0010b7860)
    ///Users/mac/Documents/go/src/BlockChainTest/processor/processor.go:71 +0x91 fp=0xc000066fa8 sp=0xc000066f40 pc=0x474f801
    //BlockChainTest/processor.(*Processor).Start.func1(0xc000132de0)
    ///Users/mac/Documents/go/src/BlockChainTest/processor/processor.go:56 +0x64 fp=0xc000066fd8 sp=0xc000066fa8 pc=0x474f994
    //runtime.goexit()
    ///Users/mac/go/src/runtime/asm_amd64.s:1333 +0x1 fp=0xc000066fe0 sp=0xc000066fd8 pc=0x405d001
    //created by BlockChainTest/processor.(*Processor).Start
    ///Users/mac/Documents/go/src/BlockChainTest/processor/processor.go:53 +0x3f
    return v && exists
}

func ForwardBlock(id string) {
    forwardedBlocks[id] = true
}

func ForwardedBlock(id string) bool {
    // TODO first check db and second check the pool
    v, exists := forwardedBlocks[id]
    return v && exists
}