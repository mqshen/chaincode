package main

import (
	"github.com/mqshen/BonusLedger/service"
	"github.com/mqshen/BonusLedger/crypto/primitives"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"crypto/ecdsa"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/op/go-logging"
	"os"

	"github.com/spf13/viper"
	"strings"
	"path/filepath"
	"fmt"
	"time"
	"encoding/base64"
)

var (
	logger = logging.MustGetLogger("app")
	fileRoot = "/Volumes/disk02/WorkspaceGroup/BlockchainWorkspace/hyperledger/certs/"
	testChaincodePath = "github.com/bonus/"
)

func writePrivateKeyToFile(priv *ecdsa.PrivateKey, name string) (error) {

	raw, _ := x509.MarshalECPrivateKey(priv)
	cooked := pem.EncodeToMemory(
		&pem.Block{
			Type:  "ECDSA PRIVATE KEY",
			Bytes: raw,
		})
	err := ioutil.WriteFile(fileRoot + name + ".priv", cooked, 0644)
	if err != nil {
		return errors.New("failed save private key for user: " + name + " reasion: " + err.Error())
	}

	raw, _ = x509.MarshalPKIXPublicKey(&priv.PublicKey)
	cooked = pem.EncodeToMemory(
		&pem.Block{
			Type:  "ECDSA PUBLIC KEY",
			Bytes: raw,
		})
	err = ioutil.WriteFile(fileRoot + name + ".pub", cooked, 0644)
	if err != nil {
		return errors.New("failed save public key for user: " + name + " reasion: " + err.Error())
	}
	return nil
}


func getUser(memberService service.MemberServices, id string, passwd []byte, registar string, registarKey *ecdsa.PrivateKey) (*ecdsa.PrivateKey, error) {
	var signPrivateKey *ecdsa.PrivateKey = nil
	if _, err := os.Stat(fileRoot + id + ".priv"); os.IsNotExist(err) {
		var userToken = passwd

		if passwd == nil {
			token, error := memberService.Register(id, 1, "institution_a", "00001", registar, registarKey)
			if error != nil {
				return nil, errors.New("error for register " + error.Error())
			} else {
				logger.Debug("register user token: %s", token.Tok)
				userToken = token.Tok
			}
		}
		signPrivateKey, _ = primitives.NewECDSAKey()
		error := memberService.Enroll(id, userToken, signPrivateKey)
		if error != nil {
			logger.Errorf("error for error %s, %s", id, error)
		} else {
			logger.Debugf("enroll %s Success", id)
			if error = writePrivateKeyToFile(signPrivateKey, id); error != nil {
				return nil, errors.New("failed write to file:" + error.Error())
			}
		}
	} else {
		cooked, err := ioutil.ReadFile(fileRoot + id + ".priv")
		if err != nil {
			return nil, errors.New("failed read admin's private key file.")
		}
		block, _ := pem.Decode(cooked)
		signPrivateKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.New("failed parse admin's private key file.")
		}
	}
	return signPrivateKey, nil
}

func testSign()  {

	address := "MIICUjCCAfigAwIBAgIRANdGtfftKk8zjuMSYyhmkHcwCgYIKoZIzj0EAwMwMTELMAkGA1UEBhMCVVMxFDASBgNVBAoTC0h5cGVybGVkZ2VyMQwwCgYDVQQDEwN0Y2EwHhcNMTYwOTE4MDEzMTMzWhcNMTYxMjE3MDEzMTMzWjBFMQswCQYDVQQGEwJVUzEUMBIGA1UEChMLSHlwZXJsZWRnZXIxIDAeBgNVBAMTF1RyYW5zYWN0aW9uIENlcnRpZmljYXRlMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEdCPMjixmIpazzuYMw7XD8OqT8OJVErSVYs+Rf/f5J8U1twNrVuq6w6GHGH7FDMXB1k5hbgEG3bTWZbMPWdLITaOB3DCB2TAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADANBgNVHQ4EBgQEAQIDBDAPBgNVHSMECDAGgAQBAgMEME0GBioDBAUGBwEB/wRAZyHUrheAIoOQ6V3SkfzQDKTQNbEUgqXKR+1tJeZglecn+rbKGX3+Pbd6ZwhB/d25V+KutJY/PSJeNlofLOfJbzBKBgYqAwQFBggEQM9abXI71r4GIzzbMVn8C8MkIYxFXbhhqY2m6qFiRXJ6bWrSEtKftyNF5Mau4l9L0tKYtzY6hrZSStfskinJfV0wCgYIKoZIzj0EAwMDSAAwRQIhAOOMI91mhLM1j8QShFmOUZsCUa2spO+NQDWx4bifZuGAAiB8hyhBVRlrjc4iQa/QJvKXb9dmynnAVuCgcBAcpWdkeQ=="
	raw, _ := base64.StdEncoding.DecodeString(address)
	priString := "MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgzPIRgXGJnu+Uq+TwkrRTAyHZIQxr5GYibyi5GADBfX6gCgYIKoZIzj0DAQehRANCAASoyGI8ISWUpEyH1hwBPhxFqrUUmRfn6+SWYavmqtB05GRhH38IHpuWKWxCYATmMwUNZK+h/N9AkxNCsdENvNNC"
	rawKey, _ := base64.StdEncoding.DecodeString(priString)
	privateKey,_ := x509.ParsePKCS8PrivateKey(rawKey)
	sign, _ := primitives.ECDSASign(privateKey, raw)
	logger.Debugf("sign sign: [%s]", base64.StdEncoding.EncodeToString(sign))

}

