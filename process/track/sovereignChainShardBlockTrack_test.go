package track_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignChainShardBlockTrack_ShouldErrNilBlockTracker(t *testing.T) {
	t.Parallel()

	scsbt, err := track.NewSovereignChainShardBlockTrack(nil)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrNilBlockTracker, err)
}

func TestNewSovereignChainShardBlockTrack_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	sbt.SetBlockProcessor(nil)

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.Nil(t, scsbt)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestNewSovereignChainShardBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()

	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

	scsbt, err := track.NewSovereignChainShardBlockTrack(sbt)
	assert.NotNil(t, scsbt)
	assert.Nil(t, err)
}

func TestSovereignChainShardBlockTrack_ComputeLongestSelfChainShouldWork(t *testing.T) {
	t.Parallel()

	lastNotarizedHeader := &block.Header{Nonce: 1}
	lastNotarizedHash := []byte("hash")
	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = 1
	shardBlockTrackArguments.ShardCoordinator = shardCoordinatorMock
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)
	selfNotarizer := &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID != shardCoordinatorMock.CurrentShard {
				return nil, nil, errors.New("wrong shard ID")
			}
			return lastNotarizedHeader, lastNotarizedHash, nil
		},
	}
	sbt.SetSelfNotarizer(selfNotarizer)
	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header, hash, headers, hashes := scsbt.ComputeLongestSelfChain()

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 0, len(hashes))
}

func TestSovereignChainShardBlockTrack_GetSelfNotarizedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	lastNotarizedHeader := &block.Header{Nonce: 1}
	lastNotarizedHash := []byte("hash")
	shardBlockTrackArguments := CreateShardTrackerMockArguments()
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = 1
	shardBlockTrackArguments.ShardCoordinator = shardCoordinatorMock
	sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)
	selfNotarizer := &mock.BlockNotarizerHandlerMock{
		GetNotarizedHeaderCalled: func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
			if shardID != shardCoordinatorMock.CurrentShard {
				return nil, nil, errors.New("wrong shard ID")
			}
			return lastNotarizedHeader, lastNotarizedHash, nil
		},
	}
	sbt.SetSelfNotarizer(selfNotarizer)
	scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

	header, hash, err := scsbt.GetSelfNotarizedHeader(core.MetachainShardId, 0)

	assert.Equal(t, lastNotarizedHeader, header)
	assert.NotNil(t, lastNotarizedHash, hash)
	assert.Nil(t, err)
}

func TestSovereignChainShardBlockTrack_ReceivedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should add extended shard header to sovereign chain tracked headers", func(t *testing.T) {
		t.Parallel()

		shardBlockTrackArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		header := &block.Header{Nonce: 1}
		headerV2 := &block.HeaderV2{Header: header}
		extendedShardHeader := &block.ShardHeaderExtended{
			Header: headerV2,
		}
		extendedShardHeaderHash := []byte("hash")
		scsbt.ReceivedHeader(extendedShardHeader, extendedShardHeaderHash)
		headers, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)

		require.Equal(t, 1, len(headers))
		assert.Equal(t, extendedShardHeader, headers[0])
	})

	t.Run("should add shard header to sovereign chain tracked headers", func(t *testing.T) {
		t.Parallel()

		shardBlockTrackArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardBlockTrackArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		header := &block.Header{Nonce: 1}
		headerHash := []byte("hash")
		scsbt.ReceivedHeader(header, headerHash)
		headers, _ := scsbt.GetTrackedHeaders(header.GetShardID())

		require.Equal(t, 1, len(headers))
		assert.Equal(t, header, headers[0])
	})
}

func TestSovereignChainShardBlockTrack_ReceivedExtendedShardHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should not add to tracked headers when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: 1001,
				},
			},
		}
		shardHeaderExtendedHash := []byte("hash")

		scsbt.ReceivedExtendedShardHeader(shardHeaderExtended, shardHeaderExtendedHash)
		headers, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)
		assert.Zero(t, len(headers))
	})

	t.Run("should add to tracked headers when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: 1000,
				},
			},
		}
		shardHeaderExtendedHash := []byte("hash")

		scsbt.ReceivedExtendedShardHeader(shardHeaderExtended, shardHeaderExtendedHash)
		headers, _ := scsbt.GetTrackedHeaders(core.SovereignChainShardId)

		require.Equal(t, 1, len(headers))
		assert.Equal(t, shardHeaderExtended, headers[0])
	})
}

func TestSovereignChainShardBlockTrack_ShouldAddExtendedShardHeaderShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should return true when first extended shard header is added", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		maxNumHeadersToKeepPerShard := uint64(scsbt.GetMaxNumHeadersToKeepPerShard())

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: maxNumHeadersToKeepPerShard + 1,
				},
			},
		}

		result := scsbt.ShouldAddExtendedShardHeader(shardHeaderExtended)
		assert.True(t, result)
	})

	t.Run("should return false when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		maxNumHeadersToKeepPerShard := uint64(scsbt.GetMaxNumHeadersToKeepPerShard())

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: maxNumHeadersToKeepPerShard + 1,
				},
			},
		}

		result := scsbt.ShouldAddExtendedShardHeader(shardHeaderExtended)
		assert.False(t, result)
	})

	t.Run("should return true when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		maxNumHeadersToKeepPerShard := uint64(scsbt.GetMaxNumHeadersToKeepPerShard())

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: maxNumHeadersToKeepPerShard,
				},
			},
		}

		result := scsbt.ShouldAddExtendedShardHeader(shardHeaderExtended)
		assert.True(t, result)
	})
}

