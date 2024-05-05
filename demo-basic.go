//go:build demo || test

// +build: demo test

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	core "abelian.info/sdk/core"
)

func (ds *DemoSet) DemoBasicBytes(args []string) {
	ds.demoCase("Create Bytes by MakeBytes(length).")
	bytes := core.MakeBytes(4)
	fmt.Printf("bytes: %v\n", bytes)
	fmt.Printf("bytes.Slice(): %v\n", bytes.Slice())

	ds.demoCase("Create Bytes by NewBytes(data).")
	data := []byte{0, 1, 2, 3}
	bytes = core.AsBytes(data)
	fmt.Printf("data: %v\n", data)
	fmt.Printf("bytes: %v\n", bytes)
	fmt.Printf("bytes.Slice(): %v\n", bytes.Slice())

	ds.demoCase("Modifying data will change Bytes.")
	copy(data, []byte{4, 5, 6, 7})
	fmt.Printf("bytes: %v\n", bytes)
	fmt.Printf("bytes.Slice(): %v\n", bytes.Slice())

	ds.demoCase("Modifying bytes.Slice() will change Bytes.")
	copy(bytes.Slice(), []byte{8, 9, 10, 11})
	fmt.Printf("bytes: %v\n", bytes)
	fmt.Printf("bytes.Slice(): %v\n", bytes.Slice())

	ds.demoCase("Demo random Bytes.")
	bytes = core.MakeRandomBytes(16)
	fmt.Printf("bytes: %v\n", bytes)
	fmt.Printf("bytes.Slice(): %v\n", bytes.Slice())
	fmt.Printf("bytes.Len(): %v\n", bytes.Len())
	fmt.Printf("bytes.HexString(): %v\n", bytes.HexString())
	fmt.Printf("bytes.Base64String(): %v\n", bytes.Base64String())
	fmt.Printf("bytes.Md5(): %v\n", bytes.Md5().HexString())
	fmt.Printf("bytes.Sha256(): %v\n", bytes.Sha256().HexString())

	ds.demoCase("Demo empty Bytes.")
	bytes = core.MakeBytes(0)
	fmt.Printf("bytes: %v\n", bytes)
	fmt.Printf("bytes.Slice(): %v\n", bytes.Slice())
	fmt.Printf("bytes.Len(): %v\n", bytes.Len())
	fmt.Printf("bytes.HexString(): %v\n", bytes.HexString())
	fmt.Printf("bytes.Base64String(): %v\n", bytes.Base64String())
	fmt.Printf("bytes.Md5(): %v\n", bytes.Md5().HexString())
	fmt.Printf("bytes.Sha256(): %v\n", bytes.Sha256().HexString())
}

