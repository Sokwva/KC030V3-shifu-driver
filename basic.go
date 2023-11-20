package main

import (
	"fmt"
	"net/http"
	"os"
	"sokwva/KC030V3-shifu-driver/client"
	"sokwva/KC030V3-shifu-driver/serializer"
	"sokwva/KC030V3-shifu-driver/server"
	"sokwva/KC030V3-shifu-driver/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

var (
	target      string = ""
	healthCheck string = ""
	//环境：container or host 容器环境下的健康检测若遇到错误会直接panic出错，主机环境下只是打印错误信息
	enviroment string = "container"
)

func commonTail(queryCmd *serializer.PacketStruct, c *gin.Context) {
	queryCmdRaw := &serializer.RawPacketStruct{}
	queryCmd.UnParsePacket(queryCmdRaw)
	b := queryCmdRaw.Marshal()
	a, err := client.Send(target, b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": server.INTERNAL_ERROR,
			"msg":  "send error",
		})
		return
	}
	respRaw := &serializer.RawPacketStruct{}
	err = respRaw.UnMarshal(a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": server.INTERNAL_ERROR,
			"msg":  "packet parse error",
		})
		return
	}
	resp := &serializer.PacketStruct{}
	resp.ParsePacket(respRaw)
	status, err := serializer.ActionAllRespParse(resp.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": server.INTERNAL_ERROR,
			"msg":  "resp parse error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": server.SUCCESS,
		"msg":  status.Status[:5],
	})
}

func getStatus(c *gin.Context) {
	queryCmd := &serializer.PacketStruct{
		Header:   "ClientToServer",
		Type:     "QueryStatus",
		ButtonNo: 0,
		Value:    []byte{},
		CheckSum: 0,
		Tail:     "ClientToServer",
	}
	commonTail(queryCmd, c)
}

func singoleBtnAction(c *gin.Context, act string) {
	btn, ok := c.GetQuery("btn")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": server.PARAM_ERROR,
			"msg":  "params error",
		})
		return
	}
	btnNum, err := strconv.Atoi(btn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": server.PARAM_ERROR,
			"msg":  "param btn error",
		})
		return
	}
	if btnNum > 5 || btnNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": server.PARAM_ERROR,
			"msg":  "param btn range error",
		})
		return
	}
	queryCmd := &serializer.PacketStruct{
		Header:   "ClientToServer",
		Type:     act,
		ButtonNo: uint(btnNum) - 1,
		Value:    []byte{},
		CheckSum: 0,
		Tail:     "ClientToServer",
	}
	commonTail(queryCmd, c)
}

func allBtnAction(c *gin.Context, act string) {
	queryCmd := &serializer.PacketStruct{
		Header:   "ClientToServer",
		Type:     act,
		ButtonNo: 0,
		Value:    []byte{},
		CheckSum: 0,
		Tail:     "ClientToServer",
	}
	commonTail(queryCmd, c)
}

func close(c *gin.Context) {
	singoleBtnAction(c, "SingleClose")
}

func open(c *gin.Context) {
	singoleBtnAction(c, "SingleOpen")
}

func closeAll(c *gin.Context) {
	allBtnAction(c, "AllClose")
}

func openAll(c *gin.Context) {
	allBtnAction(c, "AllOpen")
}

func probeDevice(c *gin.Context) {
	if utils.ProbeTCP(target) {
		c.JSON(http.StatusOK, gin.H{
			"code":    server.DEVICE_HEALHY,
			"message": "pong",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    server.DEVICE_DISCONNECTED,
			"message": "fail",
		})
	}
}

func checkHealthy() {
	timer := time.NewTicker(time.Second * 10)
	for {
		<-timer.C
		if utils.ProbeTCP(target) {
			continue
		} else {
			if enviroment == "container" {
				//容器环境下可以直接panic
				panic("device is not healthy")
			}
			if enviroment == "host" {
				//主机环境下可以直接打印错误信息
				fmt.Println("device is not healthy")
			}
		}
	}
}

func serve() {
	if healthCheck == "true" {
		go checkHealthy()
	}

	ginRoot := gin.Default()

	ginRoot.GET("/ping", probeDevice)
	ginRoot.GET("/status", getStatus)
	ginRoot.GET("/close", close)
	ginRoot.GET("/open", open)
	ginRoot.GET("/closeAll", closeAll)
	ginRoot.GET("/openAll", openAll)

	ginRoot.Run(":8080")
}

func main() {
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
		},
		Action: func(ctx *cli.Context) error {
			serve()
			return nil
		},
	}
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}
}
