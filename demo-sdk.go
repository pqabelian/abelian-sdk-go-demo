//go:build demo || test

// +build: demo test

package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	core "abelian.info/sdk/core"
)

func (ds *DemoSet) DemoSDKGetChainInfo(args []string) {
	client := ds.getDemoAbecRPCClient()

	ds.demoCase("Get the latest status of the chain.")
	_, chainInfo, err := client.GetChainInfo()
	ds.demoCheck(err)
	fmt.Printf("Result: %+v\n", *chainInfo)
}

func (ds *DemoSet) DemoSDKGetMempool(args []string) {
	client := ds.getDemoAbecRPCClient()

	ds.demoCase("Get the mempool of the chain.")
	_, chainInfo, err := client.GetMempool()
	ds.demoCheck(err)
	fmt.Printf("Result: %+v\n", *chainInfo)
}

func (ds *DemoSet) DemoSDKGetBlockOrTx(args []string) {
	// Parse demo args.
	if len(args) != 1 {
		args = []string{"-h"}
	}
	flag := flag.NewFlagSet("SDKGetBlockOrTx", flag.ContinueOnError)
	flag.Usage = func() {
		fmt.Printf("Usage: %s <height-or-hash>\n", flag.Name())
	}
	ds.demoExitOnError(flag.Parse(args))

	keyword := flag.Arg(0)
	if strings.HasPrefix(keyword, "0x") {
		keyword = keyword[2:]
	}

	isHeight := false
	height, err := strconv.ParseInt(keyword, 10, 64)
	if err == nil {
		isHeight = true
	}

	// Create RPC client.
	client := ds.getDemoAbecRPCClient()

	var hash string
	if isHeight {
		ds.demoCase("Get block hash for height %v.", height)
		_, blockHash, err := client.GetBlockHash(height)
		ds.demoCheck(err)
		hash = *blockHash
		fmt.Printf("Block hash: 0x%v\n", hash)
	} else {
		hash = keyword
	}

	ds.demoCase("Get block with hash 0x%v.", hash)
	var isBlock bool
	_, block, err := client.GetBlock(hash)
	if err == nil {
		isBlock = true
		fmt.Printf("Result: %+v\n", *block)
	} else {
		isBlock = false
		fmt.Printf("Failed to get block: %v\n", err)
	}

	if !isBlock {
		ds.demoCase("Get transaction with hash 0x%v.", hash)
		_, tx, err := client.GetRawTx(hash)
		if err == nil {
			fmt.Printf("Result: %+v\n", *tx)
		} else {
			fmt.Printf("Failed to get transaction: %v\n", err)
		}
	}
}

func (ds *DemoSet) DemoSDKGenerateCryptoKeysAndAddress(args []string) {
	doTask := func(cryptoSeed core.Bytes) {
		fmt.Printf("CryptoSeed: %v\n", cryptoSeed)
		keysAndAddress, err := core.GenerateCryptoKeysAndAddress(cryptoSeed)
		ds.demoCheck(err)
		fmt.Printf("SpendSecretKey: %v\n", keysAndAddress.SpendSecretKey)
		fmt.Printf("SerialNoSecretKey: %v\n", keysAndAddress.SerialNoSecretKey)
		fmt.Printf("ViewSecretKey: %v\n", keysAndAddress.ViewSecretKey)
		fmt.Printf("CryptoAddress: %v\n", keysAndAddress.CryptoAddress)
	}

	ds.demoCase("Generate a set of crypto keys and address.")
	cryptoSeed, err := core.GenerateSafeCryptoSeed()
	doTask(cryptoSeed)

	ds.demoCase("Generate the same set of crypto keys and address from the same crypto seed.")
	doTask(cryptoSeed)

	ds.demoCase("Generate another set of crypto keys and address from a different crypto seed.")
	cryptoSeed, err = core.GenerateSafeCryptoSeed()
	ds.demoCheck(err)
	doTask(cryptoSeed)

	ds.demoBadCase("Generate a set of crypto keys and address from an *UNSAFE* crypto seed.")
	cryptoSeed = core.MakeBytes(2*64 + 4)
	doTask(cryptoSeed)
}

func (ds *DemoSet) DemoSDKGenerateAddresses(args []string) {
	// Parse demo args.
	flag := flag.NewFlagSet("SDKGenerateAddresses", flag.ContinueOnError)
	chainID := flag.Int("chainID", int(ds.getDemoChainID()), "Chain ID (0 ~ 15).")
	ds.demoExitOnError(flag.Parse(args))

	// Define repeatable demo task.
	doTask := func() {
		cryptoSeed, err := core.GenerateSafeCryptoSeed()
		ds.demoCheck(err)
		keysAndAddress, err := core.GenerateCryptoKeysAndAddress(cryptoSeed)
		ds.demoCheck(err)

		cryptoAddress := keysAndAddress.CryptoAddress
		fmt.Printf("CryptoAddress: %v\n", cryptoAddress)

		coinAddress := cryptoAddress.GetCoinAddress()
		fmt.Printf("CoinAddress: %v\n", coinAddress)

		abelAddress := core.NewAbelAddressFromCryptoAddress(&cryptoAddress, int8(*chainID))
		fmt.Printf("AbelAddress: %v\n", abelAddress)

		shortAbelAddress := abelAddress.GetShortAbelAddress()
		fmt.Printf("ShortAbelAddress: %v\n", shortAbelAddress)
	}

	rounds := 3
	for i := 0; i < rounds; i++ {
		ds.demoCase("Generate an abel address and derive other types of addresses (%d/%d).", i+1, rounds)
		doTask()
	}
}

