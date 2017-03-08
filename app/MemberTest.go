package main

import (
	"github.com/chaincode/services"
	"github.com/hyperledger/fabric-cop/util"
	"fmt"
)

func main() {
	c := new(services.Client)
	// Set defaults
	c.ServerURL = "http://192.168.30.98:8888"
	c.HomeDir = "/Volumes/disk02/WorkspaceGroup/BlockchainWorkspace/src/github.com/chaincode/cop"//util.GetDefaultHomeDir()

	id := new(services.Identity)

	certPath := "/Volumes/disk02/WorkspaceGroup/BlockchainWorkspace/certs/"
	key, err := util.ReadFile(certPath + "key.pem")
	if err != nil {
		fmt.Printf("read key file filed: %s", err)
		return
	}
	cert, err := util.ReadFile(certPath + "cert.pem")
	if err != nil {
		fmt.Printf("read cert file filed: %s", err)
		return
	}

	ecert := services.NewSigner(key, cert, id)
	memberService := services.NewMemberSErvice(c, ecert)
	name := "test_for_test6"
	secret, err := memberService.Register(name, "client", "bank_a")
	if err != nil {
		fmt.Printf("read key file filed: %s", err)
		return
	}
	fmt.Printf("get secret: %s\n", secret)
	enrollId, err := memberService.Enroll(name, secret)
	if err != nil {
		fmt.Printf("enroll user filed: %s", err)
		return
	}
	fmt.Printf("enroll user success")
	enrollId.Store("./cop/certs/")
}
