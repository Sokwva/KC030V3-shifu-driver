package serializer

import (
	"errors"
	"sokwva/KC030V3-shifu-driver/utils"
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
	//0x0A 全开
	ActionTypeAllOpenBin = 0b00001010
	//0x0B 全关
	ActionTypeAllCloseBin = 0b00001011
	//0x0C 查询状态
	ActionTypeQueryBin = 0b00001100
	//0x0D 单开
	ActionTypeSingleOpenBin = 0b00001101
	//0x0E 单关
	ActionTypeSingleCloseBin = 0b00001110
	//0x1E 向服务器注册端点，是的服务器每7秒向端点报告一次继电器状态
	ActionTypeRegisteStatusBin = 0b00011110
	//0x1A 年 年 月 日 时 分 秒 （年由两部分组成，例如2017：14 11）
	ActionTypeSetDateTimeBin = 0b00011010
	//0x1B
	ActionTypeGetDateTimeBin = 0b00011011
	//0x2b value 3-9分别代表DIO_20到DIO_28这7个IO口的状态，01高电平，02低电平
	ActionTypeGetIOBin = 0b000101011
	//0x2d 批量设置继电器状态 3-7 五路继电器，01开，02关
	ActionTypeBatchSetBin = 0b00101101
	//0x3d
	ActionTypeDelayOpenBin = 0b00111101
	//0x3e
	ActionTypeDelayCloseBin = 0b00111110
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
	Header   byte
	Type     byte
	ButtonNo byte
	Value    []byte
	CheckSum byte
	Tail     byte
}

type PacketStruct struct {
	Header   string
	Type     string
	ButtonNo uint
	Value    []byte
	CheckSum byte
	Tail     string
}

// Struct to packet
func (me *RawPacketStruct) Marshal() []byte {
	utils.Log.Debug("start to Marshal RawPacketStruct", "raw", me)
	packet := make([]byte, SinglePkgMaxSize)
	packet = append(packet, me.Header)
	utils.Log.Debug("raw", "head", me.Header, "length", len(packet))
	packet = append(packet, me.Type)
	utils.Log.Debug("raw", "type", me.Type, "length", len(packet))
	packet = append(packet, me.ButtonNo)
	utils.Log.Debug("raw", "buttonNo", me.Type, "length", len(packet))
	packet = append(packet, me.Value...)
	utils.Log.Debug("raw", "value", me.Type, "length", len(packet))
	packet = append(packet, me.CheckSum)
	utils.Log.Debug("raw", "checksum", me.Type, "length", len(packet))
	packet = append(packet, me.Tail)
	utils.Log.Debug("raw", "tail", me.Type, "length", len(packet))
	utils.Log.Debug("Marshal RawPacketStruct done", "raw", packet, "length", len(packet))
	return packet
}

// Packet to Struct
func (me *RawPacketStruct) UnMarshal(raw []byte) error {
	utils.Log.Debug("start to UnMarshal RawPacketStruct", "raw", raw)
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
	utils.Log.Debug("UnMarshal RawPacketStruct done", "struct", me)
	return nil
}

func (me *PacketStruct) ParsePacket(raw *RawPacketStruct) {
	utils.Log.Debug("start to ParsePacket PacketStruct ", "raw", raw)

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

	if raw.Tail == TailClientToSvrBin {
		me.Tail = "ClientToServer"
	} else if raw.Tail == TailSvrToClientBin {
		me.Tail = "ServerToClient"
	}
	utils.Log.Debug("ParsePacket PacketStruct done", "struct", me)
}

func (me *PacketStruct) UnParsePacket(data *RawPacketStruct) {
	utils.Log.Debug("start to UnParsePacket PacketStruct ", "struct", me)

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
		data.Tail = TailClientToSvrBin
	} else if me.Tail == "ServerToClient" {
		data.Tail = TailSvrToClientBin
	}
	utils.Log.Debug("UnParsePacket PacketStruct done", "data", data)
}
