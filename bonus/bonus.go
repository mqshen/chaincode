/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"encoding/json"
	"bytes"
	"github.com/golang/protobuf/proto"
)

type assetIssue struct {
	Owner      string `json:"owner"`
	Balance	   int    `json:"balance"`
	Name       string `json:"name"`
}

type userAsset struct {
	Expire     int    `json:"expire"`
	Balance	   int    `json:"balance"`
}

// This struct represents an Identity
// (with its MSP identifier) to be used
// to serialize it and deserialize it
type SerializedIdentity struct {
	// The identifier of the associated membership service provider
	Mspid string `protobuf:"bytes,1,opt,name=Mspid" json:"Mspid,omitempty"`
	// the Identity, serialized according to the rules of its MPS
	IdBytes []byte `protobuf:"bytes,2,opt,name=IdBytes,proto3" json:"IdBytes,omitempty"`
}

func (m *SerializedIdentity) Reset()                    { *m = SerializedIdentity{} }
func (m *SerializedIdentity) String() string            { return proto.CompactTextString(m) }
func (*SerializedIdentity) ProtoMessage()               {}

// BonusManagementChaincode is simple chaincode implementing a basic Asset Management system
// with access control enforcement at chaincode level.
// Look here for more information on how to implement access control at chaincode level:
// https://github.com/hyperledger/fabric/blob/master/docs/tech/application-ACL.md
// An asset is simply represented by a string.
type BonusManagementChaincode struct {
}

// Init method will be called during deployment.
// The deploy transaction metadata is supposed to contain the administrator cert
//func (t *BonusManagementChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
func (t *BonusManagementChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Chaincode...")

	// Set the role of the users that are allowed to assign assets
	// The metadata will contain the role of the users that are allowed to assign assets
	adminCert, err := stub.GetCreator()
	if err != nil {
		return shim.Error("Failed getting createor" + err.Error())
	}

	fmt.Printf("admin's cert is %v\n", string(adminCert))

	if len(adminCert) == 0 {
		return shim.Error("Invalid assigner role. Empty.")
	}

	stub.PutState("admin", adminCert)

	return shim.Success(nil)
}

func (t *BonusManagementChaincode) issue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Issue...")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	admin, err := stub.GetState("admin")
	if err != nil {
		return shim.Error("Failed to get admin's cert, error: " + err.Error())
	}

	creator, err := stub.GetCreator()
	if err != nil {
		fmt.Println(err.Error())
		return shim.Error("Failed to get creator's cert, error: " + err.Error())
	}
	if bytes.Compare(admin, creator) != 0 {
		fmt.Printf("admin cert:%x", admin)
		fmt.Printf("creator cert:%x", creator)
		return shim.Error("Failed, the cert of creator and caller is not same")
	}
	assetName := args[0]
	organizationCert := args[1]
	balance, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	oldAsset, err := stub.GetState(assetName)
	if err != nil {
		return shim.Error("asset name: " + assetName + "state get error")
	}
	if oldAsset != nil {
		return shim.Error("asset already issue, please try other name")
	}

	asset := &assetIssue{organizationCert, balance, assetName}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("Issue... asset name: %s, detail: %s!", assetName, string(assetJSONasBytes))
	err = stub.PutState(assetName, assetJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("Issue...done!")

	return shim.Success(nil)
}