func (ds *DemoSet) DemoSDKTrackCoins(args []string) {
	// Load default args from demo config.
	defaultAccounts := ds.getDemoConfigStringValue("SDKTrackCoins.accounts")
	defaultScanHeightRange := ds.getDemoConfigStringValue("SDKTrackCoins.scanHeightRange")
	defaultTrackHeightRange := ds.getDemoConfigStringValue("SDKTrackCoins.trackHeightRange")

	// Parse demo args.
	flag := flag.NewFlagSet("SDKTrackCoins", flag.ContinueOnError)
	accountsArg := flag.String("accounts", defaultAccounts, "Seqnos of the tracked accounts.")
	scanHeightRangeArg := flag.String("txosHeightRange", defaultScanHeightRange, "Range of block heights to scan coins.")
	trackHeightRangeArg := flag.String("trackHeightRange", defaultTrackHeightRange, "Range of block heights to track coins.")
	ds.demoExitOnError(flag.Parse(args))

	// Create RPC client.
	client := ds.getDemoAbecRPCClient()

	ds.demoCase("Process demo args.")
	accountSeqnos := strings.Split(*accountsArg, ",")
	scanHeightRange := strings.Split(*scanHeightRangeArg, ",")
	trackHeightRange := strings.Split(*trackHeightRangeArg, ",")
	scanHeightBegin := int64(atoi(scanHeightRange[0]))
	scanHeightEnd := int64(atoi(scanHeightRange[1]))
	trackHeightBegin := int64(atoi(trackHeightRange[0]))
	trackHeightEnd := int64(atoi(trackHeightRange[1]))

	fmt.Printf("tracked account seqnos: %v\n", accountSeqnos)
	fmt.Printf("block range to scan coins: [%v, %v]\n", scanHeightBegin, scanHeightEnd)
	fmt.Printf("block range to track coins: [%v, %v]\n", trackHeightBegin, trackHeightEnd)

	ds.demoCase("Load tracked accounts.")
	demoAccounts := ds.getDemoAccounts()
	trackedAccounts := make([]*DemoAccount, len(accountSeqnos))
	for i, seqno := range accountSeqnos {
		trackedAccount := demoAccounts[atoi(seqno)]
		fmt.Printf("tracked account %v (seqno=%v): %v\n", i, seqno, trackedAccount.ShortAbelAddress)
		trackedAccounts[i] = trackedAccount
	}

	ds.demoCase("Scan blocks %d to %d to search coins belong to the tracked accounts.", scanHeightBegin, scanHeightEnd)
	accountCoins := make([][]*core.Coin, len(trackedAccounts))
	for height := scanHeightBegin; height <= scanHeightEnd; height++ {
		_, blockHash, err := client.GetBlockHash(height)
		ds.demoCheck(err)
		_, block, err := client.GetBlock(*blockHash)
		ds.demoCheck(err)
		fmt.Printf("Scanning block %v (hash=%v)...\n", height, *blockHash)
		for _, txHash := range block.TxHashes {
			_, tx, err := client.GetRawTx(txHash)
			ds.demoCheck(err)

			for txIndex, vout := range tx.Vout {
				voutData := core.MakeBytesFromHexString(vout.Script)
				coinAddress, err := core.DecodeCoinAddressFromTxOutData(voutData)
				ds.demoCheck(err)

				for i, trackedAccount := range trackedAccounts {
					if bytes.Equal(coinAddress.Fingerprint(), trackedAccount.Fingerprint) {
						txoValue, err := core.DecodeValueFromTxOutData(voutData, trackedAccount.ViewSecretKey)
						ds.demoCheck(err)
						fmt.Printf("  🔥 Found coin for tracked account %v: %v ABEL\n", i, core.NeutrinoToAbel(txoValue))
						coin := &core.Coin{
							ID:                core.CoinID{TxHash: core.MakeBytesFromHexString(txHash), Index: uint8(txIndex)},
							OwnerShortAddress: trackedAccount.ShortAbelAddress,
							OwnerAddress:      trackedAccount.AbelAddress,
							Value:             txoValue,
							SerialNumber:      nil,
							TxVoutData:        voutData,
							BlockHash:         core.MakeBytesFromHexString(*blockHash),
							BlockHeight:       height,
						}
						accountCoins[i] = append(accountCoins[i], coin)
					}
				}
			}
		}
	}
	fmt.Printf("Coins found:\n")
	totalCoins := 0
	for i, coins := range accountCoins {
		totalCoins += len(coins)
		fmt.Printf("  tracked account %v: %d coins\n", i, len(coins))
	}
	fmt.Printf("💰 Total found: %d coins\n", totalCoins)

	ds.demoCase("Prepare data for calculating coin serial numbers.")
	// Prepare CoinIDs and SerialNoSecretKeys.
	allCoins := make([]*core.Coin, 0, totalCoins)
	coinIDs := make([]*core.CoinID, 0, totalCoins)
	serialNoSecretKeys := make([]*core.CryptoKey, 0, totalCoins)
	for i, coins := range accountCoins {
		for _, coin := range coins {
			allCoins = append(allCoins, coin)
			coinIDs = append(coinIDs, &coin.ID)
			serialNoSecretKeys = append(serialNoSecretKeys, trackedAccounts[i].SerialNoSecretKey)
		}
	}
	// Prepare ring blocks data.
	allRingBlockHeights := make([]int64, 0, len(allCoins)*3)
	for _, coin := range allCoins {
		if !contains(allRingBlockHeights, coin.BlockHeight) {
			allRingBlockHeights = append(allRingBlockHeights, coin.BlockHeight)
			coinRingBlockHeights := core.GetRingBlockHeights(coin.BlockHeight)
			for _, coinRingBlockHeight := range coinRingBlockHeights {
				if !contains(allRingBlockHeights, coinRingBlockHeight) {
					allRingBlockHeights = append(allRingBlockHeights, coinRingBlockHeight)
				}
			}
		}
	}
	fmt.Printf("Ring block heights: %v\n", allRingBlockHeights)
	allRingBlocks := make(map[int64]*core.TxBlockDesc)
	for _, ringBlockHeight := range allRingBlockHeights {
		ringBlockBytes, err := client.GetBlockBytesByHeight(ringBlockHeight)
		ds.demoCheck(err)
		ringBlock := core.NewTxBlockDesc(ringBlockBytes, ringBlockHeight)
		allRingBlocks[ringBlockHeight] = ringBlock
		fmt.Printf("  ring block %d: %v\n", ringBlock.Height, ringBlock.BinData)
	}

	ds.demoCase("Calculate coin serial numbers for all coins found in a batch.")
	// Calculate coin serial numbers.
	coinSerialNumbers, err := core.DecodeCoinSerialNumbers(coinIDs, serialNoSecretKeys, allRingBlocks)
	ds.demoCheck(err)
	// Assign coin serial numbers to coins.
	for i, coin := range allCoins {
		coin.SerialNumber = core.AsBytes(coinSerialNumbers[i])
	}
	// Print result.
	fmt.Printf("Coin serial numbers calculated:\n")
	for i, coin := range allCoins {
		fmt.Printf("  coin %d: %v\n", i, coin.SerialNumber)
	}

	ds.demoCase("Track if the above coins were spent in blocks %d to %d.", trackHeightBegin, trackHeightEnd)
	totalSpentCoins := 0
	for height := trackHeightBegin; height <= trackHeightEnd; height++ {
		_, blockHash, err := client.GetBlockHash(height)
		ds.demoCheck(err)
		_, block, err := client.GetBlock(*blockHash)
		ds.demoCheck(err)
		fmt.Printf("Scanning block %v (hash=%v)...\n", height, *blockHash)
		for _, txHash := range block.TxHashes {
			_, tx, err := client.GetRawTx(txHash)
			ds.demoCheck(err)

			for _, vin := range tx.Vin {
				serialNumber := core.MakeBytesFromHexString(vin.SerialNumber)
				fmt.Printf("  Found coin serial number: %v\n", serialNumber)
				for _, coin := range allCoins {
					if bytes.Equal(coin.SerialNumber, serialNumber) {
						totalSpentCoins += 1
						fmt.Printf("  🔥 Coin %v was spent by tx %v.\n", coin.ID, txHash)
					}
				}
			}
		}
	}
	fmt.Printf("💰 Total spent: %d coins\n", totalSpentCoins)
}

