package mdgwmsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	proc "ssevss/datas"
	msg "ssevss/message"
	aa "ssevss/utils"
)

const (

	//消息类型
	LOGINMSG_TYPE = "S001"
	LOGINOUT_TYPE = "S002"
	//消息字符串填充
	SenderCompID     = "CSI"
	TargetCompID     = "SSE"
	AppVerID         = "1.00"
	MsgType_LEN      = 4
	SenderCompID_LEN = 32
	TargetCompID_LEN = 32
	AppVerID_LEN     = 8
	LOGOUTTXT_LEN    = 256
	//消息体头部长度
	MSGHEADER_LEN = 24
	//消息体长度
	LOGINMSG_BODY_LEN = 74
	MSGTAIL_LEN       = 4
	LOGINMSG_LEN      = MSGHEADER_LEN + LOGINMSG_BODY_LEN + MSGTAIL_LEN
	UINT64_LEN        = 8
	UINT32_LEN        = 4
	UINT16_LEN        = 2
)

//创建MDGW消息结构体
type MDGWMsg interface {
	//获取消息类型
	GetMsgType() [MsgType_LEN]byte
}

type MsgHeader struct {
	MsgType      [MsgType_LEN]byte
	SendingTtime uint64
	MsgSeq       uint64
	BodyLength   uint32
}

type MsgTail struct {
	CheckSum uint32
}

//登录消息
type LoginMsg struct {
	MsgHeader
	SenderCompID [SenderCompID_LEN]byte
	TargetCompID [TargetCompID_LEN]byte
	HeartBtInt   uint16
	AppVerID     [AppVerID_LEN]byte
	MsgTail
}

func (loginMsg *LoginMsg) GetMsgType() [MsgType_LEN]byte {
	return loginMsg.MsgType
}

//注销消息
type LogoutMsg struct {
	MsgHeader
	SessionStatus uint32
	Text          [LOGOUTTXT_LEN]byte
	MsgTail
}

//实现MDGWMsg接口
func (logoutMsg *LogoutMsg) GetMsgType() [MsgType_LEN]byte {
	return logoutMsg.MsgType
}

//心跳消息
type HeartBtMsg struct {
	MsgHeader
	MsgTail
}

//实现MDGWMsg接口
func (heartBtMsg *HeartBtMsg) GetMsgType() [MsgType_LEN]byte {
	return heartBtMsg.MsgType
}

//市场状态消息
type MktStatusMsg struct {
	MsgHeader
	SecurityType     uint8
	TradSesMode      uint8
	TradingSessionID [8]byte
	TotNoRelatedSym  uint32
	MsgTail
}

//实现MDGWMsg接口
func (mktStatusMsg *MktStatusMsg) GetMsgType() [MsgType_LEN]byte {
	return mktStatusMsg.MsgType
}

//行情快照
type HqSnapMsg struct {
	MsgHeader
	SecurityType      uint8
	TradSesMode       uint8
	TradeDate         uint32
	LastUpdateTime    uint32
	MDStreamID        [5]byte
	SecurityID        [8]byte
	Symbol            [8]byte
	PreClosePx        uint64
	TotalVolumeTraded uint64
	NumTrades         uint64
	TotalValueTraded  uint64
	TradingPhaseCode  [8]byte
}

//指数行情快照
//根据条目个数需要进行扩展
type IndexSnapExt struct {
	SnapData    HqSnapMsg
	NoMDEntries uint16
	MDEntryType [2]byte
	MDEntryPx   uint64
}

//竞价行情快照
type BidSnapExt struct {
	SnapData          HqSnapMsg
	NoMDEntries       uint16
	MDEntryType       [2]byte
	MDEntryPx         uint64
	MDEntrySize       uint64
	MDEntryPositionNo uint8
}

//初始化
func initLoginMsg(loginMsg *LoginMsg) {

	//按照接口规范初始化char字符串类型，通过空格填充
	for i, _ := range loginMsg.SenderCompID {
		loginMsg.SenderCompID[i] = ' '
	}

	for i, _ := range loginMsg.TargetCompID {
		loginMsg.TargetCompID[i] = ' '
	}

	for i, _ := range loginMsg.AppVerID {
		loginMsg.AppVerID[i] = ' '
	}

}

//填充消息体
func setLoginMsgBody(loginMsg *LoginMsg) {
	//填充发送ID
	var setStr []byte
	setStr = []byte(SenderCompID)
	for i, c := range setStr {
		loginMsg.SenderCompID[i] = c
	}

	//填充目标ID
	setStr = []byte(TargetCompID)
	for i, c := range setStr {
		loginMsg.TargetCompID[i] = c
	}

	//填充APP
	setStr = []byte(AppVerID)
	for i, c := range setStr {
		loginMsg.AppVerID[i] = c
	}

	//填充心跳时间
	loginMsg.HeartBtInt = 1
}

