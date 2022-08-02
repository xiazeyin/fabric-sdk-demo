package main
 
import (
        // "bccsp"
        //"crypto"
        //"crypto/rsa"
        "encoding/json"
        "errors"
        "fmt"
        "github.com/xiazeyin/fabric-chaincode-go-gm/shim"
        "github.com/xiazeyin/fabric-protos-go-gm/peer"
        //"github.com/hyperledger/fabric-chaincode-go/shim"
        //"github.com/hyperledger/fabric-protos-go/peer"
        //"gitee.com/zhaochuninhefei/fabric-chaincode-go-gm/shim"
        //"gitee.com/zhaochuninhefei/fabric-protos-go-gm/peer"
        //"math/big"
        "sort"
        "strconv"
        //"strings"
        "time"
)
 
// NFT资产
type address string
type tokenId string
type url string
type transTime int
 
type Asset struct {
        NftName          string                            `json:"nftName"`               //资产token名称
        NftSymbol        string                            `json:"nftSymbol"`             //资产token简称
        Owner            string                            `json:"owner"`                 //资产token创建人
        AssetOf          map[address]map[tokenId]url       `json:"assetOf"`               //资产token的持有人和持���资产的集合
        AssetHistory     map[tokenId]map[transTime]address `json:"assetHistory"`          //资产token转移记录
        AssetApprove     map[tokenId]address               `json:"assetApprove"`          //资产授权人
        AssetExit        map[url]bool                      `json:"assetExit"`             //资产token是否已经存在
        AssetTokenIdExit map[tokenId]bool                  `json:"assetTokenIdExit"`      //资产token是否已经存在
        TotalSupply      int                               `json:"totalSupply,omitempty"` //资产总量 0代表不限量NFT 非0  则代表限量版本的NFT
}
 
// 交易结构
type TransferAsset struct {
        _From      address `json:"_from,omitempty"`      // 转出人 / 资产owner
        _To        address `json:"_to"`                  // 确权/转移到的地址
        _TokenId   tokenId `json:"_tokenId"`             // 每一个存证确权为一个资产token 每个资产token有唯一的tokenId
        _Url       url     `json:"_url"`                 // 存证的txid  或者 资产的url
        _Data      string  `json:"_data,omitempty"`      // 备注
        _Signature string  `json:"_signature,omitempty"` // 签名
        _SignType  string  `json:"_signType,omitempty"`  // 签名类型
        _NftSymbol string  `json:"_nftSymbol"`           //资产token简称
}
 
func (u *Asset) Init(stub shim.ChaincodeStubInterface) peer.Response {
        return shim.Success(nil)
}
 
// 初始化NFT资产合约 名称 简称 创造人
func (s *Asset) initLedger(stub shim.ChaincodeStubInterface, args []string) peer.Response {
 
        if len(args) != 4 {
                return shim.Error("Incorrect number of arguments. Expecting 4")
        }
 
        name := args[0]                         // nft资产名称
        symbol := args[1]                       // nft资产简称
        owner := args[2]                        // nft资产创始人  pubkey
        totalSupply, _ := strconv.Atoi(args[3]) // nft资产是否为限量资产
 
        // 检查当前token是否已经被创建
        value, err := stub.GetState(symbol)
        if err != nil {
                return shim.Error(err.Error())
        }
        if value != nil || len(value) > 0 {
                return shim.Error("symbol 已存在")
        }
 
        assetHistory := make(map[tokenId]map[transTime]address)
        assetApprove := make(map[tokenId]address)
        assetOf := make(map[address]map[tokenId]url)
        assetExit := make(map[url]bool)
        assetTokenIdExit := make(map[tokenId]bool)
 
        // 创建新的token
        asset := &Asset{
                NftName:          name,
                NftSymbol:        symbol,
                Owner:            owner,
                AssetOf:          assetOf,
                AssetHistory:     assetHistory,
                AssetApprove:     assetApprove,
                AssetExit:        assetExit,
                AssetTokenIdExit: assetTokenIdExit,
                TotalSupply:      totalSupply,
        }
 
        //存储token
        tokenAsBytes, _ := json.Marshal(asset)
        err = stub.PutState(symbol, tokenAsBytes)
        if err != nil {
                return shim.Error(err.Error())
        }
        fmt.Printf("Init %s \n", string(tokenAsBytes))
        return shim.Success(nil)
}
 