func main() {
	primitives.InitSecurityLevel("SHA3", 256)

	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Now set the configuration file.
	viper.SetConfigName("core") // Name of config file (without extension)
	viper.AddConfigPath("./")    // Path to look for the config file in
	// Path to look for the config file in based on GOPATH
	gopath := os.Getenv("GOPATH")
	for _, p := range filepath.SplitList(gopath) {
		peerpath := filepath.Join(p, "src/github.com/mqshen/BonusLedger/app")
		viper.AddConfigPath(peerpath)
	}

	error := viper.ReadInConfig() // Find and read the config file
	if error != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error when reading config file: %s\n", error))
	}


	logger.Debug(viper.GetString("chaincode.golang.Dockerfile"))

	memberService, _ := service.InitMemberServices("101.251.195.187:50051")

	admin := "admin"

	adminSignPrivateKey, _ := getUser(memberService, admin, []byte("Xurw3yU9zI0l"), "", nil)

	//userName := "test_user3"
	//ownName := "test_user4"
	//
	//userSignPrivateKey, error := getUser(memberService, userName, []byte("vWdLCE00vJy0"), "admin", adminSignPrivateKey)
	//ownUserSignPrivateKey, error := getUser(memberService, ownName, []byte("4nXSrfoYGFCP"), "admin", adminSignPrivateKey)

	//创建用户
	userName := "deploy_for_bonus"
	//ownName := "owner_for_bonus"
	//targetName := "target_for_bonus"

	//bonusName := "test"

	userSignPrivateKey, error := getUser(memberService, userName, nil, "admin", adminSignPrivateKey)
	//ownUserSignPrivateKey, error := getUser(memberService, ownName, nil, "admin", adminSignPrivateKey)
	//targetUserSignPrivateKey, error := getUser(memberService, targetName, nil, "admin", adminSignPrivateKey)


	if error != nil {
		logger.Errorf("error for register user: %s", error)
	} else {
		logger.Debug("enroll users Success")
		peerService, _ := service.InitPeerServices("101.251.195.187:50505")

		assignCert, error := memberService.GetTcert(userName, userSignPrivateKey)

		//ownCert, error := memberService.GetTcert(ownName, ownUserSignPrivateKey)
		//targetCert, error := memberService.GetTcert(targetName, targetUserSignPrivateKey)

		//test := "ce51206d9a7eb8379fbc84d8eee87f51863897972172254c4efb9f4184774e24"
		//chaincodeId := &test

		chaincodeId, error := peerService.Deploy(assignCert, testChaincodePath)
		if error != nil {
			logger.Errorf("failed deploy transaction: %s", error.Error())
		}
		logger.Debug("Wait for 2 minutes for chaincode to be started")
		time.Sleep(2 * time.Minute)
		logger.Debugf("deploy chain code id: [%s]", chaincodeId)

		/*
		ownerAddress := base64.StdEncoding.EncodeToString(ownCert.TCert.GetCertificate().Raw)

		logger.Debugf("cert: [%s]", ownerAddress)
		temp := ownCert.TCert.GetSk()
		raw, _ := x509.MarshalECPrivateKey(temp)
		testPriv := base64.StdEncoding.EncodeToString(raw)
		logger.Debugf("private [%s]", testPriv)

		testRaw := ownCert.TCert.GetCertificate().Raw
		testSign, _ := ownCert.TCert.Sign(testRaw)
		logger.Debugf("sign: [%x]", testSign)


		logger.Debugf("assign asset for [% x]", ownCert.TCert.GetCertificate().Raw)

		id := util.GenerateUUID()

		response, error := peerService.IssueAsset(*chaincodeId, id, assignCert, ownerAddress, bonusName, 20000)
		if error != nil {
			logger.Errorf("failed ivoke transaction: %s", error.Error())
		} else {
			logger.Debug("Success issue asset")
		}
		time.Sleep(10 * time.Second)
		response, error = peerService.QueryAsset(*chaincodeId, ownCert, "organization", bonusName)
		if error != nil {
			logger.Errorf("failed query bonus: %s", error.Error())
		}
		logger.Infof("query resoonse: %s", response)

		id = util.GenerateUUID()
		error = peerService.AssignAsset(*chaincodeId, id, ownCert, targetCert, bonusName, 20170901, 200, "test asset")
		if error != nil {
			logger.Errorf("failed ivoke transaction: %s", error.Error())
		} else {
			logger.Debug("Success assign asset")
		}

		logger.Debug("Wait for 10 seconds for transfer complete")
		time.Sleep(10 * time.Second)

		response, error = peerService.QueryAsset(*chaincodeId, targetCert, "user", bonusName)
		if error != nil {
			logger.Errorf("failed query bonus: %s", error.Error())
		}
		logger.Infof("target query resoonse: %s", response)


		id = util.GenerateUUID()
		ownAddress := base64.StdEncoding.EncodeToString(ownCert.TCert.GetCertificate().Raw)
		response, error = peerService.TransferAsset(*chaincodeId, id, targetCert, ownAddress, bonusName, 100)
		if error != nil {
			logger.Errorf("failed ivoke transfet bonus: %s", error.Error())
		}
		logger.Infof("transfer response: %s", response)

		time.Sleep(10 * time.Second)
		response, error = peerService.QueryAsset(*chaincodeId, ownCert, "user", bonusName)
		if error != nil {
			logger.Errorf("failed query bonus: %s", error.Error())
		}
		logger.Infof("own query response: %s", response)

		response, error = peerService.QueryAsset(*chaincodeId, targetCert, "user", bonusName)
		if error != nil {
			logger.Errorf("failed query target bonus: %s", error.Error())
		}
		logger.Infof("target query target resoonse: %s", response)
		*/
	}
}
