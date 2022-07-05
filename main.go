package main

import (
	"fmt"
	"github.com/xiazeyin/fabric-sdk-demo/sdkInit"
	"os"
	"time"
)

const (
	cc_name    = "simplecc"
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
	//if err := sdkInit.CreateAndJoinChannel(&info); err != nil {
	//	fmt.Println(">> Create channel and join error:", err)
	//	os.Exit(-1)
	//}

	// create chaincode lifecycle
	if err := sdkInit.CreateCCLifecycle(&info, 1, false, sdk); err != nil {
		fmt.Println(">> create chaincode lifecycle error: %v", err)
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

	a := []string{"set", "ID1", "123"}
	ret, err := App.Set(a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("<--- 添加信息　--->：", ret)

	a = []string{"set", "ID2", "456"}
	ret, err = App.Set(a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("<--- 添加信息　--->：", ret)

	a = []string{"set", "ID3", "789"}
	ret, err = App.Set(a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("<--- 添加信息　--->：", ret)

	a = []string{"get", "ID3"}
	response, err := App.Get(a)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("<--- 查询信息　--->：", response)

	time.Sleep(time.Second * 10)

}

//import (
//	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/channel"
//	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/resmgmt"
//	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/fabsdk"
//	"os"
//)
//
//type FabricModel struct {
//	ConfigFile string // sdk的配置文件路径
//	ChainCodeID string	// 链码名称
//	ChaincodePath string	// 链码在工程中的存放目录
//	ChaincodeGoPath string	//
//	OrgAdmin string		// 组织的管理员用户
//	OrgName string		// config.yaml ---> organizations ---> travle
//	OrgID 	string		// 组织id
//	UserName string		// 组织的普通用户
//	ChannelID string	// 通道id
//	ChannelConfigPath string	// 组织的通道文件路径
//	OrdererName string	// config.yaml ---> orderers ---> orderer.xq.com // 将组织添加到通道时使用
//	Sdk *fabsdk.FabricSDK	// 保存实例化后的sdk
//	ResMgmtCli *resmgmt.Client	// 资源管理客户端,也需要在安装链码时候的使用
//	Channelclient *channel.Client	// 通道客户端
//	HasInit	bool		// 是否已经初始化了sdk
//}
//
//func init()  {
//	fs := FabricModel{
//		OrdererName: "orderer.xq.com",
//		ChannelID: "travlechannel",
//		ChannelConfigPath: os.Getenv("GOPATH") + "/src/driverFabricDemo/conf/channel-artifacts/travelchannel.tx",
//		ChainCodeID: "mycc",
//		ChaincodeGoPath: os.Getenv("GOPATH"),
//		ChaincodePath: "driverFabricDemo/chaincode",
//		OrgAdmin: "Admin",
//		OrgName: "travel",
//		ConfigFile: "conf/config.yaml",
//		UserName: "User1",
//	}
//	//fs.Initialization()
//	fs.HasInit = true
//}
//
//func main()  {
//	//sdk, err := fabsdk.New(config.FromFile())
//	//if err != nil {
//	//	t.Fatalf("Failed to create new SDK: %s", err)
//	//}
//	//defer sdk.Close()
//}
