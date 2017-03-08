package services

import (
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/platforms"
	"github.com/hyperledger/fabric/core/container"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/protos/utils"
	putils "github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"golang.org/x/net/context"
	"encoding/json"
	"github.com/hyperledger/fabric/peer/common"
)

const (
	chainFuncName = "chaincode"
)

type PeerServices struct {
	Signer          msp.SigningIdentity
	EndorserClient  pb.EndorserClient
}

func NewPeerServices()(*PeerServices, error)  {
	endorserClient, err := common.GetEndorserClient()
	if err != nil {
		return nil, fmt.Errorf("Error getting endorser client %s: %s", chainFuncName, err)
	}
	signer, err := common.GetDefaultSigner()
	if err != nil {
		return nil, fmt.Errorf("Error getting default signer: %s", err)
	}
	if signer == nil {
		return nil, fmt.Errorf("Error default signer is nil")
	}
	fmt.Println(signer)
	return &PeerServices{
		Signer:          signer,
		EndorserClient:  endorserClient,
	}, nil
}

func (peer *PeerServices) Deploy(chaincodeName, chaincodePath string) (string, error) {
	input := &pb.ChaincodeInput{}
	if err := json.Unmarshal([]byte("{\"args\":[\"111\"]}"), &input); err != nil {
		return "", fmt.Errorf("Chaincode argument error: %s", err)
	}

	chaincodeLang := "GOLANG"
	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[chaincodeLang]),
		ChaincodeId: &pb.ChaincodeID{Path: chaincodePath, Name: chaincodeName},
		Input:       input,
	}

	cds, err := getChaincodeBytes(spec)

	if err != nil {
		return "", fmt.Errorf("Error getting chaincode code %s: %s", chainFuncName, err)
	}

	creator, err := peer.Signer.Serialize()
	if err != nil {
		return "", fmt.Errorf("Error serializing identity for %s: %s\n", peer.Signer.GetIdentifier(), err)
	}

	uuid := util.GenerateUUID()

	p := cauthdsl.SignedByMspMember("DEFAULT")
	policyMarhsalled := putils.MarshalOrPanic(p)

	escc := "escc"
	vscc := "vscc"

	prop, err := utils.CreateDeployProposalFromCDS(uuid, "testchainid", cds, creator, policyMarhsalled, []byte(escc), []byte(vscc))
	if err != nil {
		return "", fmt.Errorf("Error creating proposal  %s: %s\n", chainFuncName, err)
	}

	var signedProp *pb.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, peer.Signer)
	if err != nil {
		return "", fmt.Errorf("Error creating signed proposal  %s: %s\n", chainFuncName, err)
	}

	proposalResponse, err := peer.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return "", fmt.Errorf("Error endorsing %s: %s\n", chainFuncName, err)
	}

	if proposalResponse != nil {
		// assemble a signed transaction (it's an Envelope message)
		env, err := utils.CreateSignedTx(prop, peer.Signer, proposalResponse)
		if err != nil {
			return "", fmt.Errorf("Could not assemble transaction, err %s", err)
		}
		fmt.Print(env)

		return "", nil
	}
	return "", nil
}


// getChaincodeBytes get chaincode deployment spec given the chaincode spec
func getChaincodeBytes(spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	var codePackageBytes []byte
	var err error
	if err = checkSpec(spec); err != nil {
		return nil, err
	}

	codePackageBytes, err = container.GetChaincodePackageBytes(spec)
	if err != nil {
		err = fmt.Errorf("Error getting chaincode package bytes: %s", err)
		return nil, err
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return chaincodeDeploymentSpec, nil
}

// checkSpec to see if chaincode resides within current package capture for language.
func checkSpec(spec *pb.ChaincodeSpec) error {
	// Don't allow nil value
	if spec == nil {
		return errors.New("Expected chaincode specification, nil received")
	}

	platform, err := platforms.Find(spec.Type)
	if err != nil {
		return fmt.Errorf("Failed to determine platform type: %s", err)
	}

	return platform.ValidateSpec(spec)
}