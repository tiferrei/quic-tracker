package masterthesis

import (
	"bytes"
	"encoding/binary"
	"io"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

type Frame interface {
	FrameType() FrameType
	writeTo(buffer *bytes.Buffer)
}
func NewFrame(buffer *bytes.Reader, conn *Connection) Frame {
	typeByte, err := buffer.ReadByte()
	if err == io.EOF {
		return nil
	} else if err != nil {
		panic(err)
	}
	buffer.UnreadByte()
	frameType := FrameType(typeByte)
	switch {
	case frameType == PaddingFrameType:
		return Frame(NewPaddingFrame(buffer))
	case frameType == ResetStreamType:
		return Frame(NewResetStream(buffer))
	case frameType == ConnectionCloseType:
		return Frame(NewConnectionCloseFrame(buffer))
	case frameType == ApplicationCloseType:
		return Frame(NewApplicationCloseFrame(buffer))
	case frameType == MaxDataType:
		return Frame(NewMaxDataFrame(buffer))
	case frameType == MaxStreamDataType:
		return Frame(NewMaxStreamDataFrame(buffer))
	case frameType == MaxStreamIdType:
		return Frame(NewMaxStreamIdFrame(buffer))
	case frameType == PingType:
		return Frame(NewPingFrame(buffer))
	case frameType == BlockedType:
		return Frame(NewBlockedFrame(buffer))
	case frameType == StreamBlockedType:
		return Frame(NewStreamBlockedFrame(buffer))
	case frameType == StreamIdBlockedType:
		return Frame(NewStreamIdNeededFrame(buffer))
	case frameType == NewConnectionIdType:
		return Frame(NewNewConnectionIdFrame(buffer))
	case frameType == StopSendingType:
		return Frame(NewStopSendingFrame(buffer))
	case frameType == AckType:
		return Frame(ReadAckFrame(buffer))
	case (frameType & StreamType) == StreamType:
		return Frame(ReadStreamFrame(buffer, conn))
	default:
		spew.Dump(buffer)
		panic(fmt.Sprintf("Unknown frame type %d", typeByte))
	}
}
type FrameType uint8

const PaddingFrameType FrameType = 0x00
const ResetStreamType FrameType = 0x01
const ConnectionCloseType FrameType = 0x02
const ApplicationCloseType FrameType = 0x03
const MaxDataType FrameType = 0x04
const MaxStreamDataType FrameType = 0x05
const MaxStreamIdType FrameType = 0x06
const PingType FrameType = 0x07
const BlockedType FrameType = 0x08
const StreamBlockedType FrameType = 0x09
const StreamIdBlockedType FrameType = 0x0a
const NewConnectionIdType FrameType = 0x0b
const StopSendingType FrameType = 0x0c
const PongType FrameType = 0x0d
const AckType FrameType = 0x0e
const StreamType FrameType = 0x10

type PaddingFrame byte

func (frame PaddingFrame) FrameType() FrameType { return PaddingFrameType }
func (frame PaddingFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
}
func NewPaddingFrame(buffer *bytes.Reader) *PaddingFrame {
	buffer.ReadByte()  // Discard frame payload
	return new(PaddingFrame)
}

type ResetStream struct {
	streamId    uint64
	errorCode   uint16
	finalOffset uint64
}
func (frame ResetStream) FrameType() FrameType { return ResetStreamType }
func (frame ResetStream) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.streamId)
	binary.Write(buffer, binary.BigEndian, frame.errorCode)
	WriteVarInt(buffer, frame.finalOffset)
}
func NewResetStream(buffer *bytes.Reader) *ResetStream {
	frame := new(ResetStream)
	buffer.ReadByte()  // Discard frame type
	frame.streamId, _ = ReadVarInt(buffer)
	binary.Read(buffer, binary.BigEndian, &frame.errorCode)
	frame.finalOffset, _ = ReadVarInt(buffer)
	return frame
}