func (ds *DemoSet) DemoSDKMakeUnsignedRawTx(args []string) {
	// Load default args from demo config.
	defaultScanHeightRange := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.scanHeightRange")
	defaultSenders := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.senders")
	defaultReceivers := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.receivers")
	defaultOutputFile := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.outputFile")

	// Parse demo args.
	flag := flag.NewFlagSet("SDKMakeUnsignedRawTx", flag.ContinueOnError)
	scanHeightRangeArg := flag.String("txosHeightRange", defaultScanHeightRange, "Range of block heights to scan coins.")
	sendersArg := flag.String("senders", defaultSenders, "Seqnos of the sender accounts.")
	receiversArg := flag.String("receivers", defaultReceivers, "Seqnos of the receiver accounts.")
	outputFileArg := flag.String("outputFile", defaultOutputFile, "Output file name.")
	ds.demoExitOnError(flag.Parse(args))

	// Create RPC client and get demo accounts.
	client := ds.getDemoAbecRPCClient()
	demoAccounts := ds.getDemoAccounts()

	ds.demoCase("Process demo args.")
	// Check block range.
	scanHeightRange := strings.Split(*scanHeightRangeArg, ",")
	scanHeightBegin := int64(atoi(scanHeightRange[0]))
	scanHeightEnd := int64(atoi(scanHeightRange[1]))
	fmt.Printf("block range to scan coins: [%v, %v]\n", scanHeightBegin, scanHeightEnd)
	if scanHeightBegin > scanHeightEnd || scanHeightBegin < 0 {
		ds.demoExitOnError(fmt.Errorf("invalid block range: [%v, %v]", scanHeightBegin, scanHeightEnd))
	}
	// Load sender accounts.
	fmt.Printf("sender seqnos: %v\n", *sendersArg)
	senderSeqnos := strings.Split(*sendersArg, ",")
	senderAccounts := make([]*DemoAccount, len(senderSeqnos))
	for i, seqno := range senderSeqnos {
		senderAccounts[i] = demoAccounts[atoi(seqno)]
		fmt.Printf("  sender account %v (seqno=%v): %v\n", i, seqno, senderAccounts[i].ShortAbelAddress)
	}
	// Load receiver accounts.
	fmt.Printf("receiver seqnos: %v\n", *receiversArg)
	receiverSeqnos := strings.Split(*receiversArg, ",")
	receiverAccounts := make([]*DemoAccount, len(receiverSeqnos))
	for i, seqno := range receiverSeqnos {
		receiverAccounts[i] = demoAccounts[atoi(seqno)]
		fmt.Printf("  receiver account %v (seqno=%v): %v\n", i, seqno, receiverAccounts[i].ShortAbelAddress)
	}
	// Check output file.
	outputPath := ds.getDemoFilePath(*outputFileArg)
	fmt.Printf("output file: %v\n", outputPath)
	if _, err := os.Stat(outputPath); err == nil {
		ds.demoExitOnError(fmt.Errorf("output file %v already exists", outputPath))
	}

	ds.demoCase("Find all coins in the block range owned by sender accounts.")
	senderTxInDescs := make([][]*core.TxInDesc, len(senderAccounts))
	for height := scanHeightBegin; height <= scanHeightEnd; height++ {
		_, blockHash, err := client.GetBlockHash(height)
		ds.demoCheck(err)
		_, block, err := client.GetBlock(*blockHash)
		ds.demoCheck(err)
		fmt.Printf("Searching block %v (hash=%v)...\n", height, *blockHash)
		for _, txHash := range block.TxHashes {
			_, tx, err := client.GetRawTx(txHash)
			ds.demoCheck(err)

			for txIndex, vout := range tx.Vout {
				voutData := core.MakeBytesFromHexString(vout.Script)
				coinAddress, err := core.DecodeCoinAddressFromTxOutData(voutData)
				ds.demoCheck(err)

				for i, senderAccount := range senderAccounts {
					if bytes.Equal(coinAddress.Fingerprint(), senderAccount.Fingerprint) {
						txoValue, err := core.DecodeValueFromTxOutData(voutData, senderAccount.ViewSecretKey)
						ds.demoCheck(err)
						fmt.Printf("  🔥 Found coin owned by sender %v: %v ABEL\n", i, core.NeutrinoToAbel(txoValue))
						txoDesc := &core.TxInDesc{
							TxOutData:  voutData,
							CoinValue:  txoValue,
							Owner:      senderAccount.ShortAbelAddress,
							Height:     height,
							TxHash:     core.MakeBytesFromHexString(txHash),
							TxOutIndex: uint8(txIndex),
						}
						senderTxInDescs[i] = append(senderTxInDescs[i], txoDesc)
					}
				}
			}
		}
	}
	fmt.Printf("Coins found:\n")
	for i, txInDescs := range senderTxInDescs {
		fmt.Printf("  sender %v: %d coins\n", i, len(txInDescs))
	}

	ds.demoCase("Pick a random coin to spend for each sender account.")
	txInDescsToSpend := make([]*core.TxInDesc, 0, len(senderAccounts))
	rand.Seed(time.Now().UnixNano())
	for _, txInDescs := range senderTxInDescs {
		if len(txInDescs) == 0 {
			continue
		}
		randIndex := rand.Intn(len(txInDescs))
		txInDescsToSpend = append(txInDescsToSpend, txInDescs[randIndex])
	}
	if len(txInDescsToSpend) == 0 {
		ds.demoExitOnError(fmt.Errorf("failed to find any coin to spend"))
	}
	fmt.Printf("TxInDescs:\n")
	for i, txInDescs := range txInDescsToSpend {
		fmt.Printf("  txInDesc[%v]: height: %d, value: %v ABEL, sender: %v, outpoint: (%s,%d)\n",
			i, txInDescs.Height, core.NeutrinoToAbel(txInDescs.CoinValue), txInDescs.Owner, txInDescs.TxHash, txInDescs.TxOutIndex)
	}

	ds.demoCase("Get estimated TxFee.")
	estimatedTxFee := client.GetEstimatedTxFee()
	fmt.Printf("TxFee: %v ABEL\n", core.NeutrinoToAbel(estimatedTxFee))

	ds.demoCase("Calculate TxOutDescs for all receivers.")
	// Calculate the value to spend by deducting TxFee from the total value in Txos.
	totalCoinValue := int64(0)
	for _, txInDesc := range txInDescsToSpend {
		totalCoinValue += txInDesc.CoinValue
	}
	spendableCoinValue := totalCoinValue - estimatedTxFee
	// Transfer all the value in sender Txos to the receivers, distribute the total value evenly.
	coinValuePerReceiver := spendableCoinValue / int64(len(receiverAccounts))
	// Create TxOutDescs for all receivers.
	txOutDescs := make([]*core.TxOutDesc, len(receiverAccounts))
	for i, receiverAccount := range receiverAccounts {
		txOutDescs[i] = &core.TxOutDesc{
			AbelAddress: receiverAccount.AbelAddress,
			CoinValue:   coinValuePerReceiver,
		}
		if i == len(receiverAccounts)-1 {
			// The last receiver gets the remaining value.
			txOutDescs[i].CoinValue += spendableCoinValue % int64(len(receiverAccounts))
		}
	}
	fmt.Printf("TxOutDescs:\n")
	for i, txOutDesc := range txOutDescs {
		fmt.Printf("  txOutDesc[%v]: value: %v ABEL, receiver: %v\n",
			i, core.NeutrinoToAbel(txOutDesc.CoinValue), txOutDesc.AbelAddress.GetShortAbelAddress())
	}

	ds.demoCase("Get ring blocks for all TxInDescs.")
	allRingBlockHeights := make([]int64, 0, len(txInDescsToSpend)*3)
	for _, txInDesc := range txInDescsToSpend {
		ringBlockHeights := core.GetRingBlockHeights(txInDesc.Height)
		for _, ringBlockHeight := range ringBlockHeights {
			if !contains(allRingBlockHeights, ringBlockHeight) {
				allRingBlockHeights = append(allRingBlockHeights, ringBlockHeight)
			}
		}
	}
	fmt.Printf("Ring block heights: %v\n", allRingBlockHeights)
	allRingBlocks := make(map[int64]*core.TxBlockDesc)
	for _, ringBlockHeight := range allRingBlockHeights {
		ringBlockBytes, err := client.GetBlockBytesByHeight(ringBlockHeight)
		ds.demoCheck(err)
		ringBlock := core.NewTxBlockDesc(ringBlockBytes, ringBlockHeight)
		allRingBlocks[ringBlockHeight] = ringBlock
		fmt.Printf("  ring block %d: %v\n", ringBlock.Height, ringBlock.BinData)
	}

	ds.demoCase("Generate an UnsignedRawTx and write it to output file.")
	txDesc := core.NewTxDesc(txInDescsToSpend, txOutDescs, estimatedTxFee, allRingBlocks)
	unsignedRawTx, err := core.GenerateUnsignedRawTx(txDesc)
	ds.demoCheck(err)
	err = os.WriteFile(outputPath, unsignedRawTx.Bytes, 0644)
	ds.demoCheck(err)
	fmt.Printf("UnsignedRawTx written to file: %v\n", outputPath)
}

