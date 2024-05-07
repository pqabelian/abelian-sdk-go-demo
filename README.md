#  A Demo Application of the Abelian Go SDK

[![GitHub Release](https://img.shields.io/badge/Latest%20version-1.0.0-blue.svg)]()
[![Made with Java](https://img.shields.io/badge/Powered%20by-Go-lightblue.svg)](https://www.java.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-orange.svg)](https://opensource.org/licenses/MIT)

This is a demo application of the Abelian Go SDK. It demonstrates how to use the Abelian Go SDK to implement common blockchain operations such as generating addresses, tracking coins, and making transactions.

## 1. Install dependencies

### 1.1. Install Go

Go version 1.11 or higher is required. Please refer to [the official Go installation guide](https://go.dev/doc/install) for details.

### 1.2. Install build tools

For Linux:

```shell
sudo apt install astyle cmake gcc ninja-build  pkg-config libssl-dev python3-pytest python3-pytest-xdist unzip xsltproc doxygen graphviz python3-yaml
```

For macOS:

```shell
brew install cmake ninja openssl@1.1 wget doxygen graphviz astyle pkg-config && pip3 install pytest pytest-xdist pyyaml
```

## 2. Build the demo application

Clone the repository:

```shell
git clone git@github.com:pqabelian/abel-sdk-go-demo.git
cd abel-sdk-go-demo
```

To build the demo application:

```shell
make
```

To clean the build:

```shell
make clean
```

Note that in the building process, the OpenSSL library is linked statically on macOS and dynamically on Linux. Therefore, to run the compiled binaries on another machine, it is required to have the OpenSSL library installed on Linux while it is not required on macOS.

## 3. Run the demo application

The demo application is built to `build/abelsdk-demo`.

Run the demo application without any argument to print the usage:

```shell
./build/abelsdk-demo
```

Output:

```shell
Usage: ./build/abelsdk-demo <DEMO_NAME> [args...]

Available demo names:
  ALOHA
  BasicAbecRPCClient
  BasicAddress
  BasicBytes
  BasicDemoAccounts
  BasicGenerateAccounts
  SDKGenerateAddresses
  SDKGenerateCryptoKeysAndAddress
  SDKGenerateCryptoSeedFromMnemonic
  SDKGenerateRandomMnemonic
  SDKGetBlockOrTx
  SDKGetChainInfo
  SDKGetMempool
  SDKMakeSignedRawTx
  SDKMakeUnsignedRawTx
  SDKSubmitSignedRawTx
  SDKTrackCoins
```

Specify a demo name to run the corresponding demo. For example, run the `ALOHA` demo:

```shell
./build/abelsdk-demo ALOHA
```

Output:

```shell
================================================================================
📀 DemoALOHA begins.
chainID: 2, args: []

✅ Say hi.
Aloha, World!

✅ Show demo config.
Demo env path: ./demo
Demo config path: ./demo/.config.json
SDKMakeSignedRawTx.inputFile: unsigned-raw-tx.json
SDKMakeSignedRawTx.outputFile: signed-raw-tx.json
SDKMakeSignedRawTx.senders: 0
SDKMakeUnsignedRawTx.outputFile: unsigned-raw-tx.json
SDKMakeUnsignedRawTx.receivers: 6,7,8,9
SDKMakeUnsignedRawTx.scanHeightRange: 3612,3615
SDKMakeUnsignedRawTx.senders: 0,1
SDKSubmitSignedRawTx.inputFile: signed-raw-tx.json
SDKTrackCoins.accounts: 0,1,2,3,4,5,6,7,8,9
SDKTrackCoins.scanHeightRange: 5115,5120
SDKTrackCoins.trackHeightRange: 9395,9400
abec.rpc.endpoint: https://testnet-rpc-exchange.abelian.info
abec.rpc.password: M+DxFwon2FYyiLgaJoTZ9qCr6Jc=
abec.rpc.username: KFf5krbZiLyfo5KaIsNb3Fr2QZs=
chainID: 2

📀 DemoALOHA ends.
================================================================================
```

## 4. Miscellaneous

### 4.1. Build for old x64 processors

The default build of liboqs for x64 processors requires some instruction sets that are not available on older processors.

To build any application using liboqs for older x64 processors, we need to slightly change the compiler options of liboqs. Specifically, in `liboqs/.CMake/compiler_opts.cmake`, we need to change the following line

```cmake
set(OQS_OPT_FLAG "-march=native")
```

to

```cmake
set(OQS_OPT_FLAG "-march=x86-64")
```

Then, rebuild the liboqs library and the application.