type ConnectionCloseFrame struct {
	errorCode          uint16
	reasonPhraseLength uint64
	reasonPhrase       string
}
func (frame ConnectionCloseFrame) FrameType() FrameType { return ConnectionCloseType }
func (frame ConnectionCloseFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	binary.Write(buffer, binary.BigEndian, frame.errorCode)
	WriteVarInt(buffer, frame.reasonPhraseLength)
	if frame.reasonPhraseLength > 0 {
		buffer.Write([]byte(frame.reasonPhrase))
	}
}
func NewConnectionCloseFrame(buffer *bytes.Reader) *ConnectionCloseFrame {
	frame := new(ConnectionCloseFrame)
	buffer.ReadByte()  // Discard frame type
	binary.Read(buffer, binary.BigEndian, &frame.errorCode)
	frame.reasonPhraseLength, _ = ReadVarInt(buffer)
	if frame.reasonPhraseLength > 0 {
		reasonBytes := make([]byte, frame.reasonPhraseLength, frame.reasonPhraseLength)
		binary.Read(buffer, binary.BigEndian, &reasonBytes)
		frame.reasonPhrase = string(reasonBytes)
	}
	return frame
}

type ApplicationCloseFrame struct {
	errorCode          uint16
	reasonPhraseLength uint64
	reasonPhrase       string
}
func (frame ApplicationCloseFrame) FrameType() FrameType { return ApplicationCloseType }
func (frame ApplicationCloseFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	binary.Write(buffer, binary.BigEndian, frame.errorCode)
	binary.Write(buffer, binary.BigEndian, frame.reasonPhraseLength)
	if frame.reasonPhraseLength > 0 {
		buffer.Write([]byte(frame.reasonPhrase))
	}
}
func NewApplicationCloseFrame(buffer *bytes.Reader) *ApplicationCloseFrame {
	frame := new(ApplicationCloseFrame)
	buffer.ReadByte()  // Discard frame type
	binary.Read(buffer, binary.BigEndian, &frame.errorCode)
	frame.reasonPhraseLength, _ = ReadVarInt(buffer)
	if frame.reasonPhraseLength > 0 {
		reasonBytes := make([]byte, frame.reasonPhraseLength, frame.reasonPhraseLength)
		binary.Read(buffer, binary.BigEndian, &reasonBytes)
		frame.reasonPhrase = string(reasonBytes)
	}
	return frame
}


type MaxDataFrame struct {
	maximumData uint64
}
func (frame MaxDataFrame) FrameType() FrameType { return MaxDataType }
func (frame MaxDataFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.maximumData)
}
func NewMaxDataFrame(buffer *bytes.Reader) *MaxDataFrame {
	frame := new(MaxDataFrame)
	buffer.ReadByte()  // Discard frame type
	frame.maximumData, _ = ReadVarInt(buffer)
	binary.Read(buffer, binary.BigEndian, &frame.maximumData)
	return frame
}

type MaxStreamDataFrame struct {
	streamId uint64
	maximumStreamData uint64
}
func (frame MaxStreamDataFrame) FrameType() FrameType { return MaxStreamDataType }
func (frame MaxStreamDataFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.streamId)
	WriteVarInt(buffer, frame.maximumStreamData)
}
func NewMaxStreamDataFrame(buffer *bytes.Reader) *MaxStreamDataFrame {
	frame := new(MaxStreamDataFrame)
	buffer.ReadByte()  // Discard frame type
	frame.streamId, _ = ReadVarInt(buffer)
	frame.maximumStreamData, _ = ReadVarInt(buffer)
	return frame
}

type MaxStreamIdFrame struct {
	maximumStreamId uint64
}
func (frame MaxStreamIdFrame) FrameType() FrameType { return MaxStreamIdType }
func (frame MaxStreamIdFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.maximumStreamId)
}
func NewMaxStreamIdFrame(buffer *bytes.Reader) *MaxStreamIdFrame {
	frame := new(MaxStreamIdFrame)
	buffer.ReadByte()  // Discard frame type
	frame.maximumStreamId, _ = ReadVarInt(buffer)
	return frame
}