func (ds *DemoSet) DemoSDKMakeUnsignedRawTxWithMemo(args []string) {
	// Load default args from demo config.
	defaultScanHeightRange := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.scanHeightRange")
	defaultSenders := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.senders")
	defaultReceivers := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.receivers")
	defaultOutputFile := ds.getDemoConfigStringValue("SDKMakeUnsignedRawTx.outputFile")

	// Parse demo args.
	flag := flag.NewFlagSet("SDKMakeUnsignedRawTx", flag.ContinueOnError)
	scanHeightRangeArg := flag.String("txosHeightRange", defaultScanHeightRange, "Range of block heights to scan coins.")
	sendersArg := flag.String("senders", defaultSenders, "Seqnos of the sender accounts.")
	receiversArg := flag.String("receivers", defaultReceivers, "Seqnos of the receiver accounts.")
	outputFileArg := flag.String("outputFile", defaultOutputFile, "Output file name.")
	ds.demoExitOnError(flag.Parse(args))

	// Create RPC client and get demo accounts.
	client := ds.getDemoAbecRPCClient()
	demoAccounts := ds.getDemoAccounts()

	ds.demoCase("Process demo args.")
	// Check block range.
	scanHeightRange := strings.Split(*scanHeightRangeArg, ",")
	scanHeightBegin := int64(atoi(scanHeightRange[0]))
	scanHeightEnd := int64(atoi(scanHeightRange[1]))
	fmt.Printf("block range to scan coins: [%v, %v]\n", scanHeightBegin, scanHeightEnd)
	if scanHeightBegin > scanHeightEnd || scanHeightBegin < 0 {
		ds.demoExitOnError(fmt.Errorf("invalid block range: [%v, %v]", scanHeightBegin, scanHeightEnd))
	}
	// Load sender accounts.
	fmt.Printf("sender seqnos: %v\n", *sendersArg)
	senderSeqnos := strings.Split(*sendersArg, ",")
	senderAccounts := make([]*DemoAccount, len(senderSeqnos))
	for i, seqno := range senderSeqnos {
		senderAccounts[i] = demoAccounts[atoi(seqno)]
		fmt.Printf("  sender account %v (seqno=%v): %v\n", i, seqno, senderAccounts[i].ShortAbelAddress)
	}
	// Load receiver accounts.
	fmt.Printf("receiver seqnos: %v\n", *receiversArg)
	receiverSeqnos := strings.Split(*receiversArg, ",")
	receiverAccounts := make([]*DemoAccount, len(receiverSeqnos))
	for i, seqno := range receiverSeqnos {
		receiverAccounts[i] = demoAccounts[atoi(seqno)]
		fmt.Printf("  receiver account %v (seqno=%v): %v\n", i, seqno, receiverAccounts[i].ShortAbelAddress)
	}
	// Check output file.
	outputPath := ds.getDemoFilePath(*outputFileArg)
	fmt.Printf("output file: %v\n", outputPath)
	if _, err := os.Stat(outputPath); err == nil {
		ds.demoExitOnError(fmt.Errorf("output file %v already exists", outputPath))
	}

	ds.demoCase("Find all coins in the block range owned by sender accounts.")
	senderTxInDescs := make([][]*core.TxInDesc, len(senderAccounts))
	for height := scanHeightBegin; height <= scanHeightEnd; height++ {
		_, blockHash, err := client.GetBlockHash(height)
		ds.demoCheck(err)
		_, block, err := client.GetBlock(*blockHash)
		ds.demoCheck(err)
		fmt.Printf("Searching block %v (hash=%v)...\n", height, *blockHash)
		for _, txHash := range block.TxHashes {
			_, tx, err := client.GetRawTx(txHash)
			ds.demoCheck(err)

			for txIndex, vout := range tx.Vout {
				voutData := core.MakeBytesFromHexString(vout.Script)
				coinAddress, err := core.DecodeCoinAddressFromTxOutData(voutData)
				ds.demoCheck(err)

				for i, senderAccount := range senderAccounts {
					if bytes.Equal(coinAddress.Fingerprint(), senderAccount.Fingerprint) {
						txoValue, err := core.DecodeValueFromTxOutData(voutData, senderAccount.ViewSecretKey)
						ds.demoCheck(err)
						fmt.Printf("  🔥 Found coin owned by sender %v: %v ABEL\n", i, core.NeutrinoToAbel(txoValue))
						txoDesc := &core.TxInDesc{
							TxOutData:  voutData,
							CoinValue:  txoValue,
							Owner:      senderAccount.ShortAbelAddress,
							Height:     height,
							TxHash:     core.MakeBytesFromHexString(txHash),
							TxOutIndex: uint8(txIndex),
						}
						senderTxInDescs[i] = append(senderTxInDescs[i], txoDesc)
					}
				}
			}
		}
	}
	fmt.Printf("Coins found:\n")
	for i, txInDescs := range senderTxInDescs {
		fmt.Printf("  sender %v: %d coins\n", i, len(txInDescs))
	}

	ds.demoCase("Pick a random coin to spend for each sender account.")
	txInDescsToSpend := make([]*core.TxInDesc, 0, len(senderAccounts))
	rand.Seed(time.Now().UnixNano())
	for _, txInDescs := range senderTxInDescs {
		if len(txInDescs) == 0 {
			continue
		}
		randIndex := rand.Intn(len(txInDescs))
		txInDescsToSpend = append(txInDescsToSpend, txInDescs[randIndex])
	}
	if len(txInDescsToSpend) == 0 {
		ds.demoExitOnError(fmt.Errorf("failed to find any coin to spend"))
	}
	fmt.Printf("TxInDescs:\n")
	for i, txInDescs := range txInDescsToSpend {
		fmt.Printf("  txInDesc[%v]: height: %d, value: %v ABEL, sender: %v, outpoint: (%s,%d)\n",
			i, txInDescs.Height, core.NeutrinoToAbel(txInDescs.CoinValue), txInDescs.Owner, txInDescs.TxHash, txInDescs.TxOutIndex)
	}

	ds.demoCase("Get estimated TxFee.")
	estimatedTxFee := client.GetEstimatedTxFee()
	fmt.Printf("TxFee: %v ABEL\n", core.NeutrinoToAbel(estimatedTxFee))

	ds.demoCase("Calculate TxOutDescs for all receivers.")
	// Calculate the value to spend by deducting TxFee from the total value in Txos.
	totalCoinValue := int64(0)
	for _, txInDesc := range txInDescsToSpend {
		totalCoinValue += txInDesc.CoinValue
	}
	spendableCoinValue := totalCoinValue - estimatedTxFee
	// Transfer all the value in sender Txos to the receivers, distribute the total value evenly.
	coinValuePerReceiver := spendableCoinValue / int64(len(receiverAccounts))
	// Create TxOutDescs for all receivers.
	txOutDescs := make([]*core.TxOutDesc, len(receiverAccounts))
	for i, receiverAccount := range receiverAccounts {
		txOutDescs[i] = &core.TxOutDesc{
			AbelAddress: receiverAccount.AbelAddress,
			CoinValue:   coinValuePerReceiver,
		}
		if i == len(receiverAccounts)-1 {
			// The last receiver gets the remaining value.
			txOutDescs[i].CoinValue += spendableCoinValue % int64(len(receiverAccounts))
		}
	}
	fmt.Printf("TxOutDescs:\n")
	for i, txOutDesc := range txOutDescs {
		fmt.Printf("  txOutDesc[%v]: value: %v ABEL, receiver: %v\n",
			i, core.NeutrinoToAbel(txOutDesc.CoinValue), txOutDesc.AbelAddress.GetShortAbelAddress())
	}

	ds.demoCase("Get ring blocks for all TxInDescs.")
	allRingBlockHeights := make([]int64, 0, len(txInDescsToSpend)*3)
	for _, txInDesc := range txInDescsToSpend {
		ringBlockHeights := core.GetRingBlockHeights(txInDesc.Height)
		for _, ringBlockHeight := range ringBlockHeights {
			if !contains(allRingBlockHeights, ringBlockHeight) {
				allRingBlockHeights = append(allRingBlockHeights, ringBlockHeight)
			}
		}
	}
	fmt.Printf("Ring block heights: %v\n", allRingBlockHeights)
	allRingBlocks := make(map[int64]*core.TxBlockDesc)
	for _, ringBlockHeight := range allRingBlockHeights {
		ringBlockBytes, err := client.GetBlockBytesByHeight(ringBlockHeight)
		ds.demoCheck(err)
		ringBlock := core.NewTxBlockDesc(ringBlockBytes, ringBlockHeight)
		allRingBlocks[ringBlockHeight] = ringBlock
		fmt.Printf("  ring block %d: %v\n", ringBlock.Height, ringBlock.BinData)
	}

	ds.demoCase("Generate an UnsignedRawTx and write it to output file.")
	txDesc, err := core.NewTxDescWithOptions(txInDescsToSpend, txOutDescs, estimatedTxFee, allRingBlocks, core.SetMemo([]byte("memo")))
	ds.demoCheck(err)
	unsignedRawTx, err := core.GenerateUnsignedRawTx(txDesc)
	ds.demoCheck(err)
	err = os.WriteFile(outputPath, unsignedRawTx.Bytes, 0644)
	ds.demoCheck(err)
	fmt.Printf("UnsignedRawTx written to file: %v\n", outputPath)
}

