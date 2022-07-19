package executor

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/icon-project/icon-bridge/cmd/e2etest/chain"
	"github.com/icon-project/icon-bridge/cmd/e2etest/chain/bsc"
	"github.com/icon-project/icon-bridge/cmd/e2etest/chain/hmny"
	"github.com/icon-project/icon-bridge/cmd/e2etest/chain/icon"
	"github.com/icon-project/icon-bridge/common/errors"
	"github.com/icon-project/icon-bridge/common/log"
)

type executor struct {
	log             log.Logger
	godKeysPerChain map[chain.ChainType][2]string
	cfgPerChain     map[chain.ChainType]*chain.ChainConfig
	clientsPerChain map[chain.ChainType]chain.ChainAPI
	sinkChanPerID   map[uint64]chan *evt
	syncChanMtx     sync.RWMutex
	stoppedChan     chan struct{}
}

func New(l log.Logger, cfgPerChain map[chain.ChainType]*chain.ChainConfig) (ex *executor, err error) {
	getKeyPairFromFile := func(walFile string, password string) (pair [2]string, err error) {
		keyReader, err := os.Open(walFile)
		if err != nil {
			err = errors.Wrapf(err, "os.Open file %v", walFile)
			return
		}
		defer keyReader.Close()

		keyStore, err := ioutil.ReadAll(keyReader)
		if err != nil {
			err = errors.Wrapf(err, "ioutil.ReadAll %v", walFile)
			return
		}
		key, err := keystore.DecryptKey(keyStore, password)
		if err != nil {
			err = errors.Wrapf(err, "keystore.Decrypt %v", walFile)
			return
		}
		privBytes := ethcrypto.FromECDSA(key.PrivateKey)
		privString := hex.EncodeToString(privBytes)
		addr := ethcrypto.PubkeyToAddress(key.PrivateKey.PublicKey)
		pair = [2]string{privString, addr.String()}
		return
	}
	ex = &executor{
		log:             l,
		cfgPerChain:     cfgPerChain,
		godKeysPerChain: make(map[chain.ChainType][2]string),
		clientsPerChain: make(map[chain.ChainType]chain.ChainAPI),
		sinkChanPerID:   make(map[uint64]chan *evt),
		syncChanMtx:     sync.RWMutex{},
		stoppedChan:     make(chan struct{}),
	}
	for name, cfg := range cfgPerChain {
		// GodKeys
		if pair, err := getKeyPairFromFile(cfg.GodWallet.Path, cfg.GodWallet.Password); err != nil {
			return nil, errors.Wrapf(err, "getKeyPairFromFile(%v)", cfg.GodWallet.Path)
		} else {
			ex.godKeysPerChain[name] = pair
		}
		//Clients
		if name == chain.HMNY {
			ex.clientsPerChain[name], err = hmny.NewApi(l, cfg)
			if err != nil {
				err = errors.Wrap(err, "hmny.NewApi ")
				return nil, err
			}
		} else if name == chain.ICON {
			ex.clientsPerChain[name], err = icon.NewApi(l, cfg)
			if err != nil {
				err = errors.Wrap(err, "icon.NewApi ")
				return nil, err
			}
		} else if name == chain.BSC {
			ex.clientsPerChain[name], err = bsc.NewApi(l, cfg)
			if err != nil {
				err = errors.Wrap(err, "bsc.NewApi ")
				return nil, err
			}
		} else {
			return nil, errors.New("Unknown Chain Type supplied from config: " + string(name))
		}
	}
	return
}

func (ex *executor) Done() <-chan struct{} {
	return ex.stoppedChan
}

func (ex *executor) getID() (uint64, error) {
	ex.syncChanMtx.RLock()
	defer ex.syncChanMtx.RUnlock()
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ { // Try 5 times to get a random number not already used
		id := uint64(rand.Intn(10000)) // human readable range
		if _, ok := ex.sinkChanPerID[id]; !ok {
			return id, nil
		}
	}
	for i := 0; i < 2; i++ {
		id := rand.Uint64() // larger search space
		if _, ok := ex.sinkChanPerID[id]; !ok {
			return id, nil
		}
	}
	return 0, errors.New("Couldn't find a unique ID")
}

func (ex *executor) addChan(id uint64, ch chan *evt) {
	ex.syncChanMtx.Lock()
	defer ex.syncChanMtx.Unlock()
	ex.sinkChanPerID[id] = ch
}

