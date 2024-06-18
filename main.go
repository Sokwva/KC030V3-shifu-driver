package main

import (
	"fmt"
	"os"
	"sokwva/KC030V3-shifu-driver/server/httpSvr"
	"sokwva/KC030V3-shifu-driver/server/mqtt"
	"sokwva/KC030V3-shifu-driver/utils"

	"github.com/urfave/cli/v2"
)

var (
	target      string = ""
	healthCheck string = ""
	//环境：container or host 容器环境下的健康检测若遇到错误会直接panic出错，主机环境下只是打印错误信息
	enviroment string = "container"
	serverType string = "http"

	mqttAddr            string = ""
	mqttUser            string = ""
	mqttPass            string = ""
	mqttName            string = ""
	mqttParentTopicPath string = ""
	logLevel            string = ""
)

func main() {
	initCli()
}

func initCli() {
	cliApp := &cli.App{
		Name:  "KC030V3-shifu-driver",
		Usage: "KC030V3-shifu-driver [options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target",
				Value:       "192.168.1.20:8080",
				Usage:       "target and port of target device.",
				Destination: &target,
			},
			&cli.StringFlag{
				Name:        "check",
				Value:       "false",
				Usage:       "loop to check target device healthy.",
				Destination: &healthCheck,
			},
			&cli.StringFlag{
				Name:        "env",
				Value:       "container",
				Usage:       "different enviroment of different unhealthy action.",
				Destination: &enviroment,
			},
			&cli.StringFlag{
				Name:        "server",
				Value:       "http",
				Usage:       "server app layer handler type: http or mqtt",
				Destination: &serverType,
			},
			&cli.StringFlag{
				Name:        "mqttAddr",
				Value:       "",
				Usage:       "mqtt broker ip and port",
				Destination: &mqttAddr,
			},
			&cli.StringFlag{
				Name:        "mqttUser",
				Value:       "",
				Usage:       "mqtt broker user",
				Destination: &mqttUser,
			},
			&cli.StringFlag{
				Name:        "mqttPass",
				Value:       "",
				Usage:       "mqtt broker password",
				Destination: &mqttPass,
			},
			&cli.StringFlag{
				Name:        "mqttName",
				Value:       "",
				Usage:       "mqtt client name",
				Destination: &mqttName,
			},
			&cli.StringFlag{
				Name:        "mqttParentTopicPath",
				Value:       "shifu/dev",
				Usage:       "mqtt parent topic path",
				Destination: &mqttParentTopicPath,
			},
			&cli.StringFlag{
				Name:        "logLevel",
				Value:       "error",
				Usage:       "log level",
				Destination: &logLevel,
			},
		},
		Action: func(ctx *cli.Context) error {
			utils.InitLogger(logLevel)
			if serverType == "http" {
				httpSvr.SetParam(target, healthCheck, enviroment)
				httpSvr.Serve()
			}
			if serverType == "mqtt" {
				mqtt.SetParam(target, healthCheck, enviroment, mqttAddr, mqttUser, mqttPass, mqttParentTopicPath, mqttName)
				mqtt.Serve()
			}
			return nil
		},
	}
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}
}
