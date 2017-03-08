package services

import (
	"fmt"
	"github.com/hyperledger/fabric-cop/util"
)


// Identity is COP's implementation of an identity
type Identity struct {
	name   string
	ecert  *Signer
	client *Client
}

// Store writes my identity info to disk
func (i *Identity) Store(filePath string) error {
	if i.client == nil {
		return fmt.Errorf("An identity with no client may not be stored")
	}
	err := util.WriteFile(filePath + i.name + "_key.pem", i.ecert.key, 0600)
	if err != nil {
		return err
	}
	return util.WriteFile(filePath + i.name + "_cert.pem", i.ecert.cert, 0644)
}