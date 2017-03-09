package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/base64"
	"bytes"
	"encoding/json"
)

const (
	insuranceTableColumn  = "InsuranceOwnership"
	bankTableColumn  = "BankOwnership"
	policyTableColumn  = "Policy"
	creditTableColumn  = "Credit"
	applyTableColumn  = "Apply"
	adminRole = "admin"
)

type Credit struct {
	Id      string 	 `json:"id,omitempty"`
	Company string 	 `json:"company,omitempty"`
	Bank 	string 	 `json:"bank,omitempty"`
	Expire  int32	 `json:"expire,omitempty"`
	Credit	int32	 `json:"credit,omitempty"`
	Rate    int32	 `json:"rate,omitempty"`
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
	State   int32	 `json:"state,omitempty"`
	Balance int32	 `json:"balance,omitempty"`
	Credits []Credit `json:"credits"`
	Applied *Apply   `json:"applied"`
}



// SimpleChaincode example simple Chaincode implementation
type InsuranceChaincode struct {
}

func (t *InsuranceChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	callerCertificate, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Failed getting call certificate, [%v]", err)
	}
	if len(callerCertificate) == 0 {
		return nil, errors.New("Invalid call asset role. Empty.")
	}


	// Create ownership table
	err = stub.CreateTable(insuranceTableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Company", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Balance", Type: shim.ColumnDefinition_INT32, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating InsuranceOwnership table.")
	}

	err = stub.CreateTable(bankTableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Bank", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Balance", Type: shim.ColumnDefinition_INT32, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating BankOwnership table.")
	}

	// Create assets issue table
	err = stub.CreateTable(policyTableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Owner", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Id", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Company", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "State", Type: shim.ColumnDefinition_INT32, Key: false},
		&shim.ColumnDefinition{Name: "Balance", Type: shim.ColumnDefinition_INT32, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating Policy table.")
	}


	// Create assets issue table
	err = stub.CreateTable(creditTableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Id", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Company", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Bank", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Expire", Type: shim.ColumnDefinition_INT32, Key: false},
		&shim.ColumnDefinition{Name: "Credit", Type: shim.ColumnDefinition_INT32, Key: false},
		&shim.ColumnDefinition{Name: "Rate", Type: shim.ColumnDefinition_INT32, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating credit table.")
	}

	// Create assets issue table
	err = stub.CreateTable(applyTableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Id", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Company", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Bank", Type: shim.ColumnDefinition_BYTES, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating credit table.")
	}

	stub.PutState(adminRole, callerCertificate)
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
	company, err := base64.StdEncoding.DecodeString(args[0])
	balance, err := strconv.Atoi(args[1])

	if isInsuranceOrBank(company, stub, 0) {
		return nil, fmt.Errorf("Failed this company is alreay issue insurance", err)
	}

	_, err = stub.InsertRow(
		insuranceTableColumn,
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_Bytes{Bytes: company}},
				&shim.Column{Value: &shim.Column_Int32{Int32: int32(balance)}}},
		})
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
	balance, err := strconv.Atoi(args[1])

	if isInsuranceOrBank(company, stub, 1) {
		return nil, fmt.Errorf("Failed this bank is alreay apply", err)
	}

	_, err = stub.InsertRow(
		bankTableColumn,
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_Bytes{Bytes: company}},
				&shim.Column{Value: &shim.Column_Int32{Int32: int32(balance)}}},
		})
	if err != nil {
		return nil, fmt.Errorf("the caller is not admin, [%v]", err)
	}

	return nil, nil
}

