<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Rosetta" src="/assets/rosetta_header2.png">
  </a>
</p>
<h3 align="center">
   Rosetta Thought
</h3>

<p align="center"><b>
ROSETTA-THOUGHT IS CONSIDERED <a href="https://en.wikipedia.org/wiki/Software_release_life_cycle#Alpha">ALPHA SOFTWARE</a>.
USE AT YOUR OWN RISK.</b><p>
<p align="center">This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).</p>

## Overview

The `rosetta-thought` repository provides an implementation of the Rosetta API for Thought (THT) based on the reference `rosetta-bitcoin` implementation for Bitcoin in Golang. 

[Rosetta](https://www.rosetta-api.org/docs/welcome.html) is an open-source specification and set of tools that makes integrating with blockchains simpler, faster, and more reliable. The Rosetta API is specified in the [OpenAPI 3.0 format](https://www.openapis.org).

Requests and responses can be crafted with auto-generated code using [Swagger Codegen](https://swagger.io/tools/swagger-codegen) or [OpenAPI Generator](https://openapi-generator.tech), are human-readable (easy to debug and understand), and can be used in servers and browsers.

## Features

* Rosetta API implementation (both Data API and Construction API)
* UTXO cache for all accounts (accessible using the Rosetta `/account/balance` API)
* Stateless, offline, curve-based transaction construction from any P2PKH Address
* Automatically prune thoughtd while indexing blocks
* Reduce sync time with concurrent block indexing
* Use [Zstandard compression](https://github.com/facebook/zstd) to reduce the size of data stored on disk without needing to write a manual byte-level encoding

## System Requirements

The `rosetta-thought` implementation has been tested on an [AWS c5.2xlarge instance](https://aws.amazon.com/ec2/instance-types/c5). This instance type has 8 vCPU and 16 GB of RAM.

## Getting Started

1. Adjust your [network settings](#network-settings) to the recommended connections.
2. Install and run Docker as directed in the [Deployment](#deployment) section below.
3. Run the [`Testnet:Online`](#testnetonline) command.

### Network Settings

To increase the load that `rosetta-thought` can handle, we recommend tunning your OS settings to allow for more connections. On a linux-based OS, you can run these commands ([source](http://www.tweaked.io/guide/kernel)):

```text
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
sysctl -w net.ipv4.tcp_max_syn_backlog=10000
sysctl -w net.core.somaxconn=10000
sysctl -p (when done)
```
_We have not tested `rosetta-thought` with `net.ipv4.tcp_tw_recycle` and do not recommend enabling it._

You should also modify your open file settings to `100000`. This can be done on a linux-based OS with the command: `ulimit -n 100000`.

### Memory-Mapped Files

`rosetta-thought` uses [memory-mapped files](https://en.wikipedia.org/wiki/Memory-mapped_file) to persist data in the `indexer`. As a result, you **must** run `rosetta-thought` on a 64-bit architecture (the virtual address space easily exceeds 100s of GBs).

If you receive a kernel OOM, you may need to increase the allocated size of swap space on your OS. There is a great tutorial for how to do this on Linux [here](https://linuxize.com/post/create-a-linux-swap-file/).

## Development

While working on improvements to this repository, we recommend that you use these commands to check your code:

* `make deps` to install dependencies
* `make test` to run tests
* `make lint` to lint the source code
* `make salus` to check for security concerns
* `make build-local` to build a Docker image from the local context
* `make coverage-local` to generate a coverage report

### Deployment

As specified in the [Rosetta API Principles](https://www.rosetta-api.org/docs/automated_deployment.html), all Rosetta implementations must be deployable via Docker and support running via either an [`online` or `offline` mode](https://www.rosetta-api.org/docs/node_deployment.html#multiple-modes).

**YOU MUST [INSTALL DOCKER](https://www.docker.com/get-started) FOR THESE INSTRUCTIONS TO WORK.**

#### Image Installation

Running these commands will create a Docker image called `rosetta-thought:latest`.

##### Installing from Source

After cloning this repository, run:

```text
make build-local
```

#### Run Docker

Running these commands will start a Docker container in [detached mode](https://docs.docker.com/engine/reference/run/#detached--d) with a data directory at `<working directory>/thought-data` and the Rosetta API accessible at port `8080`.

##### Required Arguments

**`MODE`** 
**Type:** `String`
**Options:** `ONLINE`, `OFFLINE`
**Default:** None

`MODE` determines if Rosetta can make outbound connections.

**`NETWORK`**
**Type:** `String`
**Options:** `MAINNET`  or `TESTNET`
**Default:** `TESTNET`

`NETWORK` is the Thought network to launch or communicate with.

**`PORT`**
**Type:** `Integer`
**Options:** `8080`, any compatible port number
**Default:** None

`PORT` is the port to use for Rosetta.

##### Command Examples

You can run these commands from the command line. If you cloned the repository, you can use the `make` commands shown after the examples.

###### **Mainnet:Online**

Uncloned repo:
```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/thought-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -p 8080:8080 -p 10618:10618 rosetta-thought:latest
```
Cloned repo:
```text
make run-mainnet-online
```

###### **Mainnet:Offline**

Uncloned repo:
```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-thought:latest
```
Cloned repo:
```text
make run-mainnet-offline
```

###### **Testnet:Online**

Uncloned repo:
```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/thought-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -p 8080:8080 -p 11618:11618 rosetta-thought:latest
```

Cloned repo: 
```text
make run-testnet-online
```

###### **Testnet:Offline**

Uncloned repo:
```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-thought:latest
```

Cloned repo: 
```text
make run-testnet-offline
```

## Architecture

`rosetta-thought` uses the `syncer`, `storage`, `parser`, and `server` package from [`rosetta-sdk-go`](https://github.com/coinbase/rosetta-sdk-go) instead of a new Thought-specific implementation of packages of similar functionality. Below you can find an overview of how everything fits together:

<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Architecture" src="https://www.rosetta-api.org/img/rosetta_bitcoin_architecture.jpg">
  </a>
</p>

### Concurrent Block Syncing

To speed up indexing, `rosetta-thought` uses concurrent block processing with a "wait free" design (using [the channels function](https://golangdocs.com/channels-in-golang) instead of [the sleep function](https://pkg.go.dev/time#Sleep) to signal which threads are unblocked). This allows `rosetta-thought` to fetch multiple inputs from disk while it waits for inputs that appeared in recently processed blocks to save to disk.

<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Concurrent Block Syncing" src="https://www.rosetta-api.org/img/rosetta_bitcoin_concurrent_block_synching.jpg">
  </a>
</p>

## Test the Implementation with the rosetta-cli Tool

Before validation, it is important to note that `rosetta-thought` can use prefunded accounts to automate testing (It currently doesn't, a prompt will appear to fund an address at check:construction). If choosing to utilize prefunded accounts, new accounts will have to be set for testing by modifying `rosetta-cli-conf/testnet/config.json` for either or both `Mainnet` or `Testnet`. Information on how to obtain necessary information can be found within the [prefunded accounts](#prefunded-accounts) section. Additionally, in order to have test funds returned to the sending account (minus fees), you **MUST** set the environment variable for the receiving address: `export RECIPIENT=\"receiving_address\"`

To validate `rosetta-thought`, [install `rosetta-cli`](https://github.com/coinbase/rosetta-cli#install) and run one of these commands:

* `rosetta-cli check:spec --configuration-file rosetta-cli-conf/testnet/config.json` - This command validates that the API implementation is working under Coinbase specifications.
* `rosetta-cli check:data --configuration-file rosetta-cli-conf/testnet/config.json` - This command validates that the Data API information in the `testnet` network is correct. It also ensures that the implementation does not miss any balance-changing operations.
* `rosetta-cli check:construction --configuration-file rosetta-cli-conf/testnet/config.json` - This command validates the blockchain’s construction, signing, and broadcasting.
* `rosetta-cli check:data --configuration-file rosetta-cli-conf/mainnet/config.json` - This command validates that the Data API information in the `mainnet` network is correct. It also ensures that the implementation does not miss any balance-changing operations.

Read the [How to Test your Rosetta Implementation](https://www.rosetta-api.org/docs/rosetta_test.html) documentation for additional details.

## Prefunded Accounts 

**WARNING** It is never a good idea to save private keys in plain text. Utilizing prefunded accounts will do exactly this. Using prefunded accounts is not neceessary for testing. Performing the following steps can potentially lead to loss of account and their associated balances.

To retrieve private keys:
1. Open ThoughtCore and open the debug console from the tools menu.
2. Type the command: `listunspent 1` to find an address with atleast 1000 THT for the test.
3. Type the command: `dumpprivekey "address"` to retrieve the WIF encoded private key.
4. Base58Decode the private key into bytes.
5. Remove the first byte (network byte) and the last 5 bytes (compression byte + checksum).
6. The remaining string of bytes is the raw private key for the specified address.

## Documentation

You can find the Rosetta API documentation at [rosetta-api.org](https://www.rosetta-api.org/docs/welcome.html). 

Check out the [Getting Started](https://www.rosetta-api.org/docs/getting_started.html) section to start diving into Rosetta. 

Our documentation is divided into the following sections:

* [Product Overview](https://www.rosetta-api.org/docs/welcome.html)
* [Getting Started](https://www.rosetta-api.org/docs/getting_started.html)
* [Rosetta API Spec](https://www.rosetta-api.org/docs/Reference.html)
* [Testing](https://www.rosetta-api.org/docs/rosetta_cli.html)
* [Best Practices](https://www.rosetta-api.org/docs/node_deployment.html)
* [Repositories](https://www.rosetta-api.org/docs/rosetta_specifications.html)

## Related Projects

* [rosetta-sdk-go](https://github.com/coinbase/rosetta-sdk-go) — The `rosetta-sdk-go` SDK provides a collection of packages used for interaction with the Rosetta API specification. 
* [rosetta-specifications](https://github.com/coinbase/rosetta-specifications) — Much of the SDK code is generated from this repository.
* [rosetta-cli](https://github.com/coinbase/rosetta-ecosystem) — Use the `rosetta-cli` tool to test your Rosetta API implementation. The tool also provides the ability to look up block contents and account balances.

### Sample Implementations

You can find community implementations for a variety of blockchains in the [rosetta-ecosystem](https://github.com/coinbase/rosetta-ecosystem) repository, and in the [ecosystem category](https://community.rosetta-api.org/c/ecosystem) of our community site. 

## License
This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).

© 2022 Coinbase
© 2022 Thought Network, LLC
