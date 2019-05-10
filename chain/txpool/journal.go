package txpool

import (
	"bufio"
	systemBinary "encoding/binary"
	"errors"
	"github.com/drep-project/binary"
	log "github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"io"
	"os"
	"path"
)

// errNoActiveJournal is returned if a transaction is attempted to be inserted
// into the journal, but no such file is currently open.
var errNoActiveJournal = errors.New("no active journal")

// txJournal is a rotating log of transactions with the aim of storing locally
// created transactions to allow non-executed ones to survive node restarts.
type txJournal struct {
	path   string         // Filesystem path to store the transactions at
	writer io.WriteCloser // Output stream to write new transactions into
}

// newTxJournal creates a new transaction journal to
func newTxJournal(path string) *txJournal {
	return &txJournal{
		path: path,
	}
}

// load parses a transaction journal dump from disk, loading its contents into
// the specified pool.
func (journal *txJournal) load(add func([]types.Transaction) []error) error {
	// Skip the parsing if the journal file doesn't exist at all
	if _, err := os.Stat(journal.path); os.IsNotExist(err) {
		return nil
	}
	// Open the journal for loading any past transactions
	input, err := os.Open(journal.path)
	if err != nil {
		return err
	}
	defer input.Close()

	// Inject all transactions from the journal into the pool
	//stream := journal.NewStream(input, 0)
	total, dropped := 0, 0

	// Create a method to load a limited batch of transactions and bump the
	// appropriate progress counters. Then use this method to load all the
	// journaled transactions in small-ish batches.
	loadBatch := func(txs types.Transactions) {
		for _, err := range add(txs) {
			if err != nil {
				log.Debug("Failed to add journaled transaction", "err", err)
				dropped++
			}
		}
	}

	var (
		failure error
		batch   types.Transactions
	)

	reader := bufio.NewReader(input)
	for {
		// Parse the next transaction and terminate on error
		tx := new(types.Transaction)
		if err = journal.Decode(reader, tx); err != nil {
			if err != io.EOF {
				failure = err
			}
			if len(batch) > 0 {
				loadBatch(batch)
			}
			break
		}
		// New transaction parsed, queue up for later, import if threshold is reached
		total++

		if batch = append(batch, *tx); len(batch) > 1024 {
			loadBatch(batch)
			batch = batch[:0]
		}
	}
	log.Info("Loaded local transaction journal", "transactions", total, "dropped", dropped)

	return failure
}

// insert adds the specified transaction to the local disk journal.
func (journal *txJournal) insert(tx *types.Transaction) error {
	if journal.writer == nil {
		return errNoActiveJournal
	}
	return journal.Encode(journal.writer, tx)
}

// rotate regenerates the transaction journal based on the current contents of
// the transaction pool.
func (journal *txJournal) rotate(all map[crypto.CommonAddress][]*types.Transaction) error {
	// Close the current journal (if any is open)
	if journal.writer != nil {
		if err := journal.writer.Close(); err != nil {
			return err
		}
		journal.writer = nil
	}

	if _, err := os.Stat(journal.path); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(journal.path), 0755); err != nil {
			return err
		}
	}

	// Generate a new journal with the contents of the current pool
	replacement, err := os.OpenFile(journal.path+".new", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	journaled := 0
	for _, txs := range all {

		for _, tx := range txs {
			if err = journal.Encode(replacement, tx); err != nil {
				replacement.Close()
				return err
			}
		}
		//fmt.Println("rotate max nonce ", txs[len(txs)-1].Nonce(), len(txs))
		journaled += len(txs)
	}
	replacement.Close()

	// Replace the live journal with the newly generated one
	if err = os.Rename(journal.path+".new", journal.path); err != nil {
		return err
	}
	sink, err := os.OpenFile(journal.path, os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	journal.writer = sink
	log.Debug("Regenerated local transaction journal", "transactions", journaled, "accounts", len(all))

	return nil
}

// close flushes the transaction journal contents to disk and closes the file.
func (journal *txJournal) close() error {
	var err error

	if journal.writer != nil {
		err = journal.writer.Close()
		journal.writer = nil
	}
	return err
}

func (journal *txJournal) Encode(w io.WriteCloser, v interface{}) ( error) {
	bufTx, _ := binary.Marshal(v)
	txLen := len(bufTx)
	buf := make([]byte, txLen, txLen+8)

	n := systemBinary.PutVarint(buf, int64(txLen))
	buf = append(buf[:n], bufTx...)
	_, err := w.Write(buf)
	return err
}

func (journal *txJournal) Decode(r *bufio.Reader, v interface{}) error {
	n, err := systemBinary.ReadVarint(r)
	if err != nil {
		return err
	}

	buf := make([]byte, n)
	txLen, err := r.Read(buf)
	if err != nil {
		return err
	}

	if n != int64(txLen) {
		return errors.New("maybe file damaged")
	}

	return binary.Unmarshal(buf, v)
}
