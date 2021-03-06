package wal

import (
	"encoding/gob"
	"io"
	"math"
)

const (
	CURRENT_VERSION = 1
)

type globalState struct {
}

type state struct {
	// per log file state
	Version                   byte
	RequestsSinceLastBookmark int
	RequestsSinceLastIndex    uint32
	FileOffset                int64 // the file offset at which this bookmark was created
	Index                     *index
	TotalNumberOfRequests     int
	LargestRequestNumber      uint32
	ShardLastSequenceNumber   map[uint32]uint64
	ServerLastRequestNumber   map[uint32]uint32
}

func newState() *state {
	return &state{
		Version: CURRENT_VERSION,
		Index: &index{
			Entries: make([]*indexEntry, 0),
		},
		LargestRequestNumber:    0,
		ShardLastSequenceNumber: make(map[uint32]uint64),
		ServerLastRequestNumber: make(map[uint32]uint32),
	}
}

func (self *state) recover(replay *replayRequest) {
	if self.LargestRequestNumber < replay.requestNumber {
		self.LargestRequestNumber = replay.requestNumber
	}

	lastSequenceNumber := self.ShardLastSequenceNumber[replay.shardId]

	for _, p := range replay.request.Series.Points {
		if seq := p.GetSequenceNumber(); seq > lastSequenceNumber {
			lastSequenceNumber = seq
		}
	}

	self.ShardLastSequenceNumber[replay.shardId] = lastSequenceNumber
}

func (self *state) setFileOffset(offset int64) {
	self.FileOffset = offset
}

func (self *state) getNextRequestNumber() uint32 {
	self.LargestRequestNumber++
	return self.LargestRequestNumber
}

func (self *state) continueFromState(state *state) {
	self.LargestRequestNumber = state.LargestRequestNumber
	self.ShardLastSequenceNumber = state.ShardLastSequenceNumber
	self.ServerLastRequestNumber = state.ServerLastRequestNumber
}

func (self *state) getCurrentSequenceNumber(shardId uint32) uint64 {
	return self.ShardLastSequenceNumber[shardId]
}

func (self *state) setCurrentSequenceNumber(shardId uint32, sequenceNumber uint64) {
	self.ShardLastSequenceNumber[shardId] = sequenceNumber
}

func (self *state) commitRequestNumber(serverId, requestNumber uint32) {
	self.ServerLastRequestNumber[serverId] = requestNumber
}

func (self *state) LowestCommitedRequestNumber() uint32 {
	requestNumber := uint32(math.MaxUint32)
	for _, number := range self.ServerLastRequestNumber {
		if number < requestNumber {
			requestNumber = number
		}
	}
	return requestNumber
}

func (self *state) write(w io.Writer) error {
	return gob.NewEncoder(w).Encode(self)
}

func (self *state) read(r io.Reader) error {
	return gob.NewDecoder(r).Decode(self)
}
