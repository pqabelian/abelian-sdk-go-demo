package main

/*
#include <memory.h>
*/
import "C"

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	core "abelian.info/sdk/core"
)

// Define util functions.
func contains[T string | byte | int | int32 | int64](slice []T, v T) bool {
	for _, elem := range slice {
		if elem == v {
			return true
		}
	}

	return false
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

// Define functions for packing output data.
func packToOutData(data []byte, outData []byte) error {
	size := len(data)
	if size > len(outData) {
		return errors.New("outData is too small")
	}
	C.memcpy(unsafe.Pointer(&outData[0]), unsafe.Pointer(&size), 4)
	C.memcpy(unsafe.Pointer(&outData[4]), unsafe.Pointer(&data[0]), C.size_t(size))
	return nil
}

func packToRetData(data []byte) *C.char {
	size := int32(len(data))
	retData := C.malloc(C.size_t(size) + 4)
	C.memcpy(retData, unsafe.Pointer(&size), 4)
	C.memcpy(unsafe.Pointer(uintptr(retData)+4), unsafe.Pointer(&data[0]), C.size_t(size))
	return (*C.char)(retData)
}

// Define variables and data types for handling demo resources.
//
//go:embed resources/*
var demoResources embed.FS

type DemoAccountResource struct {
	Info struct {
		ChainID       int8   `json:"netID"`
		ChainName     string `json:"network"`
		CryptoVersion int    `json:"crypto_version"`
		Mnemonics     string `json:"mnemonic_list"`
	} `json:"info"`

	AccountInfos []struct {
		SerialNo             int    `json:"No."`
		CryptoSeedHex        string `json:"CryptoSeed"`
		CryptoAddressHex     string `json:"CryptoAddress"`
		SpendSecretKeyHex    string `json:"SerializedASksp"`
		SerialNoSecretKeyHex string `json:"SerializedASksn"`
		ViewSecretKeyHex     string `json:"SerializedVSk"`
	} `json:"addresses"`
}

type DemoAccount struct {
	ChainID           int8
	ChainName         string
	CryptoVersion     int
	Mnemonics         string
	SerialNo          int
	CryptoSeed        core.Bytes
	SpendSecretKey    *core.CryptoKey
	SerialNoSecretKey *core.CryptoKey
	ViewSecretKey     *core.CryptoKey
	CryptoAddress     *core.CryptoAddress
	CoinAddress       *core.CoinAddress
	AbelAddress       *core.AbelAddress
	ShortAbelAddress  *core.ShortAbelAddress
	Fingerprint       core.Bytes
}

// Define data types and methods for demo.
type DemoSet struct{}

var ds *DemoSet = &DemoSet{}

func GetAllDemoNames() []string {
	var names []string
	demoSet := reflect.TypeOf(ds)
	for i := 0; i < demoSet.NumMethod(); i++ {
		name := demoSet.Method(i).Name
		if name[:4] == "Demo" {
			names = append(names, name[4:])
		}
	}
	return names
}

func RunDemo(name string, args []string) {
	demoFunc := reflect.ValueOf(ds).MethodByName("Demo" + name)
	if !demoFunc.IsValid() {
		panic(fmt.Errorf("demo %s is not found", name))
	}
	ds.demoInit()
	ds.demoBegin(name, args)
	demoFunc.Call([]reflect.Value{reflect.ValueOf(args)})
	ds.demoEnd(name)
}

func (*DemoSet) demoInit() {
	envDir := ds.getDemoEnvPath()
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		// Create the demo env directory.
		err = os.MkdirAll(envDir, 0755)
		ds.demoExitOnError(err)

		// Copy the the resources dir to the demo env directory.
		srcDir := "resources"
		dstDir := fmt.Sprintf("%s/.%s", envDir, srcDir)
		err = os.Mkdir(dstDir, 0755)
		ds.demoCheck(err)
		entries, err := demoResources.ReadDir(srcDir)
		ds.demoCheck(err)
		for _, entry := range entries {
			srcPath := fmt.Sprintf("%s/%s", srcDir, entry.Name())
			srcFile, err := demoResources.Open(srcPath)
			ds.demoCheck(err)
			defer srcFile.Close()

			dstPath := fmt.Sprintf("%s/%s", dstDir, entry.Name())
			dstFile, err := os.Create(dstPath)
			ds.demoCheck(err)
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			ds.demoCheck(err)
		}
	}

	// Create the demo config file if it does not exist.
	configPath := ds.getDemoConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		srcPath := fmt.Sprintf("%s/.%s/config-chain-2.json", envDir, "resources")
		srcFile, err := os.Open(srcPath)
		ds.demoExitOnError(err)
		defer srcFile.Close()

		dstFile, err := os.Create(configPath)
		ds.demoExitOnError(err)
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		ds.demoExitOnError(err)
	}
}