func (t *BonusManagementChaincode) assign(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Assign... arg length: %d", len(args))

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	assetName := args[0]
	assetJSONasBytes, err := stub.GetState(assetName)
	if err != nil {
		return shim.Error("asset state get failed, have not issued")
	}
	if assetJSONasBytes == nil {
		return shim.Error("asset have not issued")
	}

	var assetJSON assetIssue
	err = json.Unmarshal(assetJSONasBytes, &assetJSON)
	fmt.Println("asset json string:", assetJSONasBytes)
	if err != nil {
		return shim.Error("Error Failed to decode JSON of: " + assetName + " resason " + err.Error() + " json:" + string(assetJSONasBytes))
	}
	organizationCert := []byte(assetJSON.Owner)
	serializedID, err := stub.GetCreator()
	if err != nil {
		return shim.Error("can not get creator, error:" + err.Error())
	}

	sid := &SerializedIdentity{}
	err = proto.Unmarshal(serializedID, sid)
	creator := sid.IdBytes

	fmt.Println("admin cert length, %d", len(organizationCert))
	fmt.Println("creator cert length, %d", len(creator))

	fmt.Println("admin cert:%x", organizationCert)
	fmt.Println("creator cert:x", creator)

	if bytes.Compare(organizationCert, creator) != 0 {
		return shim.Error("the caller is not the asset's owner")
	}
	user := args[1]
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("amount argument is incorrect")
	}
	if amount < 0  {
		return shim.Error("the amount must not negative")
	}
	if  amount > assetJSON.Balance {
		return shim.Error("the issue balance is small than assign amount")
	}
	expire, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("expire argument is incorrect")
	}

	assetJSON.Balance -= amount
	key := assetName + user
	var userAssets []userAsset
	userAssetString, err := stub.GetState(key)
	if err != nil || userAssetString == nil {
		userAssets = []userAsset{userAsset{expire, amount}}
	} else {
		err = json.Unmarshal(userAssetString, &userAssets)
		if err != nil {
			return shim.Error("unmarshal user balance failed " + err.Error())
		}
		var find = false
		var index = 0
		for i, userAsset := range userAssets {
			if userAsset.Expire == expire {
				userAsset.Balance += amount
				find = true
				break
			}
			if userAsset.Expire < expire {
				index = i
			}
		}
		if !find {
			asset := userAsset{expire, amount}
			userAssets = append(userAssets[:index], append([]userAsset{asset}, userAssets[index:]...)...)
		}
	}
	userAssetResult, err := json.Marshal(userAssets)
	if err != nil {
		return shim.Error("marshal user's asset failed")
	}

	fmt.Println("set key: %s , detail: %s", key, string(userAssetResult))
	err = stub.PutState(key, userAssetResult)
	if err != nil {
		return shim.Error("store user's asset failed")
	}

	return shim.Success(nil)
}

func calculateTransferArray(userAssetString []byte, expire, amount int) ([]byte, []userAsset, error) {
	var userAssets []userAsset
	err := json.Unmarshal(userAssetString, &userAssets)
	if err != nil {
		return nil, nil, errors.New("Failed decod inf owner")
	}
	var startIndex = 0
	var index = 0
	var remainAmount = amount
	var lastAmount = 0
	for i, sourceAsset := range userAssets {
		fmt.Printf("expire: %d, balance: %d, target expire: %d\n", sourceAsset.Expire, sourceAsset.Balance, expire)
		if sourceAsset.Expire < expire {
			startIndex = i + 1
			index = startIndex
		} else if sourceAsset.Expire >= expire {
			if sourceAsset.Balance <= remainAmount {
				remainAmount -= sourceAsset.Balance
				index = i + 1
			} else {
				lastAmount = remainAmount
				remainAmount = 0
				break
			}
		}
	}
	if remainAmount > 0 {
		return nil, nil, errors.New("balance is less then transfer amount")
	}
	var startArray = make([]userAsset, 0)
	if startIndex > 0 {
		startArray = userAssets[:startIndex]
	}
	transferArray := userAssets[startIndex:index]
	fmt.Printf("start index: %d\n", startIndex)
	remainArray := userAssets[index:]
	fmt.Printf("last amount: %d\n", lastAmount)
	dest := make([]userAsset, len(transferArray));
	copy(dest, transferArray);
	if lastAmount > 0 {
		remainArray[0].Balance -= lastAmount
		asset := userAsset{remainArray[0].Expire, lastAmount}
		dest = append(dest, asset)
	}
	for _, asset := range dest {
		fmt.Printf("11transfer array asset: %d, %d\n", asset.Expire, asset.Balance)
	}
	remainArray = append(startArray, remainArray...)
	for _, asset := range dest {
		fmt.Printf("22transfer array asset: %d, %d\n", asset.Expire, asset.Balance)
	}
	userAssetResult, err := json.Marshal(remainArray)
	if err != nil {
		return nil, nil, errors.New("marshal user's asset failed")
	}

	return userAssetResult, dest, nil
}

func calculateInsert(userAssetString []byte, inserts []userAsset) ([]byte, error) {
	var userAssets []userAsset
	err := json.Unmarshal(userAssetString, &userAssets)
	if err != nil {
		return nil, errors.New("Failed decodinf owner")
	}
	var index = 0
	for _, insert := range inserts {
		fmt.Println("1111")
		for j, targetAsset := range userAssets[index:] {
			fmt.Println("1111")
			if insert.Expire == targetAsset.Expire {
				userAssets[index + j].Balance += insert.Balance
				fmt.Printf("target asset balance: %d", targetAsset.Balance)
				index = j + 1
				break
			} else if insert.Expire < targetAsset.Expire {
				asset := userAsset{insert.Expire, insert.Balance}
				startIndex := index + j
				fmt.Printf("start index: %d\n", startIndex)
				userAssets = append(userAssets[:startIndex], append([]userAsset{asset}, userAssets[startIndex:]...)...)
				index = startIndex + 1
				break
			} else {
				asset := userAsset{insert.Expire, insert.Balance}
				userAssets = append(userAssets, asset)
				index = j + 1
			}
		}
	}
	userAssetResult, err := json.Marshal(userAssets)
	if err != nil {
		return nil, errors.New("marshal target user's asset failed")
	}
	return userAssetResult, nil
}