func setLoginMsgHeader(loginMsg *LoginMsg, sendingTtime, msgSeq uint64) {
	//填充消息类型
	var setStr []byte
	setStr = []byte(LOGINMSG_TYPE)
	for i, c := range setStr {
		loginMsg.MsgType[i] = c
	}

	//填充消息序号
	loginMsg.MsgSeq = msgSeq
	msgSeq = msgSeq + 1

	//填充发送时间
	//获取当前时间
	loginMsg.SendingTtime = sendingTtime

	//消息体长度
	loginMsg.BodyLength = LOGINMSG_BODY_LEN
}

//计算登录消息校验和并填充MsgTail字段，返回字节数组buffer，后续进行发送
func calLoginMsgChkSum(loginMsg *LoginMsg) *bytes.Buffer {
	//将数据包中的字段放入到字节数组中，计算校验和
	buf := new(bytes.Buffer)
	//写入消息类型
	buf.Write(loginMsg.MsgType[:])
	//按照大端方式写入整数
	//写入发送时间
	binary.Write(buf, binary.BigEndian, loginMsg.SendingTtime)
	//写入消息序号
	binary.Write(buf, binary.BigEndian, loginMsg.MsgSeq)
	//写入消息体长度
	binary.Write(buf, binary.BigEndian, loginMsg.BodyLength)
	//写入发送ID
	buf.Write(loginMsg.SenderCompID[:])
	//写入目标ID
	buf.Write(loginMsg.TargetCompID[:])
	//写入心跳时间
	binary.Write(buf, binary.BigEndian, loginMsg.HeartBtInt)
	//写入版本信息
	buf.Write(loginMsg.AppVerID[:])

	//计算
	chksum := aa.CalCheckSum(buf.Bytes(), MSGHEADER_LEN+LOGINMSG_BODY_LEN)
	loginMsg.CheckSum = chksum
	//写入校验和
	binary.Write(buf, binary.BigEndian, loginMsg.CheckSum)
	return buf
}

//创建登录消息
func NewLoginMsg(sendingTtime, msgSeq uint64) (*LoginMsg, *bytes.Buffer) {
	loginMsg := &LoginMsg{}
	//初始化登录消息
	initLoginMsg(loginMsg)
	//填充消息体
	setLoginMsgBody(loginMsg)
	//填充消息头部
	setLoginMsgHeader(loginMsg, sendingTtime, msgSeq)
	//计算数据包校验和，并填充校验值
	buf := calLoginMsgChkSum(loginMsg)
	return loginMsg, buf
}

func GetMsgHeader(msgHeader *MsgHeader, b []byte, len int) {
	//获取消息头部
	buf := bytes.NewReader(b)
	fmt.Println("mesage header buf:", buf.Len())

	err := binary.Read(buf, binary.BigEndian, msgHeader)

	if err != nil {
		fmt.Println("get message header failed:", err)
	}
}

//通过byte序列，获取一个消息
func GetMsgFromBytes(b []byte, msglen int) MDGWMsg {

	buf := bytes.NewReader(b)
	//根据消息类型，登录成功消息
	if bytes.Equal(b[:MsgType_LEN], []byte(LOGINMSG_TYPE)) {
		loginMsg := &LoginMsg{}
		fmt.Println("it's login msg")
		err := binary.Read(buf, binary.BigEndian, loginMsg)
		if err != nil {
			fmt.Println("read msg from bytes err:", err)
		}
		return loginMsg
	} else if bytes.Equal(b[:MsgType_LEN], []byte(LOGINOUT_TYPE)) {
		logoutMsg := &LogoutMsg{}
		fmt.Println("it's logout msg")
		err := binary.Read(buf, binary.BigEndian, logoutMsg)
		if err != nil {
			fmt.Println("read msg from bytes err:", err)
		}

		return logoutMsg
	}

	return nil

}

//解析消息
func ParseMsg(b []byte, msglen int) {
	mdgwmsg := GetMsgFromBytes(b, msglen)
	switch v := mdgwmsg.(type) {
	case *msg.LoginMsg:
		fmt.Println("verify mdgw get msg is loginMsg", v.MsgType)
		proc.ProcLoginMsg(v)
	case *msg.LogoutMsg:
		fmt.Println("verify mdgw get msg is logoutMsg", v.MsgType)
		proc.ProcLogoutMsg(v)
	case *msg.MktStatusMsg:
		fmt.Println("market status msg:", v.MsgType)
	default:
		fmt.Printf("other msg type")

	}
}
