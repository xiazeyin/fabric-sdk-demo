package sdkInit

import (
	"fmt"
	"github.com/xiazeyin/fabric-sdk-go-gm/pkg/client/channel"
)

func (t *Application) Set(args []string) (string, error) {
	var tempArgs [][]byte
	for i := 1; i < len(args); i++ {
		tempArgs = append(tempArgs, []byte(args[i]))
	}

	request := channel.Request{ChaincodeID: t.SdkEnvInfo.ChaincodeID, Fcn: args[0], Args: tempArgs}
	response, err := t.SdkEnvInfo.ChClient.Execute(request)
	if err != nil {
		// 资产转移失败
		return "", err
	}

	fmt.Println("============== response:", response)
	fmt.Println("payload:", string(response.Payload))

	return string(response.TransactionID), nil
}
