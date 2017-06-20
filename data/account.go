package main

import (
	"fmt"
	"github.com/mqshen/data/belinkcode/shim"
	"github.com/mqshen/data/protos"
	"encoding/json"
)

type AccountManagementBelinkcode struct {
}

func (t *AccountManagementBelinkcode) Init(stub shim.BelinkcodeStubInterface) protos.Response {
	fmt.Println("success init belink code")
	return shim.Success([]byte("test"))
}
// Invoke is called for every Invoke transactions. The chaincode may change
// its state variables
func (t *AccountManagementBelinkcode) Invoke(stub shim.BelinkcodeStubInterface) protos.Response {
	fmt.Println("success invoke belink code")
	results, error := stub.GetQueryResult("wl_bonus_transaction", "100161", "bonus > 5")
	if error != nil {
		return shim.Error(fmt.Sprintf("faile query result:", error.Error()))
	}
	result, error := json.Marshal(results)
	if error != nil {
		return shim.Error(fmt.Sprintf("format result error:", error.Error()))
	}
	return shim.Success([]byte(result))
}

func main() {
	err := shim.Start(new(AccountManagementBelinkcode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: " + err.Error())
	}
}