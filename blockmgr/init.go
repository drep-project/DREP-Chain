package blockmgr

import (
	dlog "github.com/drep-project/DREP-Chain/pkgs/log"
)

const (
	maxHeaderHashCountReq = 48  //The maximum number of header hash requests
	maxBlockCountReq      = 8   //The maximum number of header hash requests
	maxSyncSleepTime      = 200 //During the synchronization process, each cycle rests for 200 milliseconds
	maxNetworkTimeout     = 30  //Maximum network timeout
	maxLivePeer           = 50
	broadcastRatio        = 3    //BroadcastRatio broadcasts one third as many non-local messages
	maxTxsCount           = 1024 //The maximum number of transmission transactions
	pendingTimerCount     = 2    //When synchronizing blocks, the maximum number of concurrent coroutines of fetch block requests

	MODULENAME = "blockmgr"
)

var (
	log = dlog.EnsureLogger(MODULENAME)
)
