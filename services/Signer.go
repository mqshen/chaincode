package services

func newIdentity(client *Client, name string, key []byte, cert []byte) *Identity {
	id := new(Identity)
	id.name = name
	id.ecert = NewSigner(key, cert, id)
	id.client = client
	return id
}

// Signer represents a signer
// Each identity may have multiple signers, currently one ecert and multiple tcerts
type Signer struct {
	name   string
	key    []byte
	cert   []byte
	id     *Identity
	client *Client
}
