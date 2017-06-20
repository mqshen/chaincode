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
type AlgorithmChaincode struct {
}

// Init method will be called during deployment.
// The deploy transaction metadata is supposed to contain the administrator cert
//func (t *BonusManagementChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
func (t *AlgorithmChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
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

func (t *AlgorithmChaincode) apply(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}

func (t *AlgorithmChaincode) auth(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}
// Invoke will be called for every transaction.
// Supported functions are the following:
// "assign(asset, owner)": to assign ownership of assets. An asset can be owned by a single entity.
// Only an administrator can call this function.
// "transfer(asset, newOwner)": to transfer the ownership of an asset. Only the owner of the specific
// asset can call this function.
// An asset is any string to identify it. An owner is representated by one of his ECert/TCert.
func (t *AlgorithmChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle different functions
	if function == "apply" {
		// Assign ownership
		return t.apply(stub, args)
	} else if function == "auth" {
		// Assign ownership
		return t.auth(stub, args)
	}
	//return nil, nil
	return shim.Error("Received unknown function invocation:" + function)
}

func main() {
	err := shim.Start(new(AlgorithmChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: " + err.Error())
	}
}