func (ds *DemoSet) DemoBasicAddress(args []string) {
	ds.demoBadCase("Create a random address with invalid data (nil).")
	address := core.NewAddress(nil, core.ANY_ADDRESS_TYPE, core.MakeRandomBytes(32))
	fmt.Printf("address: %v\n", address)
	fmt.Printf("address.Validate(): %v\n", address.Validate())

	ds.demoBadCase("Create a random address with invalid fingerprint.")
	address = core.NewAddress(core.MakeRandomBytes(32), core.ANY_ADDRESS_TYPE, nil)
	fmt.Printf("address: %v\n", address)
	fmt.Printf("address.Validate(): %v\n", address.Validate())

	ds.demoBadCase("Create a random coin address with invalid data length.")
	coinAddress := core.NewCoinAddress(core.MakeRandomBytes(32))
	fmt.Printf("coinAddress: %v\n", coinAddress)
	fmt.Printf("coinAddress.Validate(): %v\n", coinAddress.Validate())

	ds.demoCase("Create a valid random coin address.")
	coinAddress = core.NewCoinAddress(core.MakeRandomBytes(core.COIN_ADDRESS_LENGTH))
	fmt.Printf("coinAddress: %v\n", coinAddress)

	ds.demoCase("Create a valid random crypto address and derive other addresses.")
	cryptoAddressData := core.MakeRandomBytes(core.CRYPTO_ADDRESS_LENGTH)
	copy(cryptoAddressData, []byte{0x00, 0x00, 0x00, 0x00})
	cryptoAddress := core.NewCryptoAddress(cryptoAddressData)
	fmt.Printf("cryptoAddress: %v\n", cryptoAddress)
	coinAddress = cryptoAddress.GetCoinAddress()
	fmt.Printf("coinAddress: %v\n", coinAddress)
	chainID := int8(3)
	abelAddress := core.NewAbelAddressFromCryptoAddress(cryptoAddress, chainID)
	fmt.Printf("abelAddress: %v\n", abelAddress)
	shortAbelAddress := abelAddress.GetShortAbelAddress()
	fmt.Printf("shortAbelAddress: %v\n", shortAbelAddress)

	ds.demoCase("Check if the fingerprints of the above addresses are identical.")
	fmt.Printf("cryptoAddress.fingerprint: %v\n", cryptoAddress.Fingerprint())
	fmt.Printf("coinAddress.fingerprint: %v\n", coinAddress.Fingerprint())
	fmt.Printf("abelAddress.fingerprint: %v\n", abelAddress.Fingerprint())
	fmt.Printf("shortAbelAddress.fingerprint: %v\n", shortAbelAddress.Fingerprint())
	isIdentical := bytes.Equal(cryptoAddress.Fingerprint(), coinAddress.Fingerprint()) &&
		bytes.Equal(cryptoAddress.Fingerprint(), abelAddress.Fingerprint()) &&
		bytes.Equal(cryptoAddress.Fingerprint(), shortAbelAddress.Fingerprint())
	fmt.Printf("isIdentical: %v\n", isIdentical)
	if !isIdentical {
		panic("fingerprint mismatch")
	}
}

func (ds *DemoSet) DemoBasicGenerateAccounts(args []string) {
	// Parse demo args.
	flag := flag.NewFlagSet("GenerateAccounts", flag.ContinueOnError)
	chainID := flag.Int("chainID", int(ds.getDemoChainID()), "Chain ID (0 ~ 15).")
	count := flag.Int("count", 10, "Number of accounts to generate.")
	ds.demoExitOnError(flag.Parse(args))

	type Account struct {
		SeqNo             int    `json:"No."`
		CryptoSeed        string `json:"CryptoSeed"`
		SpendSecretKey    string `json:"SerializedASksp"`
		SerialNoSecretKey string `json:"SerializedASksn"`
		ViewSecretKey     string `json:"SerializedVSk"`
		CryptoAddress     string `json:"CryptoAddress"`
	}

	type AccountsFileData struct {
		Info struct {
			CryptoVersion int    `json:"crypto_version"`
			Mnemonics     string `json:"mnemonic_list"`
			Network       string `json:"network"`
			NetID         int    `json:"netID"`
		} `json:"info"`
		Accounts []*Account `json:"addresses"`
	}

	ds.demoCase("Generate %d accounts for chain %d.", *count, *chainID)
	accounts := make([]*Account, *count)
	for i := 0; i < *count; i++ {
		accounts[i] = &Account{SeqNo: i}

		cryptoSeed, err := core.GenerateSafeCryptoSeed()
		ds.demoCheck(err)
		accounts[i].CryptoSeed = cryptoSeed.HexString()

		keysAndAddress, err := core.GenerateCryptoKeysAndAddress(cryptoSeed)
		ds.demoCheck(err)
		accounts[i].SpendSecretKey = keysAndAddress.SpendSecretKey.HexString()
		accounts[i].SerialNoSecretKey = keysAndAddress.SerialNoSecretKey.HexString()
		accounts[i].ViewSecretKey = keysAndAddress.ViewSecretKey.HexString()
		accounts[i].CryptoAddress = keysAndAddress.CryptoAddress.HexString()
	}

	outputPath := ds.getDemoFilePath(fmt.Sprintf("accounts-chain-%d.json", *chainID))
	ds.demoCase("Save the accounts to %s.", outputPath)
	// Create data to save.
	accountsFileData := &AccountsFileData{}
	accountsFileData.Info.CryptoVersion = 0
	accountsFileData.Info.Mnemonics = "none"
	if *chainID == 0 {
		accountsFileData.Info.Network = "mainnet"
	} else if *chainID == 3 {
		accountsFileData.Info.Network = "simnet"
	} else {
		accountsFileData.Info.Network = "unknown"
	}
	accountsFileData.Info.NetID = *chainID
	accountsFileData.Accounts = accounts
	// Save data to file.
	accountsFileDataBytes, err := json.MarshalIndent(accountsFileData, "", "  ")
	ds.demoCheck(err)
	err = os.WriteFile(outputPath, accountsFileDataBytes, 0644)
	ds.demoCheck(err)
}

