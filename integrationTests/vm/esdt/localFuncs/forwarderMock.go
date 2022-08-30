package localFuncs

import (
	"math/big"

	"github.com/ElrondNetwork/arwen-wasm-vm/v1_5/arwen/elrondapi"
	mock "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/mock/context"
	test "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/testcommon"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

// MultiTransferViaAsyncMock is an exposed mock contract method
func MultiTransferViaAsyncMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("multi_transfer_via_async", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		testConfig := config.(*test.TestConfig)

		scAddress := host.Runtime().GetContextAddress()

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()
		destAddress := args[0]
		numOfTransfers := (len(args) - 1) / 3

		callData := txDataBuilder.NewBuilder()
		callData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
		callData.Bytes(destAddress)
		callData.Int(numOfTransfers) // no of triplets
		for a := 1; a < len(args); a++ {
			callData.Bytes(args[a])
		}
		callData.Str("accept_multi_funds_echo")

		value := big.NewInt(testConfig.TransferFromParentToChild).Bytes()
		err := arwenvm.RegisterAsyncCallForMockContract(host, config, scAddress, value, callData)
		if err != nil {
			host.Runtime().SignalUserError(err.Error())
			return instance
		}

		return instance

	})
}

// SyncMultiTransferMock is an exposed mock contract method
func SyncMultiTransferMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("forward_sync_accept_funds_multi_transfer", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		scAddress := host.Runtime().GetContextAddress()

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()
		destAddress := args[0]
		numOfTransfers := (len(args) - 1) / 3

		newArgs := make([][]byte, len(args)+1)
		newArgs[0] = destAddress
		newArgs[1] = big.NewInt(int64(numOfTransfers)).Bytes()
		for i := 1; i < len(args); i++ {
			newArgs[1+i] = args[i]
		}
		newArgs = append(newArgs, []byte("accept_funds_echo"))

		elrondapi.ExecuteOnDestContextWithTypedArgs(
			host,
			1_000_000,
			big.NewInt(0),
			[]byte(core.BuiltInFunctionMultiESDTNFTTransfer),
			scAddress,
			newArgs)

		return instance
	})
}

// MultiTransferExecuteMock is an exposed mock contract method
func MultiTransferExecuteMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("forward_transf_exec_accept_funds_multi_transfer", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		// destAddress + ESDT transfer tripplets (TokenIdentifier + Nonce + Amount)
		args := host.Runtime().Arguments()
		destAddress := args[0]
		numOfTransfers := (len(args) - 1) / 3

		transfers := make([]*vmcommon.ESDTTransfer, numOfTransfers)
		for i := 0; i < numOfTransfers; i++ {
			tokenStartIndex := 1 + i*parsers.ArgsPerTransfer
			transfer := &vmcommon.ESDTTransfer{
				ESDTTokenName:  args[tokenStartIndex],
				ESDTTokenNonce: big.NewInt(0).SetBytes(args[tokenStartIndex+1]).Uint64(),
				ESDTValue:      big.NewInt(0).SetBytes(args[tokenStartIndex+2]),
				ESDTTokenType:  uint32(core.Fungible),
			}
			if transfer.ESDTTokenNonce > 0 {
				transfer.ESDTTokenType = uint32(core.NonFungible)
			}
			transfers[i] = transfer
		}

		elrondapi.TransferESDTNFTExecuteWithTypedArgs(
			host,
			destAddress,
			transfers,
			1_000_000,
			[]byte("accept_multi_funds_echo"),
			[][]byte{})

		return instance
	})
}

// EmptyCallbackMock is an exposed mock contract method
func EmptyCallbackMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("callBack", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// AcceptMultiFundsEchoMock is an exposed mock contract method
func AcceptMultiFundsEchoMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("accept_multi_funds_echo", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// AcceptFundsEchoMock is an exposed mock contract method
func AcceptFundsEchoMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("accept_funds_echo", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		return instance
	})
}

// DoAsyncCallMock is an exposed mock contract method
func DoAsyncCallMock(instanceMock *mock.InstanceMock, config interface{}) {
	instanceMock.AddMockMethod("doAsyncCall", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)

		args := host.Runtime().Arguments()
		destAddress := args[0]
		egldValue := args[1]
		function := string(args[2])

		callData := txDataBuilder.NewBuilder()
		callData.Func(function)
		for a := 2; a < len(args); a++ {
			callData.Bytes(args[a])
		}

		err := arwenvm.RegisterAsyncCallForMockContract(host, config, destAddress, egldValue, callData)
		if err != nil {
			host.Runtime().SignalUserError(err.Error())
			return instance
		}

		return instance
	})
}