func TestSovereignChainShardBlockTrack_DoWhitelistWithExtendedShardHeaderIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should not whitelist when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		cache := make(map[string]struct{})
		mutCache := sync.Mutex{}
		shardArguments := CreateShardTrackerMockArguments()
		shardArguments.WhitelistHandler = &testscommon.WhiteListHandlerStub{
			AddCalled: func(keys [][]byte) {
				mutCache.Lock()
				for _, key := range keys {
					cache[string(key)] = struct{}{}
				}
				mutCache.Unlock()
			},
		}
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		txHash := []byte("hash")
		incomingMiniBlocks := []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					txHash,
				},
			},
		}
		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: process.MaxHeadersToWhitelistInAdvance + 1,
				},
			},
			IncomingMiniBlocks: incomingMiniBlocks,
		}

		scsbt.DoWhitelistWithExtendedShardHeaderIfNeeded(shardHeaderExtended)
		_, ok := cache[string(txHash)]

		assert.False(t, ok)
	})

	t.Run("should whitelist when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		cache := make(map[string]struct{})
		mutCache := sync.Mutex{}
		shardArguments := CreateShardTrackerMockArguments()
		shardArguments.WhitelistHandler = &testscommon.WhiteListHandlerStub{
			AddCalled: func(keys [][]byte) {
				mutCache.Lock()
				for _, key := range keys {
					cache[string(key)] = struct{}{}
				}
				mutCache.Unlock()
			},
		}
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		txHash := []byte("hash")
		incomingMiniBlocks := []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					txHash,
				},
			},
		}
		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: process.MaxHeadersToWhitelistInAdvance,
				},
			},
			IncomingMiniBlocks: incomingMiniBlocks,
		}

		scsbt.DoWhitelistWithExtendedShardHeaderIfNeeded(shardHeaderExtended)
		_, ok := cache[string(txHash)]

		assert.True(t, ok)
	})
}

func TestSovereignChainShardBlockTrack_IsExtendedShardHeaderOutOfRangeShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should return true when extended shard header is out of range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		nonce := uint64(8)
		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce,
				},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce + process.MaxHeadersToWhitelistInAdvance + 1,
				},
			},
		}

		assert.True(t, scsbt.IsExtendedShardHeaderOutOfRange(shardHeaderExtended))
	})

	t.Run("should return false when extended shard header is in range", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		nonce := uint64(8)
		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce,
				},
			},
		}
		shardHeaderExtendedInitHash := []byte("init_hash")

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Nonce: nonce + process.MaxHeadersToWhitelistInAdvance,
				},
			},
		}

		assert.False(t, scsbt.IsExtendedShardHeaderOutOfRange(shardHeaderExtended))
	})
}

func TestSovereignChainShardBlockTrack_ComputeLongestExtendedShardChainFromLastNotarizedShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should return error when notarized header slice for shard is nil", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		_, _, err := scsbt.ComputeLongestExtendedShardChainFromLastNotarized()

		assert.Equal(t, err, process.ErrNotarizedHeadersSliceForShardIsNil)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		scsbt, _ := track.NewSovereignChainShardBlockTrack(sbt)

		shardHeaderExtendedInit := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					RandSeed: []byte("rand seed init"),
				},
			},
		}

		shardHeaderExtendedInitHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtendedInit)

		scsbt.AddCrossNotarizedHeader(core.SovereignChainShardId, shardHeaderExtendedInit, shardHeaderExtendedInitHash)

		shardHeaderExtended1 := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:        1,
					Nonce:        1,
					PrevHash:     shardHeaderExtendedInitHash,
					PrevRandSeed: shardHeaderExtendedInit.GetRandSeed(),
					RandSeed:     []byte("rand seed 1"),
				},
			},
		}

		shardHeaderExtendedHash1, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended1)

		shardHeaderExtended2 := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:        2,
					Nonce:        2,
					PrevHash:     shardHeaderExtendedHash1,
					PrevRandSeed: shardHeaderExtended1.GetRandSeed(),
					RandSeed:     []byte("rand seed 2"),
				},
			},
		}

		shardHeaderExtendedHash2, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended2)

		shardHeaderExtended3 := &block.ShardHeaderExtended{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round:        3,
					Nonce:        3,
					PrevHash:     shardHeaderExtendedHash2,
					PrevRandSeed: shardHeaderExtended2.GetRandSeed(),
					RandSeed:     []byte("rand seed 3"),
				},
			},
		}

		shardHeaderExtendedHash3, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, shardHeaderExtended3)

		scsbt.AddTrackedHeader(shardHeaderExtended1, shardHeaderExtendedHash1)
		scsbt.AddTrackedHeader(shardHeaderExtended2, shardHeaderExtendedHash2)
		scsbt.AddTrackedHeader(shardHeaderExtended3, shardHeaderExtendedHash3)

		headers, _, _ := scsbt.ComputeLongestExtendedShardChainFromLastNotarized()

		require.Equal(t, 2, len(headers))
		assert.Equal(t, shardHeaderExtended1, headers[0])
		assert.Equal(t, shardHeaderExtended2, headers[1])
	})
}