func (ex *executor) removeChan(id uint64) {
	ex.syncChanMtx.Lock()
	defer ex.syncChanMtx.Unlock()
	if ch, ok := ex.sinkChanPerID[id]; ok {
		//ex.log.Debugf("Cleaning up after function returns. Removing channel of id %v", id)
		if ch != nil {
			close(ex.sinkChanPerID[id])
		}
		delete(ex.sinkChanPerID, id)
	}
	if len(ex.sinkChanPerID) == 0 {
		ex.log.Info("All test scripts have been completed")
		ex.stoppedChan <- struct{}{}
	}
}

func (ex *executor) getChan(id uint64) chan *evt {
	ex.syncChanMtx.RLock()
	defer ex.syncChanMtx.RUnlock()
	if _, ok := ex.sinkChanPerID[id]; ok {
		return ex.sinkChanPerID[id]
	} else {
		ex.log.Warnf("Message Target id %v does not exist", id)
	}
	return nil
}

func (ex *executor) Subscribe(ctx context.Context) {
	go func() {
		lenCls := len(ex.clientsPerChain)
		chains := make([]chain.ChainType, lenCls)
		cases := make([]reflect.SelectCase, 1+(lenCls*2))
		i := 0
		for name, cl := range ex.clientsPerChain {
			ex.log.Debugf("Start Subscription %v", name)
			sinkChan, errChan, err := cl.Subscribe(ctx)
			if err != nil {
				ex.log.Error(errors.Wrapf(err, "%v: Subscribe()", name))
			}
			chains[i] = name
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(sinkChan)}
			cases[i+lenCls] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(errChan)}
			i++
		}
		cases[len(cases)-1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
		for {
			chosen, value, ok := reflect.Select(cases)
			if !ok {
				if chosen == len(cases)-1 {
					ex.log.Error("Context cancelled. Exiting Executor")
					return
				}
				if chosen >= lenCls {
					ex.log.Debugf("Sender Closed ErrMessage ChannelID %v Client %v", chosen, chains[chosen-lenCls])
				} else {
					ex.log.Debugf("Sender Closed EvtMessage ChannelID %v Client %v", chosen, chains[chosen])
				}
				cases[chosen].Chan = reflect.ValueOf(nil)
				continue
			}

			if chosen < lenCls { // [0, lenCapi-1] is message
				res, dok := value.Interface().(*chain.EventLogInfo)
				if !dok {
					ex.log.Errorf("Got interface of type %T; Expected errorType", value)
					break
				}
				if len(res.IDs) > 0 {
					for _, id := range res.IDs {
						if dst := ex.getChan(id); dst != nil {
							dst <- &evt{chainType: chains[chosen], msg: res}
						}
					}
				} else {
					ex.log.Warnf("Message without target received %+v", res)
				}
			} else if chosen >= lenCls && chosen < 2*len(cases) {
				res, eok := value.Interface().(error)
				if !eok {
					ex.log.Errorf("Got interface of type %T; Expected errorType", value)
					break
				}
				ex.log.Errorf("ErrMessage %v %+v", chains[chosen-lenCls], res)
			} else {
				ex.log.Error("Context cancelled. Exiting Executor")
				return
			}
		}
	}()
}

func (ex *executor) Execute(ctx context.Context, srcChainName, dstChainName chain.ChainType, coinName string, scr Script) (err error) {
	id, err := ex.getID()
	if err != nil {
		return errors.Wrap(err, "getID ")
	}
	log := ex.log.WithFields(log.Fields{"pid": id})
	sinkChan := make(chan *evt)
	ex.addChan(id, sinkChan)
	defer ex.removeChan(id) // should defer be called by cb() instead to make sure cb() was done

	reqClients := map[chain.ChainType]chain.ChainAPI{}
	for k, v := range ex.clientsPerChain {
		reqClients[k] = v
	}
	reqGodKeys := map[chain.ChainType]keypair{}
	for k, v := range ex.godKeysPerChain {
		reqGodKeys[k] = keypair{PrivKey: v[0], PubKey: v[1]}
	}
	ts := &testSuite{
		id:              id,
		logger:          log,
		subChan:         sinkChan,
		clsPerChain:     reqClients,
		godKeysPerChain: reqGodKeys,
	}

	ex.log.Infof("Run ID %v %v, Transfer %v From %v To %v", id, scr.Name, coinName, srcChainName, dstChainName)
	if scr.Callback == nil {
		return errors.New("Callback function was nil")
	}
	if err := scr.Callback(ctx, srcChainName, dstChainName, coinName, ts); err != nil {
		return errors.Wrap(err, "CallBackFunc ")
	}
	ex.log.Infof("Completed Succesfully. ID %v %v, Transfer %v From %v To %v", id, scr.Name, coinName, srcChainName, dstChainName)
	// CleanupFunc removeChan() is called after cb() on function return
	// so make sure cb() returns only after all the test logic is finished
	return
}