func (ds *DemoSet) DemoSDKMakeSignedRawTx(args []string) {
	// Load default args from demo config.
	defaultSenders := ds.getDemoConfigStringValue("SDKMakeSignedRawTx.senders")
	defaultInputFile := ds.getDemoConfigStringValue("SDKMakeSignedRawTx.inputFile")
	defaultOutputFile := ds.getDemoConfigStringValue("SDKMakeSignedRawTx.outputFile")

	// Parse demo args.
	flag := flag.NewFlagSet("SDKMakeSignedRawTx", flag.ContinueOnError)
	sendersArg := flag.String("senders", defaultSenders, "Seqnos of the sender accounts.")
	inputFileArg := flag.String("inputFile", defaultInputFile, "Input file name.")
	outputFileArg := flag.String("outputFile", defaultOutputFile, "Output file name.")
	ds.demoExitOnError(flag.Parse(args))

	ds.demoCase("Process demo args.")
	// Load sender accounts.
	fmt.Printf("sender seqnos: %v\n", *sendersArg)
	demoAccounts := ds.getDemoAccounts()
	senderSeqnos := strings.Split(*sendersArg, ",")
	senderAccounts := make([]*DemoAccount, len(senderSeqnos))
	for i, seqno := range senderSeqnos {
		senderAccounts[i] = demoAccounts[atoi(seqno)]
		fmt.Printf("  sender account %v (seqno=%v): %v\n", i, seqno, senderAccounts[i].ShortAbelAddress)
	}
	// Check input file.
	inputPath := ds.getDemoFilePath(*inputFileArg)
	fmt.Printf("input file: %v\n", inputPath)
	if _, err := os.Stat(inputPath); err != nil {
		ds.demoExitOnError(fmt.Errorf("input file %v does not exist", inputPath))
	}
	// Check output file.
	outputPath := ds.getDemoFilePath(*outputFileArg)
	fmt.Printf("output file: %v\n", outputPath)
	if _, err := os.Stat(outputPath); err == nil {
		ds.demoExitOnError(fmt.Errorf("output file %v already exists", outputPath))
	}

	ds.demoCase("Read UnsignedRawTx from input file.")
	unsignedRawTxData, err := os.ReadFile(inputPath)
	ds.demoCheck(err)
	unsignedRawTx := core.NewUnsignedRawTx(unsignedRawTxData)
	fmt.Printf("UnsignedRawTx: %v\n", unsignedRawTx)

	ds.demoCase("Generate a SignedRawTx and write it to output file.")
	senderKeys := make([]*core.CryptoKeysAndAddress, len(senderAccounts))
	for i, senderAccount := range senderAccounts {
		senderKeys[i] = &core.CryptoKeysAndAddress{
			SpendSecretKey:    *senderAccount.SpendSecretKey,
			SerialNoSecretKey: *senderAccount.SerialNoSecretKey,
			ViewSecretKey:     *senderAccount.ViewSecretKey,
			CryptoAddress:     *senderAccount.CryptoAddress,
		}
	}
	signedRawTx, err := core.GenerateSignedRawTx(unsignedRawTx, senderKeys)
	ds.demoCheck(err)
	fmt.Printf("SignedRawTx: %v\n", signedRawTx.Bytes)
	fmt.Printf("Txid: %v\n", signedRawTx.Txid.HexString())
	err = os.WriteFile(outputPath, signedRawTx.Bytes, 0644)
	ds.demoCheck(err)
	fmt.Printf("SignedRawTx written to file: %v\n", outputPath)
}

