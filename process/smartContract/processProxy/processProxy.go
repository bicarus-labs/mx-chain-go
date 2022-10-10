package processProxy

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/processorV2"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/scrCommon"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("processProxy")

var _ scrCommon.TestSmartContractProcessor = (*scProcessorProxy)(nil)

type configuredProcessor uint8

const (
	procV1 = iota
	procV2
)

type scProcessorProxy struct {
	configuredProcessor configuredProcessor
	args                scrCommon.ArgsNewSmartContractProcessor
	processor           process.SmartContractProcessorFacade
	processorsCache     map[configuredProcessor]process.SmartContractProcessorFacade
	testScProcessor     scrCommon.TestSmartContractProcessor
	testProcessorsCache map[configuredProcessor]scrCommon.TestSmartContractProcessor
	mutRc               sync.Mutex
	isTestVersion       bool
}

// NewSmartContractProcessorProxy creates a smart contract processor proxy
func NewSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier) (*scProcessorProxy, error) {
	return newSmartContractProcessorProxy(args, epochNotifier, false)
}

// NewTestSmartContractProcessorProxy creates a smart contract processor proxy
func NewTestSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier) (*scProcessorProxy, error) {
	return newSmartContractProcessorProxy(args, epochNotifier, true)
}

func newSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier, isTestVersion bool) (*scProcessorProxy, error) {
	scProcessorProxy := &scProcessorProxy{
		args: scrCommon.ArgsNewSmartContractProcessor{
			VmContainer:         args.VmContainer,
			ArgsParser:          args.ArgsParser,
			Hasher:              args.Hasher,
			Marshalizer:         args.Marshalizer,
			AccountsDB:          args.AccountsDB,
			BlockChainHook:      args.BlockChainHook,
			BuiltInFunctions:    args.BuiltInFunctions,
			PubkeyConv:          args.PubkeyConv,
			ShardCoordinator:    args.ShardCoordinator,
			ScrForwarder:        args.ScrForwarder,
			TxFeeHandler:        args.TxFeeHandler,
			EconomicsFee:        args.EconomicsFee,
			TxTypeHandler:       args.TxTypeHandler,
			GasHandler:          args.GasHandler,
			GasSchedule:         args.GasSchedule,
			TxLogsProcessor:     args.TxLogsProcessor,
			BadTxForwarder:      args.BadTxForwarder,
			EnableEpochsHandler: args.EnableEpochsHandler,
			EnableEpochs:        args.EnableEpochs,
			VMOutputCacher:      args.VMOutputCacher,
			ArwenChangeLocker:   args.ArwenChangeLocker,
			IsGenesisProcessing: args.IsGenesisProcessing,
		},
		isTestVersion: isTestVersion,
	}

	scProcessorProxy.processorsCache = make(map[configuredProcessor]process.SmartContractProcessorFacade)
	if isTestVersion {
		scProcessorProxy.testProcessorsCache = make(map[configuredProcessor]scrCommon.TestSmartContractProcessor)
	}

	var err error
	err = scProcessorProxy.createProcessorV1()
	if err != nil {
		return nil, err
	}

	err = scProcessorProxy.createProcessorV2()
	if err != nil {
		return nil, err
	}

	epochNotifier.RegisterNotifyHandler(scProcessorProxy)

	return scProcessorProxy, nil
}

func (proxy *scProcessorProxy) createProcessorV1() error {
	processor, err := smartContract.NewSmartContractProcessor(proxy.args)
	proxy.processorsCache[procV1] = processor
	if proxy.isTestVersion {
		proxy.testProcessorsCache[procV1] = smartContract.NewTestScProcessor(processor)
	}
	return err
}

func (proxy *scProcessorProxy) createProcessorV2() error {
	processor, err := processorV2.NewSmartContractProcessorV2(proxy.args)
	proxy.processorsCache[procV2] = processor
	if proxy.isTestVersion {
		proxy.testProcessorsCache[procV2] = processorV2.NewTestScProcessor(processor)
	}
	return err
}

func (proxy *scProcessorProxy) setActiveProcessorV1() {
	proxy.setActiveProcessor(procV1)
}

func (proxy *scProcessorProxy) setActiveProcessorV2() {
	proxy.setActiveProcessor(procV2)
}

func (proxy *scProcessorProxy) setActiveProcessor(version configuredProcessor) {
	log.Info("processorProxy", "configured", version)
	proxy.configuredProcessor = version
	proxy.processor = proxy.processorsCache[version]
	if proxy.isTestVersion {
		proxy.testScProcessor = proxy.testProcessorsCache[version]
	}
}

func (proxy *scProcessorProxy) getProcessor() process.SmartContractProcessorFacade {
	proxy.mutRc.Lock()
	defer proxy.mutRc.Unlock()
	return proxy.processor
}

// ExecuteSmartContractTransaction delegates to selected processor
func (proxy *scProcessorProxy) ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

// ExecuteBuiltInFunction delegates to selected processor
func (proxy *scProcessorProxy) ExecuteBuiltInFunction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

// DeploySmartContract delegates to selected processor
func (proxy *scProcessorProxy) DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().DeploySmartContract(tx, acntSrc)
}

// ProcessIfError delegates to selected processor
func (proxy *scProcessorProxy) ProcessIfError(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
	return proxy.getProcessor().ProcessIfError(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked)
}

// IsPayable delegates to selected processor
func (proxy *scProcessorProxy) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	return proxy.getProcessor().IsPayable(sndAddress, recvAddress)
}

// ProcessSmartContractResult delegates to selected processor
func (proxy *scProcessorProxy) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ProcessSmartContractResult(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (proxy *scProcessorProxy) IsInterfaceNil() bool {
	return proxy == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (proxy *scProcessorProxy) EpochConfirmed(_ uint32, _ uint64) {
	proxy.mutRc.Lock()
	defer proxy.mutRc.Unlock()

	if proxy.args.EnableEpochsHandler.IsSCProcessorV2FlagEnabled() {
		proxy.setActiveProcessorV2()
		return
	}

	proxy.setActiveProcessorV1()
}

// GetCompositeTestError delegates to the selected testScProcessor
func (proxy *scProcessorProxy) GetCompositeTestError() error {
	return proxy.testScProcessor.GetCompositeTestError()
}

// GetGasRemaining delegates to the selected testScProcessor
func (proxy *scProcessorProxy) GetGasRemaining() uint64 {
	return proxy.testScProcessor.GetGasRemaining()
}

// GetAllSCRs delegates to the selected testScProcessor
func (proxy *scProcessorProxy) GetAllSCRs() []data.TransactionHandler {
	return proxy.testScProcessor.GetAllSCRs()
}

// CleanGasRefunded delegates to the selected testScProcessor
func (proxy *scProcessorProxy) CleanGasRefunded() {
	proxy.testScProcessor.CleanGasRefunded()
}
