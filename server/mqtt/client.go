package mqtt

import (
	"log"

	mqttDrv "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqttAddr            string = ""
	mqttUser            string = ""
	mqttPass            string = ""
	mqttName            string = ""
	mqttParentTopicPath string = ""
)

func Serve() {
	options := mqttDrv.NewClientOptions()
	options.AddBroker("tcp://" + mqttAddr)
	options.SetClientID(mqttName)

	mqttClient := mqttDrv.NewClient(options)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
}
