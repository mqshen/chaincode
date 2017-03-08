package services

import (
	"github.com/hyperledger/fabric-cop/api"
	"github.com/hyperledger/fabric-cop/util"
	"fmt"
	"net/http"
	"errors"
	"github.com/cloudflare/cfssl/signer"
	"encoding/base64"
)

func NewSigner(key, cert []byte, id *Identity) *Signer {
	return &Signer{
		key:    key,
		cert:   cert,
		id:     id,
		client: id.client,
	}
}

type MemberServices struct {
	//GetTCertBatch(req pb.TCertCreateReq) (pb.CertSet, error)
	//GetTCertsFromTCA(enrollPrivKey *ecdsa.PrivateKey, id string, attrhash string, attributes []string, num int) (tCertBlocks []*TCertBlock, error error)
	//
	//GetTcert(userName string, assigner *ecdsa.PrivateKey) (*TCertBlock, error)
	client *Client
	ecert  *Signer
}

func NewMemberSErvice(c *Client, ecert *Signer) (*MemberServices) {
	return &MemberServices{client: c, ecert: ecert}
}

func (member *MemberServices) Register(id string, clientType string, group string) (string, error) {
	attrs := []api.Attribute{api.Attribute{"test", "testValue"},}
	request := &api.RegistrationRequest{Name: id, Type: clientType, Group:group, Attributes: attrs}

	reqBody, err := util.Marshal(request, "RegistrationRequest")
	if err != nil {
		return "", err
	}

	secret, err := member.Post("register", reqBody)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", errors.New("failed to get secret")
	}
	secretBytes, err := base64.StdEncoding.DecodeString(secret.(string))

	return string(secretBytes), nil
}

func (member *MemberServices) Enroll(id string, secret string) (*Identity, error) {
	// Generate the CSR
	req := &api.EnrollmentRequest{
		Name:   id,
		Secret: secret,
	}

	csrPEM, key, err := member.client.GenCSR(nil, req.Name)
	if err != nil {
		return nil, err
	}

	// Get the body of the request
	sreq := signer.SignRequest{
		Hosts:   signer.SplitHosts(req.Hosts),
		Request: string(csrPEM),
		Profile: req.Profile,
		Label:   req.Label,
	}
	body, err := util.Marshal(sreq, "SignRequest")
	if err != nil {
		return nil, err
	}

	// Send the CSR to the COP server with basic auth header
	post, err := member.client.NewPost("enroll", body)
	if err != nil {
		return nil, err
	}
	post.SetBasicAuth(req.Name, req.Secret)
	result, err := member.client.SendPost(post)
	if err != nil {
		return nil, err
	}

	// Create an identity from the key and certificate in the response
	return member.client.newIdentityFromResponse(result, req.Name, key)
}

func (member *MemberServices) Post(endpoint string, reqBody []byte) (interface{}, error) {
	req, err := member.client.NewPost(endpoint, reqBody)
	if err != nil {
		return nil, err
	}
	err = member.addTokenAuthHdr(req, reqBody)
	if err != nil {
		return nil, err
	}
	return member.client.SendPost(req)
}

func (member *MemberServices) addTokenAuthHdr(req *http.Request, body []byte) error {
	cert := member.ecert.cert
	key := member.ecert.key
	token, err := util.CreateToken(cert, key, body)
	if err != nil {
		return fmt.Errorf("Failed to add token authorization header: %s", err)
	}
	req.Header.Set("authorization", token)
	return nil
}