func (ds *DemoSet) DemoBasicDemoAccounts(args []string) {
	chainID := ds.getDemoChainID()
	ds.demoCase("Load all demo accounts for chain %d.", chainID)
	accounts := ds.getDemoAccounts()
	for i := range accounts {
		fmt.Printf("accounts[%d]: seqno=%v, shortAddress=%v\n", i, accounts[i].SerialNo, accounts[i].ShortAbelAddress)
	}

	ds.demoCase("Dump the addresses of demo accounts.")
	outputDir := ds.getDemoFilePath("accounts")
	os.MkdirAll(outputDir, 0755)
	for i := range accounts {
		outputPath := ds.getDemoFilePath(fmt.Sprintf("accounts/chain-%d-account-%d.abeladdress", chainID, i))
		content := fmt.Sprintf("%s\n%s\n", accounts[i].AbelAddress.HexString(), accounts[i].ShortAbelAddress.HexString())
		file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		defer file.Close()
		_, err = file.WriteString(content)
		ds.demoCheck(err)
	}
	fmt.Printf("Dumped %d addresses to %s.\n", len(accounts), outputDir)
}

func (ds *DemoSet) DemoBasicAbecRPCClient(args []string) {
	client := ds.getDemoAbecRPCClient()

	ds.demoCase("Call GetChainInfo().")
	respBytes, chainInfo, err := client.GetChainInfo()
	ds.demoCheck(err)
	fmt.Printf("Response bytes: %s\n", respBytes)
	fmt.Printf("Response value: %+v\n", *chainInfo)

	ds.demoCase("Call GetBlockHash(0).")
	respBytes, blockHash0, err := client.GetBlockHash(0)
	ds.demoCheck(err)
	fmt.Printf("Response bytes: %s\n", respBytes)
	fmt.Printf("Response value: %+v\n", *blockHash0)

	ds.demoCase("Call GetBlockHash(1).")
	respBytes, blockHash1, err := client.GetBlockHash(1)
	ds.demoCheck(err)
	fmt.Printf("Response bytes: %s\n", respBytes)
	fmt.Printf("Response value: %+v\n", *blockHash1)

	ds.demoCase("Call GetBlock(%v).", blockHash1)
	respBytes, block, err := client.GetBlock(*blockHash1)
	ds.demoCheck(err)
	fmt.Printf("TxHashes: %v\n", block.TxHashes)

	ds.demoCase("Call GetRawTx(%s).", block.TxHashes[0])
	respBytes, rawTx, err := client.GetRawTx(block.TxHashes[0])
	ds.demoCheck(err)
	fmt.Printf("Response bytes: %s|| ... omitted ... ||%s\n", respBytes.Slice()[:1024], respBytes.Slice()[respBytes.Len()-1024:])
	fmt.Printf("Response value: %+v\n", *rawTx)
}
