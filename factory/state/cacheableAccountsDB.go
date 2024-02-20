package state

import (
	"sync"

	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("CacheableAccountsDB")

type CacheableAccountsDB struct {
	state.AccountsAdapter
	Cache    map[string]map[string]vmcommon.AccountHandler
	mutCache sync.RWMutex
}

func (cadb *CacheableAccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	// log.Info("GetExistingAccount")
	cadb.mutCache.RLock()
	currentMap, ok := cadb.Cache[string(address)]
	cadb.mutCache.RUnlock()

	if !ok {
		cadb.mutCache.Lock()
		defer cadb.mutCache.Unlock()
		currentMap = make(map[string]vmcommon.AccountHandler)
		cadb.Cache[string(address)] = currentMap
	}

	account, ok := currentMap[string(address)]
	if ok {
		return account, nil
	}

	account, err := cadb.AccountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	currentMap[string(address)] = account
	return account, nil
}

func (cadb *CacheableAccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	//log.Info("LoadAccount")
	cadb.mutCache.RLock()
	currentMap, ok := cadb.Cache[string(address)]
	//log.Info("LoadAccount", "currentMap", currentMap, "ok", ok)
	cadb.mutCache.RUnlock()

	if !ok {
		cadb.mutCache.Lock()
		defer cadb.mutCache.Unlock()
		currentMap = make(map[string]vmcommon.AccountHandler)
		cadb.Cache[string(address)] = currentMap
	}

	if ok {
		account, ok := currentMap[string(address)]
		//	log.Info("LoadAccount", "account", currentMap, "ok", ok)
		if ok {
			return account, nil
		}

	}

	account, err := cadb.AccountsAdapter.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	//	log.Info("LoadAccount", "setting currentMap with account - started", address, "account", account)
	currentMap[string(address)] = account
	//	log.Info("LoadAccount", "setting currentMap with account - finished", address, "account", account)
	return account, nil
}

func (cadb *CacheableAccountsDB) SaveAccount(account vmcommon.AccountHandler) error {
	//	log.Info("SaveAccount")
	userAccount, ok := account.(state.BaseAccountHandler)
	if ok {
		hasCode := len(userAccount.GetCode()) > 0
		if hasCode {
			err := cadb.AccountsAdapter.SaveAccount(account)
			if err != nil {
				return err
			}
		}
	}

	cadb.mutCache.RLock()
	currentMap, ok := cadb.Cache[string(account.AddressBytes())]
	cadb.mutCache.RUnlock()

	if !ok {
		cadb.mutCache.Lock()
		defer cadb.mutCache.Unlock()
		currentMap = make(map[string]vmcommon.AccountHandler)
		cadb.Cache[string(account.AddressBytes())] = currentMap
	}

	currentMap[string(account.AddressBytes())] = account
	return nil
}

func (cadb *CacheableAccountsDB) Commit() ([]byte, error) {
	//	log.Info("Commit")
	cadb.mutCache.Lock()
	defer cadb.mutCache.Unlock()

	for _, accMaps := range cadb.Cache {
		for _, account := range accMaps {
			err := cadb.AccountsAdapter.SaveAccount(account)
			if err != nil {
				return nil, err
			}
		}
	}

	cadb.Cache = make(map[string]map[string]vmcommon.AccountHandler)
	return cadb.AccountsAdapter.Commit()
}