func (*DemoSet) demoBegin(name string, args []string) {
	fmt.Printf("================================================================================\n")
	fmt.Printf("📀 Demo%s begins.\n", name)
	fmt.Printf("chainID: %v, args: %v\n", ds.getDemoChainID(), args)
}

func (*DemoSet) demoEnd(name string) {
	fmt.Printf("\n")
	fmt.Printf("📀 Demo%s ends.\n", name)
	fmt.Printf("================================================================================\n")
}

func (*DemoSet) demoCase(format string, v ...interface{}) {
	fmt.Printf("\n✅ "+format+"\n", v...)
}

func (*DemoSet) demoBadCase(format string, v ...interface{}) {
	fmt.Printf("\n❌ "+format+"\n", v...)
}

func (*DemoSet) demoCheck(err error) {
	if err != nil {
		panic(err)
	}
}

func (ds *DemoSet) demoExitOnError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("\n")
		fmt.Printf("🛑 Demo exits on error.\n")
		fmt.Printf("================================================================================\n")
		os.Exit(1)
	}
}

func (ds *DemoSet) getDemoEnvPath() string {
	envPath := os.Getenv("ABELSDK_DEMO_ENV")
	if envPath == "" {
		envPath = "./demo"
	}
	return envPath
}

func (ds *DemoSet) getDemoFilePath(relPath string) string {
	return ds.getDemoEnvPath() + "/" + relPath
}

func (ds *DemoSet) getDemoConfigPath() string {
	return ds.getDemoFilePath(".config.json")
}

var demoConfig map[string]string

func (ds *DemoSet) getDemoConfig() map[string]string {
	if demoConfig == nil {
		configPath := ds.getDemoConfigPath()
		data, err := os.ReadFile(configPath)
		ds.demoCheck(err)

		err = json.Unmarshal(data, &demoConfig)
		ds.demoCheck(err)
	}

	return demoConfig
}

func (ds *DemoSet) getDemoConfigStringValue(key string) string {
	config := ds.getDemoConfig()
	return config[key]
}

func (ds *DemoSet) getDemoConfigNumericValue(key string) int64 {
	stringValue := ds.getDemoConfigStringValue(key)
	numericValue, err := strconv.ParseInt(stringValue, 10, 64)
	ds.demoCheck(err)
	return numericValue
}

func (ds *DemoSet) getDemoChainID() int8 {
	return int8(ds.getDemoConfigNumericValue("chainID"))
}

