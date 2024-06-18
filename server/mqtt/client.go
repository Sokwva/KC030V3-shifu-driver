package mqtt

import (
	"encoding/json"
	"log"
	"sokwva/KC030V3-shifu-driver/client"
	"sokwva/KC030V3-shifu-driver/serializer"
	"sokwva/KC030V3-shifu-driver/utils"
	"strconv"
	"strings"
	"time"

	mqttDrv "github.com/eclipse/paho.mqtt.golang"
)

var (
	target              string = ""
	mqttAddr            string = ""
	mqttUser            string = ""
	mqttPass            string = ""
	mqttName            string = ""
	mqttParentTopicPath string = ""
	mqttClient          mqttDrv.Client
	healthCheck         string = ""
	enviroment          string = ""

	checkChan chan bool = make(chan bool)
)

func SetParam(targetExt string, healthCheckExt string, enviromentExt string, add string, user string, pass string, parentPath string, name string) {
	target = targetExt
	healthCheck = healthCheckExt
	enviroment = enviromentExt
	mqttAddr = add
	mqttUser = user
	mqttPass = pass
	mqttParentTopicPath = parentPath
	mqttName = name
}

func Serve() {
	options := mqttDrv.NewClientOptions()
	options.AddBroker("tcp://" + mqttAddr)
	options.SetClientID(mqttName)
	options.SetUsername(mqttUser)
	options.SetPassword(mqttPass)

	mqttClient = mqttDrv.NewClient(options)
	defer mqttClient.Disconnect(3)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	Pub(mqttName + " driver is ready.")
	Sub()
	if healthCheck == "true" {
		go checkHealthy()
	}
	<-checkChan
}

func Pub(payload string) bool {
	token := mqttClient.Publish(mqttParentTopicPath+mqttName, 0, false, payload)
	return token.Wait()
}

func Sub() bool {
	topic := mqttParentTopicPath + mqttName
	token := mqttClient.Subscribe(topic, 0, EventHandler)
	return token.Wait()
}

func commonBody(queryCmd *serializer.PacketStruct, msg mqttDrv.Message) *serializer.PacketStruct {
	queryCmdRaw := &serializer.RawPacketStruct{}
	queryCmd.UnParsePacket(queryCmdRaw)
	b := queryCmdRaw.Marshal()
	a, err := client.Send(target, b)
	if err != nil {
		Pub(string(msg.Payload()) + " send faild.")
		return nil
	}
	respRaw := &serializer.RawPacketStruct{}
	err = respRaw.UnMarshal(a)
	if err != nil {
		Pub(string(msg.Payload()) + " response struct raw parse faild.")
		return nil
	}
	resp := &serializer.PacketStruct{}
	resp.ParsePacket(respRaw)
	return resp
}

func commonAction(baseicCmd string, cmds []string, msg mqttDrv.Message) {
	btnNum, err := strconv.Atoi(cmds[1])
	if err != nil {
		Pub(string(msg.Payload()) + " param button number(param 2) is not valid uint.")
	}
	queryCmd := &serializer.PacketStruct{
		Header:   "ServerToClient",
		Type:     baseicCmd,
		ButtonNo: uint(btnNum) - 1,
		Value:    []byte{},
		CheckSum: 0,
		Tail:     "ServerToClient",
	}
	resp := commonBody(queryCmd, msg)
	status, err := serializer.ActionAllRespParse(resp.Value)
	if err != nil {
		Pub(string(msg.Payload()) + " response struct parse faild.")
		return
	}
	respMsg, err := json.Marshal(status)
	if err != nil {
		Pub(string(msg.Payload()) + " response parse faild.")
	}
	Pub(string(respMsg))
}

func EventHandler(ctx mqttDrv.Client, msg mqttDrv.Message) {
	//common command: [cmd] [target]
	//cmd: get close open
	//target(generally relay number):0->all,1-5->relay number,<empty>->query
	utils.Log.Debug("event", "from-topic", string(msg.Topic()), "payload:", string(msg.Payload()))
	if len(msg.Payload()) == 0 || len(msg.Payload()) > 10 {
		utils.Log.Info("event ignore invalid length message")
		return
	}
	cmds := strings.Split(string(msg.Payload()), " ")
	switch cmds[0] {
	case "get":
		queryCmd := &serializer.PacketStruct{
			Header:   "ClientToServer",
			Type:     "QueryStatus",
			ButtonNo: 0,
			Value:    []byte{},
			CheckSum: 0,
			Tail:     "ClientToServer",
		}
		resp := commonBody(queryCmd, msg)
		status, err := serializer.ActionAllRespParse(resp.Value)
		if err != nil {
			Pub(string(msg.Payload()) + " response struct parse faild.")
			return
		}
		respMsg, err := json.Marshal(status)
		if err != nil {
			Pub(string(msg.Payload()) + " response parse faild.")
		}
		Pub(string(respMsg))
	case "close":
		if cmds[1] == "" {
			Pub(string(msg.Payload()) + " param button number(param 2) is empty.")
		}
		if cmds[1] == "0" {
			commonAction("AllClose", cmds, msg)
			return
		}
		commonAction("SingleClose", cmds, msg)
	case "open":
		if cmds[1] == "" {
			Pub(string(msg.Payload()) + " param button number(param 2) is empty.")
		}
		if cmds[1] == "0" {
			commonAction("AllOpen", cmds, msg)
			return
		}
		commonAction("SingleOpen", cmds, msg)
	}
}

func checkHealthy() {
	timer := time.NewTicker(time.Second * 10)
	for {
		utils.Log.Info("checking...")
		<-timer.C
		if utils.ProbeTCP(target) {
			continue
		} else {
			if enviroment == "host" {
				utils.Log.Info("device is not healthy.")
				Pub("device " + target + " is unhealthy.")
			}
			if enviroment == "container" {
				checkChan <- false
				panic("device is not healthy.")
			}
		}
	}
}
