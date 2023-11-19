package cli

import (
	"fmt"
	"os"
	"sokwva/KC030V3-shifu-driver/client"
	"sokwva/KC030V3-shifu-driver/serializer"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	cliApp := &cli.App{
		Name:  "KC030V3-shifu-driver.exe",
		Usage: "KC030V3-shifu-driver.exe [options]",
		Commands: []*cli.Command{
			{
				Name:     "get",
				Usage:    "get 192.168.1.2:8080",
				Category: "basic get",
				Action: func(ctx *cli.Context) error {
					var cliQueryTarget string = ctx.Args().First()
					if cliQueryTarget != "" {
						if len(strings.Split(cliQueryTarget, ":")) != 2 {
							fmt.Println("wrong param,exp: 192.168.1.2:8080")
							return nil
						}
						queryCmd := &serializer.PacketStruct{
							Header:   "ClientToServer",
							Type:     "QueryStatus",
							ButtonNo: 0,
							Value:    []byte{},
							CheckSum: 0,
							Tail:     "ClientToServer",
						}
						queryCmdRaw := &serializer.RawPacketStruct{}
						queryCmd.UnParsePacket(queryCmdRaw)
						b := queryCmdRaw.Marshal()
						a, err := client.Send(cliQueryTarget, b)
						if err != nil {
							fmt.Println(err)
							return nil
						}
						fmt.Println("rawByte:\n", a)
						respRaw := &serializer.RawPacketStruct{}
						respRaw.UnMarshal(a)
						fmt.Println("resp: \n", respRaw)
						resp := &serializer.PacketStruct{}
						resp.ParsePacket(respRaw)
						fmt.Println("resp: \n", resp)
						status, err := serializer.ActionAllRespParse(resp.Value)
						if err != nil {
							return err
						}
						fmt.Println("switch status: \n", status)

					} else {
						fmt.Println("no param")
					}
					return nil
				},
			},
		},
	}
	err := cliApp.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}

}
