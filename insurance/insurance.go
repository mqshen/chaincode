package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/base64"
	"bytes"
	"encoding/json"
	"strings"
)

const (
	insurancePrefix = "insurance_"
	bankPrefix = "bank_"
	userPrefix = "user_"
	creditPrefix = "credit_"
	applyPrefix = "apply_"
	adminRole = "admin"
)

type Credit struct {
	Id      string 	 `json:"id,omitempty"`
	Company string 	 `json:"company,omitempty"`
	Bank 	string 	 `json:"bank,omitempty"`
	Expire  int	 `json:"expire,omitempty"`
	Credit	int	 `json:"credit,omitempty"`
	Rate    int	 `json:"rate,omitempty"`
}

type Apply struct {
	Id      string 	 `json:"id,omitempty"`
	Company string 	 `json:"company,omitempty"`
	Bank 	string 	 `json:"bank,omitempty"`
}

type Policy struct {
	Owner   string   `json:"owner,omitempty"`
	Id      string 	 `json:"id,omitempty"`
	Company string 	 `json:"company,omitempty"`
	State   int	 `json:"state,omitempty"`
	Balance int	 `json:"balance,omitempty"`
}

type PolicyResult struct {
	Insurance Policy   `json:"insurance,omitempty"`
	Credits   []Credit `json:"credits,omitempty"`
	Applied   *Apply   `json:"Applied,omitempty"`
}

// SimpleChaincode example simple Chaincode implementation
type InsuranceChaincode struct {
}

func (t *InsuranceChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	admin, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	if len(admin) == 0 {
		return nil, errors.New("Invalid call asset role. Empty.")
	}

	stub.PutState(adminRole, admin)
	return nil, nil
}

