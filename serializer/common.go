package serializer

import (
	"errors"
	"sync"
)

const (
	//byte
	SinglePkgMaxSize = 20
	MaxSize          = 100
	HeaderSize       = 1
	TypeSize         = 1
	BtnNumSize       = 1
	ValueSize        = 15
	ChecksumSize     = 1
	TailSize         = 1
)

const (
	//0xAA
	HeaderSvrToClientBin = 0b10101010
	//0xCC
	HeaderClientToSvrBin = 0b11001100
	//0xBB
	TailSvrToClientBin = 0b10111011
	//0xDD
	TailClientToSvrBin = 0b11011101
	//0xA
	ActionTypeAllOpenBin = 0b1010
	//0xB
	ActionTypeAllCloseBin = 0b1011
	//0xC
	ActionTypeQueryBin = 0b1100
	//0xD
	ActionTypeSingleOpenBin = 0b1101
	//0xE
	ActionTypeSingleCloseBin = 0b1110
	//0x1E 向服务器注册端点，是的服务器每7秒向端点报告一次继电器状态
	ActionTypeRegisteStatusBin = 0b11110
	//0x1A 年 年 月 日 时 分 秒 （年由两部分组成，例如2017：14 11）
	ActionTypeSetDateTimeBin = 0b11010
	//0x1B
	ActionTypeGetDateTimeBin = 0b11011
	//0x2b value 3-9分别代表DIO_20到DIO_28这7个IO口的状态，01高电平，02低电平
	ActionTypeGetIOBin = 0b101011
	//0x2d 批量设置继电器状态 3-7 五路继电器，01开，02关
	ActionTypeBatchSetBin = 0b101101
	//0x3d
	ActionTypeDelayOpenBin = 0b111101
	//0x3e
	ActionTypeDelayCloseBin = 0b111110
)

// 0        1      2           3                                                             18         19
// +--------+------+-----------+-------------------------------------------------------------+----------+------+
// | Header | Type | ButtonNo. |                            Value                            | CheckSum | Tail |
// +--------+------+-----------+-------------------------------------------------------------+----------+------+
//
//	  1B       1B       1B                                   15B                                  1B       1B
//	Svr->C:                                                                                              Svr->C:
//	 0xAA                                                                                                 0xBB
//	C->Svr:                                                                                              C->Svr:
//	 0xCC                                                                                                 0xDD
type RawPacketStruct struct {
	sync.RWMutex
	Header   byte
	Type     byte
	ButtonNo byte
	Value    []byte
	CheckSum byte
	Tail     byte
}

type PacketStruct struct {
	sync.RWMutex
	Header   string
	Type     string
	ButtonNo uint
	Value    []byte
	CheckSum byte
	Tail     string
}

// Struct to packet
func (me *RawPacketStruct) Marshal() []byte {
	me.Lock()
	defer me.Unlock()
	packet := make([]byte, MaxSize)
	packet = append(packet, me.Header)
	packet = append(packet, me.Type)
	packet = append(packet, me.ButtonNo)
	packet = append(packet, me.Value...)
	packet = append(packet, me.CheckSum)
	packet = append(packet, me.Tail)
	return packet
}

// Packet to Struct
func (me *RawPacketStruct) UnMarshal(raw []byte) error {
	me.Lock()
	defer me.Unlock()
	if len(raw) > SinglePkgMaxSize {
		return errors.New("raw packet size error")
	}
	index := 0
	me.Header = raw[0]
	index += HeaderSize
	me.Type = raw[1]
	index += TypeSize
	me.ButtonNo = raw[2]
	index += BtnNumSize
	me.Value = raw[index : index+ValueSize]
	index += ValueSize
	me.CheckSum = raw[18]
	index += ChecksumSize
	me.Tail = raw[19]

	return nil
}

func (me *PacketStruct) ParsePacket(raw *RawPacketStruct) {
	me.Lock()
	defer me.Unlock()

	if raw.Header == HeaderClientToSvrBin {
		me.Header = "ClientToServer"
	} else if raw.Header == HeaderSvrToClientBin {
		me.Header = "ServerToClient"
	}

	if raw.Type == ActionTypeAllCloseBin {
		me.Type = "AllClose"
	} else if raw.Type == ActionTypeAllOpenBin {
		me.Type = "AllOpen"
	} else if raw.Type == ActionTypeSingleOpenBin {
		me.Type = "SingleOpen"
	} else if raw.Type == ActionTypeSingleCloseBin {
		me.Type = "SingleClose"
	} else if raw.Type == ActionTypeQueryBin {
		me.Type = "QueryStatus"
	} else if raw.Type == ActionTypeRegisteStatusBin {
		me.Type = "RegisteStatus"
	}

	me.ButtonNo = uint(me.ButtonNo)
	me.Value = raw.Value
	me.CheckSum = raw.CheckSum

	if raw.Header == TailClientToSvrBin {
		me.Tail = "ClientToServer"
	} else if raw.Header == TailSvrToClientBin {
		me.Tail = "ServerToClient"
	}

}

func (me *PacketStruct) UnParsePacket(data *RawPacketStruct) {
	me.Lock()
	defer me.Unlock()

	if me.Header == "ClientToServer" {
		data.Header = HeaderClientToSvrBin
	} else if me.Header == "ServerToClient" {
		data.Header = HeaderSvrToClientBin
	}

	if me.Type == "AllClose" {
		data.Type = ActionTypeAllCloseBin
	} else if me.Type == "AllOpen" {
		data.Type = ActionTypeAllOpenBin
	} else if me.Type == "SingleOpen" {
		data.Type = ActionTypeSingleOpenBin
	} else if me.Type == "SingleClose" {
		data.Type = ActionTypeSingleCloseBin
	} else if me.Type == "QueryStatus" {
		data.Type = ActionTypeQueryBin
	} else if me.Type == "RegisteStatus" {
		data.Type = ActionTypeRegisteStatusBin
	}

	data.ButtonNo = byte(me.ButtonNo)
	data.Value = me.Value
	data.CheckSum = me.CheckSum

	if me.Tail == "ClientToServer" {
		data.Header = TailClientToSvrBin
	} else if me.Tail == "ServerToClient" {
		data.Header = TailSvrToClientBin
	}

}