func (ds *DemoSet) DemoSDKSubmitSignedRawTx(args []string) {
	// Load default args from demo config.
	defaultInputFile := ds.getDemoConfigStringValue("SDKSubmitSignedRawTx.inputFile")

	// Parse demo args.
	flag := flag.NewFlagSet("SDKSubmitSignedRawTx", flag.ContinueOnError)
	inputFileArg := flag.String("inputFile", defaultInputFile, "Input file name.")

	// Create RPC client.
	client := ds.getDemoAbecRPCClient()

	ds.demoCase("Process demo args.")
	// Check input file.
	inputPath := ds.getDemoFilePath(*inputFileArg)
	fmt.Printf("input file: %v\n", inputPath)
	if _, err := os.Stat(inputPath); err != nil {
		ds.demoExitOnError(fmt.Errorf("input file %v does not exist", inputPath))
	}

	ds.demoCase("Read UnsignedRawTx from input file.")
	signedRawTx, err := os.ReadFile(inputPath)
	ds.demoCheck(err)
	fmt.Printf("Length of SignedRawTx: %d\n", len(signedRawTx))

	ds.demoCase("Get the latest status of the chain.")
	_, chainInfo, err := client.GetChainInfo()
	ds.demoCheck(err)
	fmt.Printf("Result: %+v\n", *chainInfo)

	ds.demoCase("Submit SignedRawTx.")
	hexString := hex.EncodeToString(signedRawTx)
	_, txHash, err := client.SendRawTx(hexString)
	ds.demoCheck(err)
	fmt.Printf("Returned tx hash: %v\n", *txHash)
}

func (ds *DemoSet) DemoSDKGenerateRandomMnemonic(args []string) {
	mnemonic, err := core.GenerateRandomMnemonic()
	ds.demoCheck(err)
	fmt.Printf("Mnemonic: %v\n", strings.Join(mnemonic, ","))
}

func (ds *DemoSet) DemoSDKGenerateCryptoSeedFromMnemonic(args []string) {
	mnemonic := strings.Split("boost,forum,win,black,access,come,thunder,apple,lake,trip,school,romance,face,appear,rifle,dilemma,unknown,shield,juice,aspect,genuine,bottom,push,exclude", ",")
	sequenceNumber := uint64(0)

	cryptoSeed, err := core.GenerateCryptoSeedFromMnemonic(mnemonic, sequenceNumber)
	ds.demoCheck(err)
	fmt.Printf("Mnemonic: %v\n", strings.Join(mnemonic, ","))
	fmt.Printf("sequenceNumber: %v\n", sequenceNumber)
	fmt.Printf("CryptoSeed: %x\n", cryptoSeed)
}