// invoke 执行
func (u *Asset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
        funcName, args := stub.GetFunctionAndParameters()
        switch funcName {
        case "mine":
                return u.mine(stub, args)
        case "initLedger":
                return u.initLedger(stub, args)
        case "transFrom":
                return u.transFrom(stub, args)
        case "approve":
                return u.approve(stub, args)
        case "withDrawal":
                return u.withDrawal(stub, args)
        case "balanceof":
                return u.balanceof(stub, args)
        case "getApproved":
                return u.getApproved(stub, args)
        case "ownerof":
                return u.ownerof(stub, args)
        case "tokenURL":
                return u.tokenURL(stub, args)
        case "name":
                return u.name(stub, args)
        case "owner":
                return u.owner(stub, args)
        case "symbol":
                return u.symbol(stub, args)
        case "totalSupply":
                return u.totalSupply(stub, args)
        case "tokenBasic":
                return u.tokenBasic(stub, args)
		case "save":
				return u.save(stub, args)
        default:
                return shim.Error("no such function")
        }
}
 
// 确权资产
func (u *Asset) mine(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 6 {
                return shim.Error("expect six args")
        }
 
        // 生成资产的唯一id  uuid 由外部传入 因为fabric机制原因 内务无法统一uuid
        transferAsset := TransferAsset{
                _NftSymbol: args[0],
                _From:      address(args[1]),
                _To:        address(args[1]),
                _TokenId:   tokenId(args[2]),
                _Url:       url(args[3]),
                _Signature: args[4],
                _SignType:  args[5],
        }
        fmt.Printf("mine tokenname:%s \n _to:%s \n _tokenId:%s \n url:%s...\n", transferAsset._NftSymbol, transferAsset._To, transferAsset._TokenId, transferAsset._Url)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(transferAsset._NftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //只有存在，也就是初始化过了的nft token 进行下一步   验签
        /*
        veriftResul, errVerify := u.onlyOwner(transferAsset)
        if errVerify != nil || veriftResul == false {
                return shim.Error("Signature verification failed")
        }*/
 
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        if nftToken.AssetTokenIdExit[transferAsset._TokenId] {
                return shim.Error("tokenId already exit")
        }
 
        // 资产的添加
        errMine := nftToken.mineOnly(stub, &transferAsset)
        if errMine != nil {
                return shim.Error(errMine.Error())
        }
 
        // 存储
        assetBytes, err := json.Marshal(&nftToken)
        if err != nil {
                return shim.Error(err.Error())
        }
        err = stub.PutState(args[0], assetBytes)
        if err != nil {
                return shim.Error(err.Error())
        }
        fmt.Println("putState success...")
        return shim.Success(nil)
}
 
//确权时更改资产持有属性
func (u *Asset) mineOnly(stub shim.ChaincodeStubInterface, trans *TransferAsset) error {
 
        //判断是否已经存在该url
        _, ok := u.AssetExit[trans._Url]
        if !ok {
                //添加资产持有记录
                if u.AssetOf[trans._To] == nil {
                        mc := make(map[tokenId]url)
                        mc[trans._TokenId] = trans._Url
                        u.AssetOf[trans._To] = mc
                } else {
                        u.AssetOf[trans._To][trans._TokenId] = trans._Url
                }
 
                //添加资产已经存在的标志
                u.AssetExit[trans._Url] = true
 
                if u.AssetHistory[trans._TokenId] == nil {
                        //添加资产转移记录
                        mh := make(map[transTime]address)
                        t1, _ := stub.GetTxTimestamp()
                        time1 := time.Unix(t1.Seconds, int64(t1.Nanos))
                        timeMi:=time1.UnixNano()/1e6
 
                        mh[transTime(timeMi)] = trans._To
                        u.AssetHistory[trans._TokenId] = mh
                } else {
                        t1, _ := stub.GetTxTimestamp()
                        time1 := time.Unix(t1.Seconds, int64(t1.Nanos))
                        timeMi:=time1.UnixNano()/1e6
                        u.AssetHistory[trans._TokenId][transTime(timeMi)] = trans._To
                }
 
                u.AssetTokenIdExit[trans._TokenId] = true
 
                //判断资产总量是否到达限制
				if u.TotalSupply != 0 {
					if len(u.AssetExit) > u.TotalSupply {
                        fmt.Println("当前nft资产已达到限制")
                        return errors.New("TotalSupply already run out")
					}
				}
                
                return nil
        } else {
                return errors.New("url already exist")
        }
 
}
 
//交易时更改资产持有属性
func (u *Asset) transOnly(stub shim.ChaincodeStubInterface, trans *TransferAsset) error {
 
        //修改资产持有记录 1 删除当前from的资产 2 增加to的资产
        delete(u.AssetOf[(trans._From)], trans._TokenId)
 
        if u.AssetOf[(trans._To)] == nil {
                mc := make(map[tokenId]url)
                mc[trans._TokenId] = trans._Url
                u.AssetOf[(trans._To)] = mc
        } else {
                u.AssetOf[(trans._To)][trans._TokenId] = trans._Url
        }
 
        //添加资产转移记录
        t1, _ := stub.GetTxTimestamp()
        time1 := time.Unix(t1.Seconds, int64(t1.Nanos))
        timeMi:=time1.UnixNano()/1e6
        u.AssetHistory[trans._TokenId][transTime(timeMi)] = trans._To
        return nil
 
}
 
//资产转移
func (u *Asset) transFrom(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 6 {
                return shim.Error("expect six args")
        }
 
        transferAsset := TransferAsset{
                _NftSymbol: args[0],
                _From:      address(args[1]),
                _To:        address(args[2]),
                _TokenId:   tokenId(args[3]),
                _Signature: args[4],
                _SignType:  args[5],
        }
        fmt.Printf("transFrom tokenname:%s \n _to:%s \n _tokenId:%s \n url:%s...\n", transferAsset._NftSymbol, transferAsset._To, transferAsset._TokenId, transferAsset._Url)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(transferAsset._NftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol not exist")
        }
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        if !nftToken.AssetTokenIdExit[transferAsset._TokenId] {
                return shim.Error("_tokenId not exist")
        }
 
        // 判断是否有权限进行交易 是否是token持有人 是否是token授权人
        beProved :=false
        if nftToken.AssetOf[transferAsset._From] != nil && nftToken.AssetOf[transferAsset._From][transferAsset._TokenId] == "" {
                if(nftToken.AssetApprove[transferAsset._TokenId]==transferAsset._From){
                        fmt.Println("tokenId will be approved trans")
                        beProved = true
                }else {
                        return shim.Error("your have No permission")
                }
        }
 
        // 判断授权后的交易是否交易给自己
        var keys []int
        for k := range nftToken.AssetHistory[tokenId(transferAsset._TokenId)] {
                keys = append(keys, int(k))
        }
        sort.Ints(keys)
        _keyInt := keys[len(keys)-1]
        _Owner := nftToken.AssetHistory[tokenId(transferAsset._TokenId)][transTime(_keyInt)]
        if transferAsset._To ==_Owner {
                return shim.Error("You are not allowed to sell it to yourself")
        }
 
        transferAsset._Url = nftToken.AssetOf[_Owner][transferAsset._TokenId]
        //只有存在���也就是初始化过了的nft token 进行下一步   验签
        /*
        veriftResul, errVerify := nftToken.onlyOwner(transferAsset)
        if errVerify != nil || veriftResul == false {
                return shim.Error("Signature verification failed")
        }*/
 
        // 资产的转移
        errMine := nftToken.transOnly(stub, &transferAsset)
        if errMine != nil {
                return shim.Error(errMine.Error())
        }
        // 如果是授权账户发起的 则取消授权
        if beProved{
                delete(nftToken.AssetApprove,transferAsset._TokenId)
        }
 
        // 存储
        assetBytes, err := json.Marshal(&nftToken)
        if err != nil {
                return shim.Error(err.Error())
        }
        err = stub.PutState(args[0], assetBytes)
        if err != nil {
                return shim.Error(err.Error())
        }
        fmt.Println("putState success...")
        return shim.Success(nil)
}
 
//资产授权--授权后，资产权限归平台所有，资产所有人无法转移
func (u *Asset) approve(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 5 {
                return shim.Error("expect five args")
        }
 
        _nftSymbol := args[0]
        _approved := args[1]
        _tokenId := args[2]
        //_signature := args[3]
        //_signType := args[4]
 
        // todo  缺少_approved地址的合法性校验
 
        fmt.Printf("approve _nftSymbol:%s \n _approved:%s \n  _tokenId:%s \n", _nftSymbol, _approved, _tokenId)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        // nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        var keys []int
        for k := range nftToken.AssetHistory[tokenId(_tokenId)] {
                keys = append(keys, int(k))
        }
        sort.Ints(keys)
        _keyInt := keys[len(keys)-1]
        _Owner := nftToken.AssetHistory[tokenId(_tokenId)][transTime(_keyInt)]
        if _Owner == address(_approved) {
                return shim.Error("You can't delegate to yourself")
        }
 
        //_url := nftToken.AssetOf[_Owner][tokenId(_tokenId)]
 /*
        transferAsset := TransferAsset{
                _Url:       _url,
                _From:      _Owner,
                _Signature: _signature,
                _SignType:  _signType,
        }
 */
        //  验签
        /*
        veriftResul, errVerify := nftToken.onlyOwner(transferAsset)
        if errVerify != nil || veriftResul == false {
                return shim.Error("Signature verification failed")
        }*/
 
        // 资产的授权
        if len(nftToken.AssetApprove[tokenId(_tokenId)]) != 0 {
                return shim.Error("Authorization cannot be repeated")
        } else {
                nftToken.AssetApprove[tokenId(_tokenId)] = address(_approved)
        }
 
        // 存储
        assetBytes, err := json.Marshal(&nftToken)
        if err != nil {
                return shim.Error(err.Error())
        }
        err = stub.PutState(args[0], assetBytes)
        if err != nil {
                return shim.Error(err.Error())
        }
        fmt.Println("putState success...")
        return shim.Success(nil)
}
 
//资产提现--授权的资产 取消授权  即为提现，提现后的资产可以交易转移，授权后未提现的，无法转移
func (u *Asset) withDrawal(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 4 {
                return shim.Error("expect four args")
        }
 
        _nftSymbol := args[0]
        _tokenId := args[1]
        //_signature := args[2]
        //_signType := args[3]
 
        fmt.Printf("withDrawal _nftSymbol:%s \n  _tokenId:%s \n", _nftSymbol, _tokenId)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol not exist ")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        //判断是否已经授权过
        if len(nftToken.AssetApprove[tokenId(_tokenId)]) == 0 {
                return shim.Error("have not approve,no promission do")
        }
 
        var keys []int
        for k := range nftToken.AssetHistory[tokenId(_tokenId)] {
                keys = append(keys, int(k))
        }
        sort.Ints(keys)
        //_keyInt := keys[len(keys)-1]
        //_Owner := nftToken.AssetHistory[tokenId(_tokenId)][transTime(_keyInt)]
 
        //_url := nftToken.AssetOf[_Owner][tokenId(_tokenId)]
 /*
        transferAsset := TransferAsset{
                _Url:       _url,
                _From:      _Owner,
                _Signature: _signature,
                _SignType:  _signType,
        }
*/ 
        //  验签
        /*
        veriftResul, errVerify := nftToken.onlyOwner(transferAsset)
        if errVerify != nil || veriftResul == false {
                return shim.Error("Signature verification failed")
        }
        */
 
        // 提现
        delete(nftToken.AssetApprove, tokenId(_tokenId))
 
        // 存储
        assetBytes, err := json.Marshal(&nftToken)
        if err != nil {
                return shim.Error(err.Error())
        }
        err = stub.PutState(args[0], assetBytes)
        if err != nil {
                return shim.Error(err.Error())
        }
        fmt.Println("putState success...")
        return shim.Success(nil)
}
 
//查询资产余额
func (u *Asset) balanceof(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 2 {
                return shim.Error("expect two args")
        }
 
        _nftSymbol := args[0]
        _owner := args[1]
        fmt.Printf("balanceof _nftSymbol:%s \n  _owner:%s \n", _nftSymbol, _owner)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        _ownerMap := nftToken.AssetOf[address(_owner)]
 
        _ownerMapAsBytes, _ := json.Marshal(_ownerMap)
 
        fmt.Println("putState success...")
        return shim.Success(_ownerMapAsBytes)
}
 
//查询资产是否已经被授权  返回 true /false
func (u *Asset) getApproved(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 2 {
                return shim.Error("expect two args")
        }
 
        _nftSymbol := args[0]
        _tokenId := args[1]
        fmt.Printf("getApproved _nftSymbol:%s \n  _tokenId:%s \n", _nftSymbol, _tokenId)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        if len(nftToken.AssetApprove[tokenId(_tokenId)]) == 0 {
                res, _ := json.Marshal(false)
                return shim.Success(res)
        }
        res, _ := json.Marshal(true)
        return shim.Success(res)
}
 
//资产tokenId 获取实际持有人
func (u *Asset) ownerof(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 2 {
                return shim.Error("expect two args")
        }
 
        _nftSymbol := args[0]
        _tokenId := args[1]
 
        fmt.Printf("ownerof _nftSymbol:%s \n  _tokenId:%s \n", _nftSymbol, _tokenId)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        // 判断tokenId存不存在
        if len(nftToken.AssetHistory[tokenId(_tokenId)]) == 0 {
                return shim.Error("tokenId 不存在")
        }
 
        //查找实际控制人
        var keys []int
        for k := range nftToken.AssetHistory[tokenId(_tokenId)] {
                keys = append(keys, int(k))
        }
        sort.Ints(keys)
        _keyInt := keys[len(keys)-1]
        _Owner := nftToken.AssetHistory[tokenId(_tokenId)][transTime(_keyInt)]
 
        ownerBytes, _ := json.Marshal(_Owner)
 
        return shim.Success(ownerBytes)
}
 
//资产tokenId 获取token的url指向 源文件所在位置
func (u *Asset) tokenURL(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 2 {
                return shim.Error("expect two args")
        }
 
        _nftSymbol := args[0]
        _tokenId := args[1]
 
        fmt.Printf("ownerof _nftSymbol:%s \n  _tokenId:%s \n", _nftSymbol, _tokenId)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        // 判断tokenId存不存在
        if len(nftToken.AssetHistory[tokenId(_tokenId)]) == 0 {
                return shim.Error("tokenId 不存在")
        }
 
        //查找实际控制人
        var keys []int
        for k := range nftToken.AssetHistory[tokenId(_tokenId)] {
                keys = append(keys, int(k))
        }
        sort.Ints(keys)
        _keyInt := keys[len(keys)-1]
        _Owner := nftToken.AssetHistory[tokenId(_tokenId)][transTime(_keyInt)]
 
        // 获取url
        urlFinal := nftToken.AssetOf[_Owner][tokenId(_tokenId)]
 
        urlBytes, _ := json.Marshal(urlFinal)
        return shim.Success(urlBytes)
}
 
//资产 获取token name
func (u *Asset) name(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 1 {
                return shim.Error("expect one args")
        }
 
        _nftSymbol := args[0]
 
        fmt.Printf("name _nftSymbol:%s \n", _nftSymbol)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        nameBytes, _ := json.Marshal(nftToken.NftName)
        return shim.Success(nameBytes)
}
 
//资产 获取token owner
func (u *Asset) owner(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 1 {
                return shim.Error("expect one args")
        }
 
        _nftSymbol := args[0]
 
        fmt.Printf("owner _nftSymbol:%s \n", _nftSymbol)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        ownerBytes, _ := json.Marshal(nftToken.Owner)
        return shim.Success(ownerBytes)
}
 
//资产 获取token symbol
func (u *Asset) symbol(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 1 {
                return shim.Error("expect one args")
        }
 
        _nftSymbol := args[0]
 
        fmt.Printf("symbol _nftSymbol:%s \n", _nftSymbol)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        nameBytes, _ := json.Marshal(nftToken.NftSymbol)
        return shim.Success(nameBytes)
}
 
//资产 获取token totalSupply
func (u *Asset) totalSupply(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 1 {
                return shim.Error("expect one args")
        }
 
        _nftSymbol := args[0]
 
        fmt.Printf("totalSupply _nftSymbol:%s \n", _nftSymbol)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
 
        //nfttoken
        nftToken := Asset{}
        json.Unmarshal(assetValue, &nftToken)
 
        TotalSupplyBytes, _ := json.Marshal(nftToken.TotalSupply)
        return shim.Success(TotalSupplyBytes)
}
 
//资产 获取token 所有基本信息
func (u *Asset) tokenBasic(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 1 {
                return shim.Error("expect one args")
        }
 
        _nftSymbol := args[0]
 
        fmt.Printf("totalSupply _nftSymbol:%s \n", _nftSymbol)
 
        //验证  当前的 nft token symbol简称存不存在
        assetValue, errAsset := stub.GetState(_nftSymbol)
        if errAsset != nil {
                return shim.Error(errAsset.Error())
        }
        if assetValue == nil || len(assetValue) == 0 {
                return shim.Error("_symbol 不存在")
        }
        return shim.Success(assetValue)
}

// 哈希存证
func (u *Asset) save(stub shim.ChaincodeStubInterface, args []string) peer.Response {
        if len(args) != 1 {
                return shim.Error("expect one args")
        }
 
        err := stub.PutState(args[0], []byte("1"))
		if err != nil {
			return shim.Error(err.Error())
		}
 
        return shim.Success(nil)
}
 
// 获取公钥
/*func getRSAPubkey(codeStr string) (*rsa.PublicKey, error) {
        n := &big.Int{}
        _, flag := n.SetString(codeStr, 16)
        if flag != true {
                return nil, errors.New(fmt.Sprintf("publicKey.n=%s err", codeStr))
        }
        // 构造返回PublicKey引用
        publicKeyObj := &rsa.PublicKey{}
        publicKeyObj.E = 65537
        publicKeyObj.N = n
        return publicKeyObj, nil
}*/
 
// 获取公钥
/*
func getRSAPubkey(codeStr string) (pubKey bccsp.Key, err error) {
        pubArr := strings.Split(codeStr, "|")
        if pubArr == nil || len(pubArr) != 2 {
                return nil, errors.New("pubkey array length invalid")
        }
        e_str := pubArr[0]
        n_str := pubArr[1]
 
        //转换e
        e, err := strconv.Atoi(e_str)
        if err != nil {
                return nil, errors.New("strconv e error")
        }
        //转换n
        n := &big.Int{}
        _, flag := n.SetString(n_str, 16)
        if flag != true {
                return nil, errors.New("strconv n error")
        }
        //构造key
        publickObj := &rsa.PublicKey{}
        publickObj.E = e
        publickObj.N = n
        return &bccsp.RSAPublicKey{publickObj}, nil
}
*/
 
//验证owner所有人 验签
/*
func (u *Asset) onlyOwner(transferAsset TransferAsset) (bool, error) {
        veriftAccount := ""
 
        //是否授权 授权优先验证授权账户 未授权 验证from存不存在，不存在，代表是mine确权资产，使用to验签，存在，代表是资产转移
        if len(u.AssetApprove[transferAsset._TokenId]) != 0 {
                veriftAccount = string(u.AssetApprove[transferAsset._TokenId])
        } else {
                veriftAccount = string(transferAsset._From)
        }
 
        var bccspInstance, _ = bccsp.NewBccsp()
        if transferAsset._SignType == "RSA" {
                //验签
                signBytes := HexToByte(transferAsset._Signature)
 
                // 验签的转换
                var hash bccsp.HashOpts
                var pubKey bccsp.Key
                rsaPubkey, err := getRSAPubkey(veriftAccount)
                if err != nil {
                        return false, err
                }
                pubKey = rsaPubkey
 
                hash = &bccsp.SHA1Opts{}
                bytesUrl, _ := json.Marshal(transferAsset._Url)
                digest, err := bccspInstance.Hash(bytesUrl, hash)
 
                isVerify, err := bccspInstance.Verify(pubKey, signBytes, digest, &bccsp.RsaPKCS1V15Opts{Hash: crypto.SHA1})
                fmt.Println("isVerify:%s", isVerify)
                if err != nil {
                        return false, err
                }
                fmt.Println("check sign success...")
                return true, nil
        } else {
                return true, nil
        }
}
*/
 
//16进制字符串转[]byte
func HexToByte(hex string) []byte {
        length := len(hex) / 2
        slice := make([]byte, length)
        rs := []rune(hex)
 
        for i := 0; i < length; i++ {
                s := string(rs[i*2 : i*2+2])
                value, _ := strconv.ParseInt(s, 16, 10)
                slice[i] = byte(value & 0xFF)
        }
        return slice
}
 
func main() {
        err := shim.Start(new(Asset))
        if err != nil {
                fmt.Println("poe chaincode start error")
        }
        return
}