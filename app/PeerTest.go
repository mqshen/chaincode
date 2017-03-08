package main

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"github.com/hyperledger/fabric/peer/common"
	"github.com/chaincode/services"
	"os"
	"path/filepath"
)

// Constants go here.
const cmdRoot = "core"

//InitConfig initializes viper config
func InitConfig(cmdRoot string) error {
	var alternativeCfgPath = os.Getenv("PEER_CFG_PATH")
	if alternativeCfgPath != "" {
		viper.AddConfigPath(alternativeCfgPath) // Path to look for the config file in
	} else {
		viper.AddConfigPath("./") // Path to look for the config file in
		// Path to look for the config file in based on GOPATH
		gopath := os.Getenv("GOPATH")
		for _, p := range filepath.SplitList(gopath) {
			peerpath := filepath.Join(p, "src/github.com/hyperledger/fabric/peer")
			viper.AddConfigPath(peerpath)
		}
	}

	// Now set the configuration file.
	viper.SetConfigName(cmdRoot) // Name of config file (without extension)

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return fmt.Errorf("Fatal error when reading %s config file: %s\n", cmdRoot, err)
	}

	return nil
}

func main() {
	viper.SetEnvPrefix(cmdRoot)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	err := InitConfig(cmdRoot)
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error when initializing %s config : %s\n", cmdRoot, err))
	}

	err = common.InitConfig(cmdRoot)
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error when initializing %s config : %s\n", cmdRoot, err))
	}

	// Init the MSP
	// TODO: determine the location of this config file
	var mspMgrConfigDir = viper.GetString("peer.mspConfigPath")

	fmt.Printf("root path: %s, get msp file path: %s", os.Getenv("PEER_CFG_PATH"), mspMgrConfigDir)
	err = common.InitCrypto(mspMgrConfigDir)
	if err != nil { // Handle errors reading the config file
		panic(err.Error())
	}

	peer, err := services.NewPeerServices()
	if err != nil {
		fmt.Printf("create peer services filed: %s \n", err)
		return
	}
	code, err := peer.Deploy("bonust", "github.com/chaincode/certificate")
	if err != nil {
		fmt.Printf("deploy chaincode filed: %s \n", err)
		return
	}
	fmt.Printf("deploy chaincode success: %s", code)
}