type PingFrame struct {
	length uint8
	data []byte
}
func (frame PingFrame) FrameType() FrameType { return PingType }
func (frame PingFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	buffer.WriteByte(frame.length)
	if frame.length > 0 {
		buffer.Write(frame.data)
	}
}
func NewPingFrame(buffer *bytes.Reader) *PingFrame {
	frame := new(PingFrame)
	buffer.ReadByte()  // Discard frame type
	frame.length, _ = buffer.ReadByte()
	if frame.length > 0 {
		frame.data = make([]byte, frame.length, frame.length)
		buffer.Read(frame.data)
	}
	return frame
}

type BlockedFrame struct {
	offset uint64
}
func (frame BlockedFrame) FrameType() FrameType { return BlockedType }
func (frame BlockedFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.offset)
}
func NewBlockedFrame(buffer *bytes.Reader) *BlockedFrame {
	frame := new(BlockedFrame)
	buffer.ReadByte()  // Discard frame type
	frame.offset, _ = ReadVarInt(buffer)
	return frame
}

type StreamBlockedFrame struct {
	streamId uint64
	offset   uint64
}
func (frame StreamBlockedFrame) FrameType() FrameType { return StreamBlockedType }
func (frame StreamBlockedFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.streamId)
	WriteVarInt(buffer, frame.offset)
}
func NewStreamBlockedFrame(buffer *bytes.Reader) *StreamBlockedFrame {
	frame := new(StreamBlockedFrame)
	buffer.ReadByte()  // Discard frame type
	frame.streamId, _ = ReadVarInt(buffer)
	frame.offset, _ = ReadVarInt(buffer)
	return frame
}

type StreamIdBlockedFrame struct {
	streamId uint64
}
func (frame StreamIdBlockedFrame) FrameType() FrameType { return StreamIdBlockedType }
func (frame StreamIdBlockedFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.streamId)
}
func NewStreamIdNeededFrame(buffer *bytes.Reader) *StreamIdBlockedFrame {
	frame := new(StreamIdBlockedFrame)
	buffer.ReadByte()  // Discard frame type
	frame.streamId, _ = ReadVarInt(buffer)
	return frame
}

type NewConnectionIdFrame struct {
	sequence            uint64
	connectionId        uint64
	statelessResetToken [16]byte
}
func (frame NewConnectionIdFrame) FrameType() FrameType { return NewConnectionIdType }
func (frame NewConnectionIdFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.sequence)
	WriteVarInt(buffer, frame.connectionId)
	binary.Write(buffer, binary.BigEndian, frame.statelessResetToken)
}
func NewNewConnectionIdFrame(buffer *bytes.Reader) *NewConnectionIdFrame {
	frame := new(NewConnectionIdFrame)
	buffer.ReadByte()  // Discard frame type
	frame.sequence, _ = ReadVarInt(buffer)
	frame.connectionId, _ = ReadVarInt(buffer)
	binary.Read(buffer, binary.BigEndian, &frame.statelessResetToken)
	return frame
}

type StopSendingFrame struct {
	streamId  uint64
	errorCode uint16
}
func (frame StopSendingFrame) FrameType() FrameType { return StopSendingType }
func (frame StopSendingFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.streamId)
	binary.Write(buffer, binary.BigEndian, frame.errorCode)
}
func NewStopSendingFrame(buffer *bytes.Reader) *StopSendingFrame {
	frame := new(StopSendingFrame)
	buffer.ReadByte()  // Discard frame type
	frame.streamId, _ = ReadVarInt(buffer)
	binary.Read(buffer, binary.BigEndian, &frame.errorCode)
	return frame
}

type PongFrame struct {
	PingFrame
}

func (frame PongFrame) FrameType() FrameType { return PongType }

func NewPongFrame(buffer *bytes.Reader) *PongFrame {
	frame := new(PongFrame)
	buffer.ReadByte()  // Discard frame type
	frame.length, _ = buffer.ReadByte()
	if frame.length > 0 {
		frame.data = make([]byte, frame.length, frame.length)
		buffer.Read(frame.data)
	}
	return frame
}

