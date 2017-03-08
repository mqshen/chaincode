package main

import (
	"fmt"
)

func main() {

	jsonString := "[{\"expire\":20171010,\"balance\":50}]"
	targetUserAssetString := "[{\"expire\":20171001,\"balance\":50}]"
	userAssetResult, transferArray, err := calculateTransferArray([]byte(jsonString), 20170302, 20)
	if err != nil {
		fmt.Printf("user asset result: %s\n", err)
	}
	for _, asset := range transferArray {
		fmt.Printf("transfer array asset: %d, %d\n", asset.Expire, asset.Balance)
	}
	targetUserAssetResult, err := calculateInsert([]byte(targetUserAssetString), transferArray)

	fmt.Printf("user asset result:%s\n", string(userAssetResult))
	if err != nil {
		fmt.Printf("target user asset result: %s\n", err)
	}
	fmt.Printf("target user asset result:%s\n", string(targetUserAssetResult))
}
