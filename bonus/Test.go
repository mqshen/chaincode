package main

//import (
//	"fmt"
//	"encoding/json"
//)
//
//
//func main() {
//
//	jsonString := "[{\"expire\":20171101,\"amount\":80},{\"expire\":20171201,\"amount\":160}]"
//	targetUserAssetString := "[{\"expire\":20171101,\"amount\":120},{\"expire\":20171201,\"amount\":20}]"
//	transferDetail := "[{\"expire\":20171101,\"amount\":20,\"ext\":\"1\"},{\"expire\":20171201,\"amount\":20,\"ext\":\"\"}]"
//
//	var details []UserAsset
//	err := json.Unmarshal([]byte(transferDetail), &details)
//	if err != nil {
//		fmt.Printf("Failed decod transfer detail")
//		return
//	}
//	var userAssets []UserAsset
//	err = json.Unmarshal([]byte(jsonString), &userAssets)
//	if err != nil {
//		fmt.Printf("Failed decod inf owner")
//		return
//	}
//
//	var targetUserAssetResult []byte
//	targetUserAssetResult = []byte(targetUserAssetString)
//	var targetUserAssets []UserAsset
//	if err == nil && targetUserAssetResult !=  nil {
//		err = json.Unmarshal(targetUserAssetResult, &targetUserAssets)
//		if err != nil {
//			fmt.Printf("Failed decod target's asset detail")
//			return
//		}
//	}
//
//	remainArray := userAssets
//	var transferArray []UserAsset
//	for _, detail := range details {
//		remainArray, transferArray, err = calculateTransferArray(remainArray, detail.Expire, detail.Amount)
//		if err != nil {
//			fmt.Printf("calculate transfer array failed: " + err.Error())
//			return
//		}
//		if targetUserAssets ==  nil {
//			targetUserAssets = transferArray
//		} else {
//			test, _ := json.Marshal(transferArray)
//			fmt.Printf("target user json:%s \n", test)
//			fmt.Println("222")
//			targetUserAssets, err = calculateInsert(targetUserAssets, transferArray)
//		}
//	}
//	targetUserAssetJson, err := json.Marshal(targetUserAssets)
//	fmt.Printf("target user json:%s \n", targetUserAssetJson)
//
//	userAssetResult, err := json.Marshal(remainArray)
//	if err != nil {
//		fmt.Printf("marshal user's asset failed")
//	}
//	fmt.Printf("source user json:%s \n", userAssetResult)
//}