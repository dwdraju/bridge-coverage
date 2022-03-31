package hmny

import (
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/harmony-one/harmony/core/types"
	"github.com/icon-project/btp/common/errors"
)

var (
	bigOne  = big.NewInt(1)
	bigZero = big.NewInt(0)

	mdbPool = sync.Pool{
		New: func() interface{} { return memorydb.New() },
	}
)

func HexToAddress(s string) common.Address {
	return common.HexToAddress(s)
}

// mutates underlying byte slice
func reverseBytes(b []byte) []byte {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
	return b
}

func Retry(attempts int, sleep time.Duration, f func() error) error {
	err := f()
	if err != nil {
		if attempts--; attempts > 0 {
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, f)
		}
		return err
	}
	return nil
}

func receiptProof(receiptTrie *trie.Trie, key []byte) ([][]byte, error) {
	proofSet := light.NewNodeSet()
	err := receiptTrie.Prove(key, 0, proofSet)
	if err != nil {
		return nil, err
	}
	proofs := make([][]byte, 0)
	for _, node := range proofSet.NodeList() {
		proofs = append(proofs, node)
	}
	return proofs, nil
}

func receiptsTrie(receipts []*types.Receipt) (*trie.Trie, error) {
	mdb := mdbPool.Get().(*memorydb.Database)
	defer mdbPool.Put(mdb)

	tr, err := trie.New(common.Hash{}, trie.NewDatabase(mdb))
	if err != nil {
		return nil, errors.Wrapf(err, "trie.New: %v", err)
	}
	for i, r := range receipts {
		path, err := rlp.EncodeToBytes(uint64(i))
		if err != nil {
			return nil, err
		}
		rcpt, err := rlp.EncodeToBytes(r)
		if err != nil {
			return nil, err
		}
		tr.Update(path, rcpt)
	}
	_, err = tr.Commit(nil)
	if err != nil {
		return nil, err
	}
	return tr, nil
}