func (t *BonusManagementChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Transfer...")

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	serializedID, err := stub.GetCreator()
	sid := &SerializedIdentity{}
	err = proto.Unmarshal(serializedID, sid)
	owner := sid.IdBytes


	if err != nil {
		return shim.Error("Failed decodinf owner")
	}
	assetName := args[0]
	ownerKey := assetName + string(owner)

	targetUser := args[1]

	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("amount argument is incorrect")
	}
	if amount < 0  {
		return shim.Error("the amount must not negative")
	}
	lastExpire, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("last expire argument is incorrect")
	}
	if lastExpire < 0  {
		return shim.Error("the last expire must not negative")
	}

	fmt.Println("get asset for key: %s", ownerKey)
	// Verify ownership
	userAssetString, err := stub.GetState(ownerKey)
	if err != nil || userAssetString == nil {
		return shim.Error("the user did not have the asset:" + assetName)
	}

	userAssetResult, transferArray, err := calculateTransferArray(userAssetString, lastExpire, amount)
	if err != nil {
		return shim.Error("calculate transfer error:" + err.Error())
	}
	targetKey := assetName + targetUser
	var targetUserAssetResult []byte
	targetUserAssetString, err := stub.GetState(targetKey)
	if err != nil || targetUserAssetString ==  nil {
		targetUserAssetResult, err = json.Marshal(transferArray)
	} else {
		targetUserAssetResult, err = calculateInsert(targetUserAssetString, transferArray)
	}
	err = stub.PutState(ownerKey, userAssetResult)
	if err != nil {
		return shim.Error("store user's asset failed")
	}
	err = stub.PutState(targetKey, targetUserAssetResult)
	if err != nil {
		return shim.Error("store target user's asset failed")
	}

	fmt.Println("Transfer...done")

	return shim.Success(nil)
}


// Invoke will be called for every transaction.
// Supported functions are the following:
// "assign(asset, owner)": to assign ownership of assets. An asset can be owned by a single entity.
// Only an administrator can call this function.
// "transfer(asset, newOwner)": to transfer the ownership of an asset. Only the owner of the specific
// asset can call this function.
// An asset is any string to identify it. An owner is representated by one of his ECert/TCert.
func (t *BonusManagementChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle different functions
	if function == "issue" {
		// Assign ownership
		return t.issue(stub, args)
	} else if function == "assign" {
		// Assign ownership
		return t.assign(stub, args)
	} else if function == "transfer" {
		// Transfer ownership
		return t.transfer(stub, args)
	}else if function == "query" {
		// Query owner
		return t.query(stub, "user", args)
	} else if function == "queryOrg" {
		// Query owner
		return t.query(stub, "organization", args)
	}
	//return nil, nil
	return shim.Error("Received unknown function invocation")
}

func (t *BonusManagementChaincode) queryUserBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 2 {
		fmt.Println("Incorrect number of arguments. Expecting name of an asset to query")
		return shim.Error("Incorrect number of arguments. Expecting name of an asset to query")
	}

	// Who is the owner of the asset?

	owner := args[0]
	if err != nil {
		return shim.Error("Failed decoding owner")
	}

	assetName := args[1]
	fmt.Printf("Arg [%s]\n", assetName)
	ownerKey := assetName + owner
	userAssetString, err := stub.GetState(ownerKey)
	return shim.Success(userAssetString)
}

func (t *BonusManagementChaincode) queryOrganizationBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	assetName := args[0]
	userAssetString, err := stub.GetState(assetName)
	if err != nil {
		return shim.Error("Failed decoding owner")
	}
	fmt.Printf("query... asset name: %s, detail: %s!", assetName, string(userAssetString))
	return shim.Success(userAssetString)
}


// Query callback representing the query of a chaincode
// Supported functions are the following:
// "query(asset)": returns the owner of the asset.
// Anyone can invoke this function.
func (t *BonusManagementChaincode) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Printf("Query [%s]", function)

	if function == "user" {
		return t.queryUserBalance(stub, args)
	} else if function == "organization" {
		return t.queryOrganizationBalance(stub, args)
	} else {
		return shim.Error("Invalid query function name. Expecting \"query\"")

	}
}

func main() {
	err := shim.Start(new(BonusManagementChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