// Transaction makes payment of X units from A to B
func (t *InsuranceChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "issue" {
		// Assign ownership
		return t.issue(stub, args)
	} else if function == "bank" {
		// Assign ownership
		return t.bank(stub, args)
	} else if function == "assign" {
		// Assign ownership
		return t.assign(stub, args)
	} else if function == "credit" {
		// Transfer ownership
		return t.credit(stub, args)
	} else if function == "apply" {
		// Transfer ownership
		return t.apply(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

func (t *InsuranceChaincode) issue(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("start issue")
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	callerCertificate, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	admin, err := stub.GetState(adminRole)
	if err != nil {
		return nil, fmt.Errorf("Failed getting admin certificate, [%v]", err)
	}
	if bytes.Compare(callerCertificate, admin) != 0 {
		return nil, fmt.Errorf("the caller is not admin")
	}

	key := insurancePrefix + args[0]
	if isInsuranceOrBank(key, stub) {
		return nil, fmt.Errorf("Failed this company is alreay issue insurance", err)
	}

	stub.PutState(key, []byte(args[1]))
	if err != nil {
		return nil, fmt.Errorf("the caller is not admin, [%v]", err)
	}

	return nil, nil
}

func (t *InsuranceChaincode) bank(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("start issue")
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	callerCertificate, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	admin, err := stub.GetState(adminRole)
	if err != nil {
		return nil, fmt.Errorf("Failed getting admin certificate, [%v]", err)
	}
	if bytes.Compare(callerCertificate, admin) != 0 {
		return nil, fmt.Errorf("the caller is not admin")
	}
	company, err := base64.StdEncoding.DecodeString(args[0])

	fmt.Printf("issue for company: [%x]", company)
	key := bankPrefix + args[0]
	if isInsuranceOrBank(key, stub) {
		return nil, fmt.Errorf("Failed this company is alreay issue insurance", err)
	}

	stub.PutState(key, []byte(args[1]))
	if err != nil {
		return nil, fmt.Errorf("the caller is not admin, [%v]", err)
	}

	return nil, nil
}

func isInsuranceOrBank(key string, stub shim.ChaincodeStubInterface) (bool) {
	balance, err := stub.GetState(key)

	if err != nil || balance != nil {
		return false
	}

	return true
}

func (t *InsuranceChaincode) assign(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("start assign")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	callerCertificate, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	company := base64.StdEncoding.EncodeToString(callerCertificate)
	if !isInsuranceOrBank(company, stub) {
		return nil, fmt.Errorf("caller is not insurance company, [%x]", callerCertificate)
	}
	owner := args[0]
	id := args[1]
	balance, err := strconv.Atoi(args[2])
	key := userPrefix + owner

	policy := Policy{owner, id, company, 0, balance}

	var policies []Policy
	userAssetString, err := stub.GetState(key)
	if err != nil || userAssetString == nil {
		policies = []Policy{policy}
	} else {
		err = json.Unmarshal(userAssetString, &policies)
		if err != nil {
			return nil, fmt.Errorf("unmarshal user balance failed " + err.Error())
		}
		policies = append(policies, policy)
	}

	userAssetResult, err := json.Marshal(policies)
	if err != nil {
		return nil, fmt.Errorf("marshal user's asset failed")
	}

	fmt.Println("set key: %s , detail: %s", key, string(userAssetResult))
	err = stub.PutState(key, userAssetResult)
	if err != nil {
		return nil, fmt.Errorf("store user's asset failed")
	}
	return nil, nil
}

func (t *InsuranceChaincode) credit(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("start assign")

	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}
	callerCertificate, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	bankId := base64.StdEncoding.EncodeToString(callerCertificate)
	key := bankPrefix + bankId
	if !isInsuranceOrBank(key, stub) {
		return nil, errors.New("caller is not bank")
	}
	company := args[0]
	id := args[1]
	expire, err := strconv.Atoi(args[2])
	limit, err := strconv.Atoi(args[3])
	rate, err := strconv.Atoi(args[4])
	credit := Credit{id, company, bankId, expire, limit, rate}

	insuranceId := creditPrefix + company + id
	var credits []Credit
	userAssetString, err := stub.GetState(insuranceId)
	if err != nil || userAssetString == nil {
		credits = []Credit{credit}
	} else {
		err = json.Unmarshal(userAssetString, &credits)
		if err != nil {
			return nil, fmt.Errorf("unmarshal user balance failed " + err.Error())
		}
		for _, c := range credits {
			if strings.Compare(c.Bank, bankId) == 0 {
				return nil, fmt.Errorf("already credit this insurance")
			}
		}
		credits = append(credits, credit)
	}
	userAssetResult, err := json.Marshal(credits)
	if err != nil {
		return nil, fmt.Errorf("marshal user's asset failed")
	}

	fmt.Println("set key: %s , detail: %s", key, string(userAssetResult))
	err = stub.PutState(key, userAssetResult)
	if err != nil {
		return nil, fmt.Errorf("store user's asset failed")
	}

	return nil, nil
}

func (t *InsuranceChaincode) apply(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("start assign")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	owner, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	user := base64.StdEncoding.EncodeToString(owner)
	company := args[0]
	id := args[1]
	bank := args[2]

	key := userPrefix + user

	var policies []Policy
	userAssetString, err := stub.GetState(key)
	if err != nil || userAssetString == nil {
		return nil, fmt.Errorf("user did not have this insurance")
	} else {
		err = json.Unmarshal(userAssetString, &policies)
		if err != nil {
			return nil, fmt.Errorf("unmarshal user balance failed " + err.Error())
		}
		flag := false
		for _, p := range policies {
			if strings.Compare(p.Company, company) == 0 && strings.Compare(p.Id, id) == 0{
				flag = true
				break
			}
		}
		if !flag {
			return nil, fmt.Errorf("user did not have this insurance")
		}
	}

	applyId := applyPrefix + company + id
	applyString, err := stub.GetState(applyId)
	if err != nil || applyString != nil {
		return nil, fmt.Errorf("this insurance is already apply")
	}
	apply := Apply{id, company, bank}
	userAssetResult, err := json.Marshal(apply)
	if err != nil {
		return nil, fmt.Errorf("marshal user's asset failed")
	}

	fmt.Println("set key: %s , detail: %s", key, string(userAssetResult))
	err = stub.PutState(key, userAssetResult)
	if err != nil {
		return nil, fmt.Errorf("store user's asset failed")
	}

	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *InsuranceChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "user" {
		return t.queryUser(stub, args)
	} else if function == "bank" {
		return t.queryBank(stub, args)
	} else {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	return nil, nil
}

func queryBank(stub shim.ChaincodeStubInterface, id , company string) ([]Credit, error) {
	insuranceId := creditPrefix + company + id
	var credits []Credit
	userAssetString, err := stub.GetState(insuranceId)
	if err != nil || userAssetString == nil {
		return credits, nil
	} else {
		err = json.Unmarshal(userAssetString, &credits)
		if err != nil {
			return nil, fmt.Errorf("unmarshal user balance failed " + err.Error())
		}
	}
	return credits, nil
}

func queryApply(stub shim.ChaincodeStubInterface, id, company string) (*Apply, error) {
	insuranceId := applyPrefix + company + id
	var apply Apply
	userAssetString, err := stub.GetState(insuranceId)
	if err != nil || userAssetString == nil {
		return nil, nil
	} else {
		err = json.Unmarshal(userAssetString, &apply)
		if err != nil {
			return nil, fmt.Errorf("unmarshal user balance failed " + err.Error())
		}
	}
	return &apply, nil
}

func (t *InsuranceChaincode) queryUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of an asset to query 1")
	}
	owner := args[0]
	key := userPrefix + owner

	var policies []Policy
	var issurances []PolicyResult
	userAssetString, err := stub.GetState(key)
	if err != nil || userAssetString == nil {
		return nil, nil
	} else {
		err = json.Unmarshal(userAssetString, &policies)
		if err != nil {
			return nil, fmt.Errorf("unmarshal user balance failed " + err.Error())
		}
		for _, p := range policies {
			credits, error := queryBank(stub, p.Id, p.Company)
			if error != nil {
				return nil, fmt.Errorf("query credit failed" + err.Error())
			}
			apply, error := queryApply(stub, p.Id, p.Company)
			if error != nil {
				return nil, fmt.Errorf("query credit apply" + err.Error())
			}
			result := PolicyResult{p, credits, apply}
			issurances = append(issurances, result)
		}
	}
	jSONasBytes, err := json.Marshal(issurances)
	if err != nil || userAssetString == nil {
		return nil, fmt.Errorf("failed marshal issurance" + err.Error())
	}
	return jSONasBytes, nil
}

func (t *InsuranceChaincode) queryBank(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	return nil, nil
}

func main() {
	err := shim.Start(new(InsuranceChaincode))
	if err != nil {
		fmt.Printf("Error starting insurance chaincode: %s", err)
	}
}