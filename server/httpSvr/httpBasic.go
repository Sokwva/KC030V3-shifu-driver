package httpSvr

import (
	"fmt"
	"net/http"
	"sokwva/KC030V3-shifu-driver/client"
	"sokwva/KC030V3-shifu-driver/serializer"
	"sokwva/KC030V3-shifu-driver/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	target      string = ""
	healthCheck string = ""
	enviroment  string = "container"
)

func commonTail(queryCmd *serializer.PacketStruct, c *gin.Context) {
	queryCmdRaw := &serializer.RawPacketStruct{}
	queryCmd.UnParsePacket(queryCmdRaw)
	b := queryCmdRaw.Marshal()
	a, err := client.Send(target, b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": INTERNAL_ERROR,
			"msg":  "send error",
		})
		return
	}
	respRaw := &serializer.RawPacketStruct{}
	err = respRaw.UnMarshal(a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": INTERNAL_ERROR,
			"msg":  "packet parse error",
		})
		return
	}
	resp := &serializer.PacketStruct{}
	resp.ParsePacket(respRaw)
	status, err := serializer.ActionAllRespParse(resp.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": INTERNAL_ERROR,
			"msg":  "resp parse error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
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
			"code": PARAM_ERROR,
			"msg":  "params error",
		})
		return
	}
	btnNum, err := strconv.Atoi(btn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": PARAM_ERROR,
			"msg":  "param btn error",
		})
		return
	}
	if btnNum > 5 || btnNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": PARAM_ERROR,
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
			"code":    DEVICE_HEALHY,
			"message": "pong",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    DEVICE_DISCONNECTED,
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

func ServeSetParam(targetExt string, healthCheckExt string, enviromentExt string) {
	target = targetExt
	healthCheck = healthCheckExt
	enviroment = enviromentExt
}

func Serve() {
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
