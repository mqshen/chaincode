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

type UserAsset struct {
	Expire     int    `json:"expire"`
	Amount	   int    `json:"amount"`
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
type AccountManagementChaincode struct {
}

// Init method will be called during deployment.
// The deploy transaction metadata is supposed to contain the administrator cert
//func (t *BonusManagementChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
func (t *AccountManagementChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Chaincode...")

	// Set the role of the users that are allowed to assign assets
	// The metadata will contain the role of the users that are allowed to assign assets
	adminCert, err := stub.GetCreator()
	if err != nil {
		return shim.Error("Failed getting createor" + err.Error())
	}

	fmt.Println("admin's cert is " + string(adminCert))

	if len(adminCert) == 0 {
		return shim.Error("Invalid assigner role. Empty.")
	}

	stub.PutState("admin", adminCert)

	return shim.Success(nil)
}

func (t *AccountManagementChaincode) issue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
		fmt.Println("admin cert:" + string(admin))
		fmt.Println("creator cert:" + string(creator))
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
	fmt.Println("Issue... asset name: " + assetName + ", detail: " + assetName, string(assetJSONasBytes))
	err = stub.PutState(assetName, assetJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("Issue...done!")

	return shim.Success(nil)
}

func (t *AccountManagementChaincode) assign(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Assign... arg length: %d", len(args))

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	assetName := args[0]
	fmt.Println("get Issue... asset name: %s" + assetName)
	assetJSONasBytes, err := stub.GetState(assetName)
	if err != nil {
		return shim.Error("asset state get failed, have not issued")
	}
	if assetJSONasBytes == nil {
		return shim.Error("asset have not issued:" + assetName)
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
	detail := args[2]

	ownerKey := assetName + user
	err = stub.PutState(ownerKey, []byte(detail))
	if err != nil {
		return shim.Error("store user's asset failed")
	}

	return shim.Success(nil)
}

// Invoke will be called for every transaction.
// Supported functions are the following:
// "assign(asset, owner)": to assign ownership of assets. An asset can be owned by a single entity.
// Only an administrator can call this function.
// "transfer(asset, newOwner)": to transfer the ownership of an asset. Only the owner of the specific
// asset can call this function.
// An asset is any string to identify it. An owner is representated by one of his ECert/TCert.
func (t *AccountManagementChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle different functions
	if function == "issue" {
		// Assign ownership
		return t.issue(stub, args)
	} else if function == "assign" {
		// Assign ownership
		return t.assign(stub, args)
	} else if function == "query" {
		// Query owner
		return t.query(stub, "user", args)
	} else if function == "queryOrg" {
		// Query owner
		return t.query(stub, "organization", args)
	}
	//return nil, nil
	return shim.Error("Received unknown function invocation:" + function)
}

func (t *AccountManagementChaincode) queryUserBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	fmt.Println("Arg [%s]" + assetName)
	ownerKey := assetName + owner
	userAssetString, err := stub.GetState(ownerKey)
	return shim.Success(userAssetString)
}

func (t *AccountManagementChaincode) queryOrganizationBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	assetName := args[0]
	userAssetString, err := stub.GetState(assetName)
	if err != nil {
		return shim.Error("Failed decoding owner")
	}
	fmt.Println("query... asset name: " + assetName + ", detail: " + string(userAssetString))
	return shim.Success(userAssetString)
}


// Query callback representing the query of a chaincode
// Supported functions are the following:
// "query(asset)": returns the owner of the asset.
// Anyone can invoke this function.
func (t *AccountManagementChaincode) query(stub shim.ChaincodeStubInterface, function string, args []string) pb.Response {
	fmt.Println("Query " + function)

	if function == "user" {
		return t.queryUserBalance(stub, args)
	} else if function == "organization" {
		return t.queryOrganizationBalance(stub, args)
	} else {
		return shim.Error("Invalid query function name. Expecting \"query\"")

	}
}

func main() {
	err := shim.Start(new(AccountManagementChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: " + err.Error())
	}
}