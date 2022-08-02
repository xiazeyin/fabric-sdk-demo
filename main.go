package main

import (
	"fmt"
	"github.com/xiazeyin/fabric-sdk-demo/sdkInit"
	"log"
	"time"

	"os"
)

const (
	cc_name    = "nft"
	cc_version = "1.0.0"
)

var App sdkInit.Application

func main() {
	// init orgs information

	orgs := []*sdkInit.OrgInfo{
		{
			OrgAdminUser:  "Admin",
			OrgName:       "Org1",
			OrgMspId:      "Org1MSP",
			OrgUser:       "User1",
			OrgPeerNum:    1,
			OrgAnchorFile: "/data/gopath/src/fabric-sdk-demo/fixtures/channel-artifacts/Org1MSPanchors.tx",
		},
		{
			OrgAdminUser:  "Admin",
			OrgName:       "Org2",
			OrgMspId:      "Org2MSP",
			OrgUser:       "User1",
			OrgPeerNum:    1,
			OrgAnchorFile: "/data/gopath/src/fabric-sdk-demo/fixtures/channel-artifacts/Org2MSPanchors.tx",
		},
	}

	// init sdk env info
	info := sdkInit.SdkEnvInfo{
		ChannelID:        "mychannel",
		ChannelConfig:    "/data/gopath/src/fabric-sdk-demo/fixtures/channel-artifacts/mychannel.tx",
		Orgs:             orgs,
		OrdererAdminUser: "Admin",
		OrdererOrgName:   "OrdererOrg",
		OrdererEndpoint:  "orderer.example.com",
		ChaincodeID:      cc_name,
		ChaincodePath:    "/data/gopath/src/fabric-sdk-demo/chaincode/",
		ChaincodeVersion: cc_version,
	}

	// sdk setup
	sdk, err := sdkInit.Setup("config.yaml", &info)
	if err != nil {
		fmt.Println(">> SDK setup error:", err)
		os.Exit(-1)
	}

	// create channel and join
	if err := sdkInit.CreateAndJoinChannel(&info); err != nil {
		fmt.Println(">> Create channel and join error:", err)
		os.Exit(-1)
	}

	// create chaincode lifecycle
	if err := sdkInit.CreateCCLifecycle(&info, 1, false, sdk); err != nil {
		fmt.Println(">> create chaincode lifecycle error:", err)
		os.Exit(-1)
	}

	// invoke chaincode set status
	fmt.Println(">> 通过链码外部服务设置链码状态......")

	if err := info.InitService(info.ChaincodeID, info.ChannelID, info.Orgs[0], sdk); err != nil {

		fmt.Println("InitService successful")
		os.Exit(-1)
	}

	App = sdkInit.Application{
		SdkEnvInfo: &info,
	}
	fmt.Println(">> 设置链码状态完成")

	defer info.EvClient.Unregister(sdkInit.BlockListener(info.EvClient))
	defer info.EvClient.Unregister(sdkInit.ChainCodeEventListener(info.EvClient, info.ChaincodeID))

	a := []string{"initLedger", "qychain", "qy", "0xa7Ad9207DE1417198FCc62FbF1a16B7f35C96ABf", "0"}
	ret, err := App.Set(a)
	if err != nil {
		log.Println(err)
	}
	log.Println("<--- 添加信息　--->：", ret)

	a = []string{"mine", "qy", "0xa7Ad9207DE1417198FCc62FbF1a16B7f35C96ABf", "0x31d837ce79219cf256ed105e9cc0c40c1422b2405a8f2366df0cae3c7ac57aa9", "https://nftimg.stars-mine.com/1e3de73d-aabc-11ec-b4dd-0242ac110002.jpeg", "xx", "RSA"}
	ret, err = App.Set(a)
	if err != nil {
		log.Println(err)
	}
	log.Println("<--- 添加信息　--->：", ret)

	a = []string{"ownerof", "qy", "0x31d837ce79219cf256ed105e9cc0c40c1422b2405a8f2366df0cae3c7ac57aa9"}
	response, err := App.Set(a)
	if err != nil {
		log.Println(err)
	}
	log.Println("<--- 查询信息　--->：", response)

	time.Sleep(time.Second * 10)
}
