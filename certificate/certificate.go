package main

import (
	"fmt"
	"errors"
	"encoding/json"

	"github.com/hyperledger/fabric/accesscontrol/impl"
	"github.com/hyperledger/fabric/core/chaincode/shim"

	pb "github.com/hyperledger/fabric/protos/peer"
	"bytes"
	"encoding/base64"
)

// SimpleChaincode example simple Chaincode implementation
type CertificateChaincode struct {
}

type issuer struct {
	issuer     string `json:"issuer"`
	CertType   string `json:"certType"`
}

type certificate struct {
	CertType   string `json:"certType"`
	ID	   string `json:"id"`
	State      int    `json:"state"`  //0:正常， 1:吊销
	content    string `json:"content"`
	Owner      string `json:"owner"`
}

var indexName = "owner~organize~cert~id"

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(CertificateChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *CertificateChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// Set the admin
	// The metadata will contain the certificate of the administrator
	//adminCert, err := stub.GetCallerCertificate()
	//if err != nil {
	//	fmt.Println("Failed getting metadata")
	//	return shim.Error("Failed getting metadata.")
	//}
	//if len(adminCert) == 0 {
	//	fmt.Println("Invalid admin certificate. Empty.")
	//	return shim.Error("Invalid admin certificate. Empty.")
	//}
	adminCert := []byte{0x00}//stub.GetCallerCertificate()

	fmt.Printf("The administrator is [%x]", adminCert)

	stub.PutState("admin", adminCert)

	fmt.Println("Init Chaincode...done")

	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *CertificateChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "issue" { //create a new marble
		return t.issue(stub, args)
	} else if function == "assign" { //change owner of a specific marble
		return t.assign(stub, args)
	} else if function == "append" { //transfer all marbles of a certain color
		return t.update(stub, args)
	} else if function == "query" { //transfer all marbles of a certain color
		return t.query(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *CertificateChaincode) issue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Verify the identity of the caller
	// Only an administrator can invoker assign
	adminCertificate, err := stub.GetState("admin")
	if err != nil {
		return shim.Error("Failed fetching admin identity")
	}

	ok, err := t.isCaller(stub, adminCertificate)
	if err != nil {
		return shim.Error("Failed checking admin identity")
	}
	if !ok {
		return shim.Error("The caller is not an administrator")
	}

	organizeId := args[0]
	certName := args[1]
	organizeCert := args[2]

	certType := organizeId + "-" + certName
	oldOrganizeCert, err := stub.GetState(certType)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if oldOrganizeCert != nil {
		errorMsg := "This cert name: " + certName + " for " + organizeId + " already exists "
		fmt.Println(errorMsg)
		return shim.Error(errorMsg)
	}

	organizeCertByte, err := base64.StdEncoding.DecodeString(organizeCert)
	if err != nil {
		return shim.Error("Failed decodinf organizeCert")
	}

	err = stub.PutState(certType, organizeCertByte)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end issue certificate")
	return shim.Success(nil)
}

func (t *CertificateChaincode) assign(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	organizeId := args[0]
	certName := args[1]
	certType := organizeId + "-" + certName
	organizeCert, err := stub.GetState(certType)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if organizeCert == nil {
		errorMsg := "This cert name: " + certName + " for " + organizeId + " is not exists "
		fmt.Println(errorMsg)
		return shim.Error(errorMsg)
	}

	ok, err := t.isCaller(stub, organizeCert)
	if err != nil {
		return shim.Error("Failed checking organize owner")
	}
	if !ok {
		return shim.Error("The caller is organize owner")
	}
	id := args[2]
	content := args[3]
	owner := args[4]
	cert := &certificate{certType, id, 0, content, owner}
	certJSONasBytes, _ := json.Marshal(cert)



	ownerKey := owner + certType + id
	err = stub.PutState(ownerKey, certJSONasBytes) //rewrite the marble
	if err != nil {
		return shim.Error(err.Error())
	}

	_, err = stub.CreateCompositeKey(indexName, []string{cert.Owner, organizeId, certName, cert.ID})
	if err != nil {
		return shim.Error(err.Error())
	}
	////  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	////  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	//value := []byte{0x00}
	//stub.PutState(colorNameIndexKey, value)

	fmt.Println("- end assign certificate")
	return shim.Success(nil)
}

func (t *CertificateChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("- end update certificate")
	return shim.Success(nil)
}

func (t *CertificateChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	owner := args[0]

	certResultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer certResultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	var i int
	for i = 0; certResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		ownerKey, _, err := certResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(ownerKey)
		currentOwner := compositeKeyParts[0]
		organizeId := compositeKeyParts[1]
		certName := compositeKeyParts[2]
		certId := compositeKeyParts[3]
		certType := organizeId + "-" + certName
		index := currentOwner + certType + certId

		certAsBytes, err := stub.GetState(index)
		if err != nil {
			return shim.Error("Failed to get cert: " + err.Error())
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(certAsBytes))
		bArrayMemberAlreadyWritten = true
		fmt.Printf("- found a marble from index:%s color:%s name:%s\n  ,id : %s", objectType, organizeId, certName, certId)
	}
	queryResults := buffer.Bytes()
	return shim.Success(queryResults)
}

func (t *CertificateChaincode) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	fmt.Println("Check caller...")

	// In order to enforce access control, we require that the
	// metadata contains the signature under the signing key corresponding
	// to the verification key inside certificate of
	// the payload of the transaction (namely, function name and args) and
	// the transaction binding (to avoid copying attacks)

	// Verify \sigma=Sign(certificate.sk, tx.Payload||tx.Binding) against certificate.vk
	// \sigma is in the metadata

	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed getting metadata")
	}
	payload, err := stub.GetPayload()
	if err != nil {
		return false, errors.New("Failed getting payload")
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed getting binding")
	}

	fmt.Printf("passed certificate [% x] \n", certificate)
	fmt.Printf("passed sigma [% x] \n", sigma)
	fmt.Printf("passed payload [% x] \n", payload)
	fmt.Printf("passed binding [% x] \n", binding)

	ok, err := impl.NewAccessControlShim(stub).VerifySignature(
		certificate,
		sigma,
		append(payload, binding...),
	)
	if err != nil {
		fmt.Printf("Failed checking signature [%s] \n", err)
		return ok, err
	}
	if !ok {
		fmt.Println("Invalid signature")
	}

	fmt.Println("Check caller...Verified!")

	return ok, err
}