func (ds *DemoSet) DemoALOHA(args []string) {
	if len(args) == 0 {
		args = []string{"World"}
	}

	ds.demoCase("Say hi.")
	fmt.Printf("Aloha, %v!\n", strings.Join(args, " "))

	ds.demoCase("Show demo config.")
	fmt.Printf("Demo env path: %s\n", ds.getDemoEnvPath())
	fmt.Printf("Demo config path: %s\n", ds.getDemoConfigPath())
	config := ds.getDemoConfig()

	// Sort by key.
	keys := make([]string, 0, len(config))
	for key := range config {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Print config.
	for _, key := range keys {
		value := config[key]
		fmt.Printf("%s: %s\n", key, value)
	}
}

func (ds *DemoSet) getDemoAbecRPCClient() *core.AbecRPCClient {
	endpoint := ds.getDemoConfigStringValue("abec.rpc.endpoint")
	username := ds.getDemoConfigStringValue("abec.rpc.username")
	password := ds.getDemoConfigStringValue("abec.rpc.password")
	return core.NewAbecRPCClient(endpoint, username, password)
}

func (ds *DemoSet) getDemoAccounts() map[int]*DemoAccount {
	chainID := ds.getDemoChainID()
	resourceFile := fmt.Sprintf("resources/accounts-chain-%d.json", chainID)
	data, err := demoResources.ReadFile(resourceFile)
	ds.demoCheck(err)

	var demoAccountResource DemoAccountResource
	err = json.Unmarshal(data, &demoAccountResource)
	ds.demoCheck(err)

	demoAccounts := make(map[int]*DemoAccount)
	for _, accountInfo := range demoAccountResource.AccountInfos {
		demoAccount := &DemoAccount{
			ChainID:           demoAccountResource.Info.ChainID,
			ChainName:         demoAccountResource.Info.ChainName,
			CryptoVersion:     demoAccountResource.Info.CryptoVersion,
			Mnemonics:         demoAccountResource.Info.Mnemonics,
			SerialNo:          accountInfo.SerialNo,
			CryptoSeed:        core.MakeBytesFromHexString(accountInfo.CryptoSeedHex),
			SpendSecretKey:    &core.CryptoKey{Bytes: core.MakeBytesFromHexString(accountInfo.SpendSecretKeyHex)},
			SerialNoSecretKey: &core.CryptoKey{Bytes: core.MakeBytesFromHexString(accountInfo.SerialNoSecretKeyHex)},
			ViewSecretKey:     &core.CryptoKey{Bytes: core.MakeBytesFromHexString(accountInfo.ViewSecretKeyHex)},
			CryptoAddress:     core.NewCryptoAddress(core.MakeBytesFromHexString(accountInfo.CryptoAddressHex)),
		}
		demoAccount.CoinAddress = demoAccount.CryptoAddress.GetCoinAddress()
		demoAccount.AbelAddress = core.NewAbelAddressFromCryptoAddress(demoAccount.CryptoAddress, demoAccount.ChainID)
		demoAccount.ShortAbelAddress = demoAccount.AbelAddress.GetShortAbelAddress()
		demoAccount.Fingerprint = demoAccount.ShortAbelAddress.Fingerprint()

		// Re-generate crypto keys and address from crypto seed and check if they match the resource data.
		ckaa, err := core.GenerateCryptoKeysAndAddress(demoAccount.CryptoSeed)
		ds.demoCheck(err)

		// Check spend secret key.
		if !bytes.Equal(ckaa.SpendSecretKey.Bytes, demoAccount.SpendSecretKey.Bytes) {
			ds.demoExitOnError(fmt.Errorf("crypto seed and spend secret key mismatch"))
		}

		// Check serial No. secret key.
		if !bytes.Equal(ckaa.SerialNoSecretKey.Bytes, demoAccount.SerialNoSecretKey.Bytes) {
			ds.demoExitOnError(fmt.Errorf("crypto seed and serial No. secret key mismatch"))
		}

		// Check view secret key. Omit the last 32 bytes due to a known inconsistency issue.
		vskA := ckaa.ViewSecretKey.Bytes.Slice()[0 : ckaa.ViewSecretKey.Len()-32]
		vskB := demoAccount.ViewSecretKey.Bytes.Slice()[0 : demoAccount.ViewSecretKey.Len()-32]
		if !bytes.Equal(vskA, vskB) {
			ds.demoExitOnError(fmt.Errorf("crypto seed and view secret key mismatch"))
		}

		// Check crypto address.
		if !bytes.Equal(ckaa.CryptoAddress.Data(), demoAccount.CryptoAddress.Data()) {
			ds.demoExitOnError(fmt.Errorf("crypto seed and crypto address mismatch"))
		}

		// All checks passed.
		demoAccounts[demoAccount.SerialNo] = demoAccount
	}

	return demoAccounts
}
