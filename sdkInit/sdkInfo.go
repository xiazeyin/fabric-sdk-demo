package sdkInit

import (
	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/channel"
	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/event"
	mspclient "github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/msp"
	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/resmgmt"
	contextAPI "github.com/xiazeyin/fabric-sdk-go-gm/pkg/common/providers/context"
)

type OrgInfo struct {
	OrgAdminUser          string // like "Admin"
	OrgName               string // like "Org1"
	OrgMspId              string // like "Org1MSP"
	OrgUser               string // like "User1"
	orgMspClient          *mspclient.Client
	OrgAdminClientContext *contextAPI.ClientProvider
	OrgResMgmt            *resmgmt.Client
	OrgPeerNum            int
	//Peers                 []*fab.Peer
	OrgAnchorFile string // like ./channel-artifacts/Org2MSPanchors.tx
}

type SdkEnvInfo struct {
	// 通道信息
	ChannelID     string // like "simplecc"
	ChannelConfig string // like os.Getenv("GOPATH") + "/src/github.com/hyperledger/fabric-samples/test-network/channel-artifacts/testchannel.tx"

	// 组织信息
	Orgs []*OrgInfo
	// 排序服务节点信息
	OrdererAdminUser     string // like "Admin"
	OrdererOrgName       string // like "OrdererOrg"
	OrdererEndpoint      string
	OrdererClientContext *contextAPI.ClientProvider
	// 链码信息
	ChaincodeID      string
	ChaincodeGoPath  string
	ChaincodePath    string
	ChaincodeVersion string
	ChClient         *channel.Client
	EvClient         *event.Client
}

type Application struct {
	SdkEnvInfo *SdkEnvInfo
}