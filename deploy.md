# 使用测试网络
1. 搭建网络并创建通道
```
./network.sh up createChannel -c mychannel
```
2. 部署链码
```
需要先设置下go的代理 go env -w GOPROXY=https://goproxy.cn,direct
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
```

3. 调用链码
```
设置环境变量
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
# Environment variables for Org1

export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
调用链码
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'
```

# 使用命令搭建网络
1. 通过配置文件生成所有证书
```
./bin/cryptogen generate --config=crypto-config.yaml
```
2. 生成创世块
```
./bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block -channelID fabric-channel
```
3. 创建通道
```
./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/mychannel.tx -channelID mychannel
```
4. 创建组织在通道中的锚定文件
```
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP

./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP
```
5. 启动网络
```
docker-compose up -d
```
6. 关闭网络
```
docker-compose down
docker volume prune
```
7. peer1加入通道
```
进入cli1容器
docker exec -it cli1 bash
peer1创建通道
peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/channel.tx --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
退出容器
exit
将cli1容器的通道文件复制到当前路径
docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/mychannel.block ./
将通道文件复制到cli2容器中
docker cp ./mychannel.block cli2:/opt/gopath/src/github.com/hyperledger/fabric/peer
开启另一个终端进入cli2容器
docker exec -it cli2 bash
peer2节点加入mychannel通道
peer channel join -b mychannel.block
进入cli1容器然后peer1节点加入mychannel通道
docker exec -it cli1 bash
peer channel join -b mychannel.block
在cli1容器中更新组织1的锚节点
peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org1MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
在cli2容器中更新组织2的锚节点
peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org2MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/msp/tlscacerts/tlsca.example.com-cert.pem
```
8. 将链码放到指定位置
sacc.go文件
9. 打包链码
```
进入cli1容器
docker exec -it cli1 bash
到链码所在路径
cd /opt/gopath/src/github.com/hyperledger/fabric-cluster/chaincode/go
给链码依赖包
go env -w GOPROXY=https://goproxy.cn,direct
go mod init
go mod vendor
打包链码
peer lifecycle chaincode package sacc.tar.gz --path github.com/hyperledger/fabric-cluster/chaincode/go/ --label sacc_1
把链码放到工作目录
cp sacc.tar.gz /opt/gopath/src/github.com/hyperledger/fabric/peer
将打包的链码拷贝到cli2中
退出当前容器 exit
拷贝打包的链码到当前路径
docker cp cli1:/opt/gopath/src/github.com/hyperledger/fabric/peer/sacc.tar.gz ./
将打包的链码拷贝到cli2容器中
docker cp sacc.tar.gz cli2:/opt/gopath/src/github.com/hyperledger/fabric/peer
进入cli2容器中查看是否拷贝成功
```
10. 安装链码
```
分别进入cli1和cli2然后
peer lifecycle chaincode install sacc.tar.gz
```
11. 所有组织批准
```
分别在cli1和cli2容器中执行
peer lifecycle chaincode approveformyorg --channelID mychannel --name sacc --version 1.0 --init-required --package-id sacc_1:44213cad421c117cc7ba9311e9cb8bb9f4dc6d9c6307e8d4cc5c48b49d79c55d --sequence 1 --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
查看是否执行成功
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name sacc --version 1.0 --init-required --sequence 1 --tls true --cafile  /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --output json
```
12. 提交commit
```
随便在一个组织下提交就可以
peer lifecycle chaincode commit -o orderer.example.com:7050 --channelID mychannel --name sacc --version 1.0 --init-required --sequence 1 --tls true --cafile  /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
```
13. 调用invoke
```
peer chaincode invoke -o orderer.example.com:7050 --isInit --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n sacc --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c  '{"Args":["a","bb"]}'
```
14. 查询
```
peer chaincode query -C mychannel -n sacc -c '{"Args":["query","a"]}'
```
15. 第二次调用
```
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls true --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n sacc --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt -c  '{"Args":["set","a","cc"]}'
```
# 向组织中添加peer节点
1. 修改crypto-config.yaml文件
```
- Name: Org2
    Domain: org2.example.com
    EnableNodeOUs: true
    Template:
      Count: 2
    Users:
      Count: 1
```
Count值改为2
2. 生成新的证书
```
./bin/cryptogen extend --config=crypto-config.yaml
```
3. 配置docker-compose文件并启动节点容器
新增配置文件Org2Peer1.yaml,内容为
```
version: '2.1'

volumes:
  peer1.org2.example.com:

networks:
  test:
    name: twonodes_test

services:
  peer1.org2.example.com:
    container_name: peer1.org2.example.com
    image: qychain/fabric-peer:2.2.5-0.0.1
    environment:
      #Generic peer variables
      - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
      - CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=twonodes_test
      - FABRIC_LOGGING_SPEC=INFO
      #- FABRIC_LOGGING_SPEC=DEBUG
      - CORE_PEER_TLS_ENABLED=true
      - CORE_PEER_PROFILE_ENABLED=true
      - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt
      - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key
      - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt
      # Peer specific variabes
      - CORE_PEER_ID=peer1.org2.example.com
      - CORE_PEER_ADDRESS=peer1.org2.example.com:10051
      - CORE_PEER_LISTENADDRESS=0.0.0.0:10051
      - CORE_PEER_CHAINCODEADDRESS=peer1.org2.example.com:10052
      - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:10052
      - CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer0.org2.example.com:9051
      - CORE_PEER_GOSSIP_BOOTSTRAP=peer1.org2.example.com:10051
      - CORE_PEER_LOCALMSPID=Org2MSP
    volumes:
      - /var/run/docker.sock:/host/var/run/docker.sock
      - ./crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp:/etc/hyperledger/fabric/msp
      - ./crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls:/etc/hyperledger/fabric/tls
      - peer1.org2.example.com:/var/hyperledger/production
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric/peer
    command: peer node start
    ports:
      - 10051:10051
    networks:
      - test
```
4. 启动新增的节点
```
docker-compose -f Org2Peer1.yaml up -d
```
5. 将新节点加入所需的channel中

6. 