type AckFrame struct {
	largestAcknowledged       uint64
	ackDelay                  uint64
	ackBlockCount              uint64
	ackBlocks                 []AckBlock
}
type AckBlock struct {
	gap uint64
	block uint64
}
func (frame AckFrame) FrameType() FrameType { return AckType }
func (frame AckFrame) writeTo(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, frame.FrameType())
	WriteVarInt(buffer, frame.largestAcknowledged)
	WriteVarInt(buffer, frame.ackDelay)
	WriteVarInt(buffer, frame.ackBlockCount)
	for i, ack := range frame.ackBlocks {
		if i > 0 {
			WriteVarInt(buffer, ack.gap)
		}
		WriteVarInt(buffer, ack.block)
	}
}
func ReadAckFrame(buffer *bytes.Reader) *AckFrame {
	frame := new(AckFrame)
	buffer.ReadByte()  // Discard frame byte

	frame.largestAcknowledged, _ = ReadVarInt(buffer)
	frame.ackDelay, _ = ReadVarInt(buffer)
	frame.ackBlockCount, _ = ReadVarInt(buffer)

	firstBlock := AckBlock{}
	firstBlock.block, _ = ReadVarInt(buffer)

	var i uint64
	for i = 0; i < frame.ackBlockCount; i++ {
		ack := AckBlock{}
		ack.gap, _ = ReadVarInt(buffer)
		ack.block, _ = ReadVarInt(buffer)
		frame.ackBlocks = append(frame.ackBlocks, ack)
	}
	return frame
}
func NewAckFrame(largestAcknowledged uint64, ackBlockCount uint64) *AckFrame {
	frame := new(AckFrame)
	frame.largestAcknowledged = largestAcknowledged
	frame.ackBlockCount = 0
	frame.ackDelay = 0
	frame.ackBlocks = append(frame.ackBlocks, AckBlock{0, ackBlockCount})
	return frame
}

type StreamFrame struct {
	finBit bool
	lenBit bool
	offBit bool

	streamId uint64
	offset   uint64
	length   uint64
	streamData []byte
}
func (frame StreamFrame) FrameType() FrameType { return StreamType }
func (frame StreamFrame) writeTo(buffer *bytes.Buffer) {
	typeByte := byte(frame.FrameType())
	if frame.finBit {
		typeByte |= 0x01
	}
	if frame.lenBit {
		typeByte |= 0x02
	}
	if frame.offBit {
		typeByte |= 0x04
	}
	binary.Write(buffer, binary.BigEndian, typeByte)
	WriteVarInt(buffer, frame.streamId)
	if frame.offBit {
		WriteVarInt(buffer, frame.offset)
	}
	if frame.lenBit {
		WriteVarInt(buffer, frame.length)
	}
	buffer.Write(frame.streamData)
}
func ReadStreamFrame(buffer *bytes.Reader, conn *Connection) *StreamFrame {
	frame := new(StreamFrame)
	typeByte, _ := buffer.ReadByte()
	frame.finBit = (typeByte & 0x01) == 0x01
	frame.lenBit = (typeByte & 0x02) == 0x02
	frame.offBit = (typeByte & 0x04) == 0x04

	frame.streamId, _ = ReadVarInt(buffer)
	if frame.offBit {
		frame.offset, _ = ReadVarInt(buffer)
	}
	if frame.lenBit {
		frame.length, _ = ReadVarInt(buffer)
	}
	return frame
}
func NewStreamFrame(streamId uint32, stream *Stream, data []byte, finBit bool) *StreamFrame {
	frame := new(StreamFrame)
	frame.finBit = finBit
	frame.lenBit = true
	frame.offset = stream.writeOffset
	frame.offBit = frame.offset > 0
	frame.length = uint64(len(data))
	frame.streamData = data
	stream.writeOffset += uint64(frame.length)
	return frame
}
