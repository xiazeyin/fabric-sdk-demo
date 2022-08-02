[TOC]
SDK示例,fabric-sdk-demo(代码位置在https://git.linkyee.com/gschain/fabric-sdk-demo)是SDK使用参照例子，里面包含了创建通道并加入通道如何编写、部署智能合约、智能合约的调用和查询例子,本文档使用两个例子介绍如何使用sdk,第一个例子是创建通道并加入通道,第二个例子是部署合约代码
# 创建通道并加入通道
## 1.实例化sdk
参照fabric-sdk-demo中sdkInit包中的Setup方法
```
func Setup(configFile string, info *SdkEnvInfo) (*fabsdk.FabricSDK, error) {
	// Create SDK setup for the integration tests
	var err error
	sdk, err := fabsdk.New(config.FromFile(configFile))
	if err != nil {
		return nil, err
	}

	// 为组织获得Client句柄和Context信息
	for _, org := range info.Orgs {
		org.orgMspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org.OrgName))
		if err != nil {
			return nil, err
		}
		orgContext := sdk.Context(fabsdk.WithUser(org.OrgAdminUser), fabsdk.WithOrg(org.OrgName))
		org.OrgAdminClientContext = &orgContext

		// New returns a resource management client instance.
		resMgmtClient, err := resmgmt.New(orgContext)
		if err != nil {
			return nil, fmt.Errorf("根据指定的资源管理客户端Context创建通道管理客户端失败: %v", err)
		}
		org.OrgResMgmt = resMgmtClient
	}

	// 为Orderer获得Context信息
	ordererClientContext := sdk.Context(fabsdk.WithUser(info.OrdererAdminUser), fabsdk.WithOrg(info.OrdererOrgName))
	info.OrdererClientContext = &ordererClientContext
	return sdk, nil
}
```
## 2.创建通道
```
func createChannel(signIDs []msp.SigningIdentity, info *SdkEnvInfo) error {
	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(*info.OrdererClientContext)
	if err != nil {
		return fmt.Errorf("Channel management client create error: %v", err)
	}

	// create a channel for orgchannel.tx
	req := resmgmt.SaveChannelRequest{ChannelID: info.ChannelID,
		ChannelConfigPath: info.ChannelConfig,
		SigningIdentities: signIDs}

	if _, err := chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		return fmt.Errorf("error should be nil for SaveChannel of orgchannel: %v", err)
	}

	fmt.Println(">>>> 使用每个org的管理员身份更新锚节点配置...")
	//do the same get ch client and create channel for each anchor peer as well (first for Org1MSP)
	for i, org := range info.Orgs {
		req = resmgmt.SaveChannelRequest{ChannelID: info.ChannelID,
			ChannelConfigPath: org.OrgAnchorFile,
			SigningIdentities: []msp.SigningIdentity{signIDs[i]}}

		if _, err = org.OrgResMgmt.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
			return fmt.Errorf("SaveChannel for anchor org %s error: %v", org.OrgName, err)
		}
	}
	fmt.Println(">>>> 使用每个org的管理员身份更新锚节点配置完成")
	//integration.WaitForOrdererConfigUpdate(t, configQueryClient, mc.channelID, false, lastConfigBlock)
	return nil
}
```
## 组织加入通道
```
func (rc *Client) JoinChannel(channelID string, options ...RequestOption) error {

	if channelID == "" {
		return errors.New("must provide channel ID")
	}

	opts, err := rc.prepareRequestOpts(options...)
	if err != nil {
		return errors.WithMessage(err, "failed to get opts for JoinChannel")
	}

	//resolve timeouts
	rc.resolveTimeouts(&opts)

	//set parent request context for overall timeout
	parentReqCtx, parentReqCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(opts.Timeouts[fab.ResMgmt]), contextImpl.WithParent(opts.ParentContext))
	parentReqCtx = reqContext.WithValue(parentReqCtx, contextImpl.ReqContextTimeoutOverrides, opts.Timeouts)
	defer parentReqCancel()

	targets, err := rc.calculateTargets(opts.Targets, opts.TargetFilter)
	if err != nil {
		return errors.WithMessage(err, "failed to determine target peers for JoinChannel")
	}

	if len(targets) == 0 {
		return errors.WithStack(status.New(status.ClientStatus, status.NoPeersFound.ToInt32(), "no targets available", nil))
	}

	orderer, err := rc.requestOrderer(&opts, channelID)
	if err != nil {
		return errors.WithMessage(err, "failed to find orderer for request")
	}

	ordrReqCtx, ordrReqCtxCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.OrdererResponse), contextImpl.WithParent(parentReqCtx))
	defer ordrReqCtxCancel()

	genesisBlock, err := resource.GenesisBlockFromOrderer(ordrReqCtx, channelID, orderer, resource.WithRetry(opts.Retry))
	if err != nil {
		return errors.WithMessage(err, "genesis block retrieval failed")
	}

	joinChannelRequest := resource.JoinChannelRequest{
		GenesisBlock: genesisBlock,
	}

	peerReqCtx, peerReqCtxCancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeoutType(fab.ResMgmt), contextImpl.WithParent(parentReqCtx))
	defer peerReqCtxCancel()
	err = resource.JoinChannel(peerReqCtx, joinChannelRequest, peersToTxnProcessors(targets), resource.WithRetry(opts.Retry))
	if err != nil {
		return errors.WithMessage(err, "join channel failed")
	}

	return nil
}
```

# 部署合约
## 实例化，参照上一步的实例化步骤
## 部署合约并初始化
参照sdkSetting.go中的CreateCCLifecycle方法
```
func CreateCCLifecycle(info *SdkEnvInfo, sequence int64, upgrade bool, sdk *fabsdk.FabricSDK) error {
	if len(info.Orgs) == 0 {
		return fmt.Errorf("the number of organization should not be zero.")
	}
	// Package cc
	fmt.Println(">> 开始打包链码......")
	label, ccPkg, err := packageCC(info.ChaincodeID, info.ChaincodeVersion, info.ChaincodePath)
	if err != nil {
		return fmt.Errorf("pakcagecc error: %v", err)
	}
	packageID := lcpackager.ComputePackageID(label, ccPkg)
	fmt.Println(">> 打包链码成功")

	// Install cc
	fmt.Println(">> 开始安装链码......")
	if err := installCC(label, ccPkg, info.Orgs); err != nil {
		return fmt.Errorf("installCC error: %v", err)
	}

	// Get installed cc package
	if err := getInstalledCCPackage(packageID, info.Orgs[0]); err != nil {
		return fmt.Errorf("getInstalledCCPackage error: %v", err)
	}

	// Query installed cc
	if err := queryInstalled(packageID, info.Orgs[0]); err != nil {
		return fmt.Errorf("queryInstalled error: %v", err)
	}
	fmt.Println(">> 安装链码成功")

	// Approve cc
	fmt.Println(">> 组织批准智能合约定义......")
	if err := approveCC(packageID, info.ChaincodeID, info.ChaincodeVersion, sequence, info.ChannelID, info.Orgs, info.OrdererEndpoint); err != nil {
		return fmt.Errorf("approveCC error: %v", err)
	}

	// Query approve cc
	if err:=queryApprovedCC(info.ChaincodeID, sequence, info.ChannelID, info.Orgs);err!=nil{
		return fmt.Errorf("queryApprovedCC error: %v", err)
	}
	fmt.Println(">> 组织批准智能合约定义完成")

	// Check commit readiness
	fmt.Println(">> 检查智能合约是否就绪......")
	if err:=checkCCCommitReadiness(packageID, info.ChaincodeID, info.ChaincodeVersion, sequence, info.ChannelID, info.Orgs); err!=nil{
		return fmt.Errorf("checkCCCommitReadiness error: %v", err)
	}
	fmt.Println(">> 智能合约已经就绪")

	// Commit cc
	fmt.Println(">> 提交智能合约定义......")
	if err:=commitCC(info.ChaincodeID, info.ChaincodeVersion, sequence, info.ChannelID, info.Orgs, info.OrdererEndpoint);err!=nil{
		return fmt.Errorf("commitCC error: %v", err)
	}
	// Query committed cc
	if err:=queryCommittedCC(info.ChaincodeID, info.ChannelID, sequence, info.Orgs); err!=nil{
		return fmt.Errorf("queryCommittedCC error: %v", err)
	}
	fmt.Println(">> 智能合约定义提交完成")

	// Init cc
	fmt.Println(">> 调用智能合约初始化方法......")
	if err:=initCC(info.ChaincodeID, upgrade, info.ChannelID, info.Orgs[0], sdk); err!=nil{
		return fmt.Errorf("initCC error: %v", err)
	}
	fmt.Println(">> 完成智能合约初始化")
	return nil
}
```

# SDK包官方使用文档
链接:https://github.com/hyperledger/fabric-sdk-go/blob/v1.0.0/doc.go,具体内部方法说明参照fabric-sdk官方说明文档，链接如下内容中的Reference内容链接
```
// Package fabricsdk enables Go developers to build solutions that interact with Hyperledger Fabric.
//
// Packages for end developer usage
//
// pkg/fabsdk: The main package of the Fabric SDK. This package enables creation of contexts based on
// configuration. These contexts are used by the client packages listed below.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/fabsdk
//
// pkg/client/channel: Provides channel transaction capabilities.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/channel
//
// pkg/client/event: Provides channel event capabilities.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/event
//
// pkg/client/ledger: Enables queries to a channel's underlying ledger.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/ledger
//
// pkg/client/resmgmt: Provides resource management capabilities such as installing chaincode.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt
//
// pkg/client/msp: Enables identity management capability.
// Reference: https://godoc.org/github.com/hyperledger/fabric-sdk-go/pkg/client/msp
```