func isInsuranceOrBank(company []byte, stub shim.ChaincodeStubInterface, classify int) (bool) {
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes: company}}
	columns = append(columns, col1)

	tableName := insuranceTableColumn
	if classify != 0 {
		tableName = bankTableColumn
	}

	rowChannel, err := stub.GetRows(tableName, columns)
	if err != nil {
		return false
	}
	var allRows []shim.Row

	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				allRows = append(allRows, row)
			}
		}
		if rowChannel == nil {
			break
		}
	}
	return len(allRows) > 0
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
	if !isInsuranceOrBank(callerCertificate, stub, 0) {
		return nil, errors.New("caller is not insurance company")
	}
	owner, err := base64.StdEncoding.DecodeString(args[0])
	id := args[1]
	balance, err := strconv.Atoi(args[2])

	_, err = stub.InsertRow(
		insuranceTableColumn,
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_Bytes{Bytes: owner}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: []byte(id)}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: callerCertificate}},
				&shim.Column{Value: &shim.Column_Int32{Int32: 0}},
				&shim.Column{Value: &shim.Column_Int32{Int32: int32(balance)}}},
		})
	if err != nil {
		return nil, errors.New("Failed assign insurance")
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
	if !isInsuranceOrBank(callerCertificate, stub, 1) {
		return nil, errors.New("caller is not insurance company")
	}
	company, err := base64.StdEncoding.DecodeString(args[0])
	id := args[1]
	expire, err := strconv.Atoi(args[2])
	credit, err := strconv.Atoi(args[3])
	rate, err := strconv.Atoi(args[4])

	_, err = stub.InsertRow(
		creditTableColumn,
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_Bytes{Bytes: []byte(id)}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: company}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: callerCertificate}},
				&shim.Column{Value: &shim.Column_Int32{Int32: int32(expire)}},
				&shim.Column{Value: &shim.Column_Int32{Int32: int32(credit)}},
				&shim.Column{Value: &shim.Column_Int32{Int32: int32(rate)}}},
		})
	if err != nil {
		return nil, errors.New("Failed credit")
	}
	return nil, nil
}

func (t *InsuranceChaincode) apply(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("start assign")

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	owner, err := stub.GetCallerCertificate()
	company, err := base64.StdEncoding.DecodeString(args[0])
	id := args[1]
	bank, err := base64.StdEncoding.DecodeString(args[2])


	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes: owner}}
	columns = append(columns, col1)

	col2 := shim.Column{Value: &shim.Column_String_{String_: id}}
	columns = append(columns, col2)

	col3 := shim.Column{Value: &shim.Column_Bytes{Bytes: company}}
	columns = append(columns, col3)


	rowChannel, err := stub.GetRows(policyTableColumn, columns)
	if err != nil {
		return nil, errors.New("failed to query policy")
	}
	var allRows []shim.Row

	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				allRows = append(allRows, row)
			}
		}
		if rowChannel == nil {
			break
		}
	}
	if len(allRows) == 0 {
		return nil, fmt.Errorf("Failed getting user's issurance: %s", id)
	}

	_, err = stub.InsertRow(
		applyTableColumn,
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: id}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: company}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: bank}}},
		})
	if err != nil {
		return nil, errors.New("Failed apply")
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

func queryBank(stub shim.ChaincodeStubInterface, id string, company []byte) ([]Credit, error) {
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: id}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_Bytes{Bytes: company}}
	columns = append(columns, col2)

	rowChannel, err := stub.GetRows(creditTableColumn, columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to queyr credit table ")
	}

	var credits []Credit
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				bank := row.Columns[3].GetString_()
				expire := row.Columns[4].GetInt32()
				limit := row.Columns[5].GetInt32()
				rate := row.Columns[6].GetInt32()
				credit := Credit{id, string(company), bank, expire, limit, rate}
				credits = append(credits, credit)
			}
		}
		if rowChannel == nil {
			break
		}
	}
	return credits, nil
}

func queryApply(stub shim.ChaincodeStubInterface, id string, company []byte) (*Apply, error) {
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: id}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_Bytes{Bytes: company}}
	columns = append(columns, col2)

	rowChannel, err := stub.GetRows(applyTableColumn, columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to queyr credit table ")
	}

	var apply *Apply
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				bank := row.Columns[3].GetString_()
				apply = &Apply{id, string(company), bank}
				break
			}
		}
		if rowChannel == nil {
			break
		}
	}
	return apply, nil
}

func (t *InsuranceChaincode) queryUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of an asset to query 1")
	}
	owner, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return nil, fmt.Errorf("Failed decoding owner")
	}
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes: owner}}
	columns = append(columns, col1)
	rowChannel, err := stub.GetRows(policyTableColumn, columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to query user")
	}
	var polices []Policy

	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				id := row.Columns[1].GetString_()
				company := row.Columns[2].GetBytes()
				state := row.Columns[3].GetInt32()
				balance := row.Columns[3].GetInt32()
				credits, err := queryBank(stub, id, company)
				if err != nil {
					return nil, fmt.Errorf("Failed query credit")
				}
				apply, err := queryApply(stub, id, company)
				if err != nil {
					return nil, fmt.Errorf("Failed query apply")
				}
				policy := Policy{string(owner), id, string(company), state, balance, credits, apply}
				polices = append(polices, policy)
			}
		}
		if rowChannel == nil {
			break
		}
	}
	jSONasBytes, _ := json.Marshal(polices)
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