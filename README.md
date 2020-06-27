DREP Chain
====


[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)


## What is drep?

drep is a full node implementation of Drep written in Go (golang).

It acts as a fully-validating chain daemon for the Drep cryptocurrency.  drep
maintains the entire past transactional ledger of Drep and allows relaying of
transactions to other Drep nodes around the world.

This software is currently under active development.  It is extremely stable and
has been in production use since February 2018.


## What is a full node?

The term 'full node' is short for 'fully-validating node' and refers to software
that fully validates all transactions and blocks, as opposed to trusting a 3rd
party.  In addition to validating transactions and blocks, nearly all full nodes
also participate in relaying transactions and blocks to other full nodes around
the world, thus forming the peer-to-peer network that is the backbone of the
Drep cryptocurrency.

The full node distinction is important, since full nodes are not the only type
of software participating in the Drep peer network. For instance, there are
'lightweight nodes' which rely on full nodes to serve the transactions, blocks,
and cryptographic proofs they require to function, as well as relay their
transactions to the rest of the global network.

## Why run drep?

As described in the previous section, the Drep cryptocurrency relies on having
a peer-to-peer network of nodes that fully validate all transactions and blocks
and then relay them to other full nodes.

Running a full node with drep contributes to the overall security of the
network, increases the available paths for transactions and blocks to relay.

In terms of individual benefits, since drep fully validates every block and
transaction, it provides the highest security and privacy possible when used in
conjunction with a wallet that also supports directly connecting to it in full
validation mode, such as [wallet](https://drep.top/appdrep1.2.0.apk).

## Minimum Recommended Specifications (drep only)

* 128 GB disk space (as of September 2018, increases over time)
* 4 GB memory (RAM)
* ~150MB/day download, ~1.5GB/day upload
  * Plus one-time initial download of the entire block chain
* Windows 7/8.x/10 (server preferred), macOS, Linux
* High uptime

## Getting Started

So, you've decided to help the network by running a full node.  Great!  Running
drep is simple.  All you need to do is install drep on a machine that is
connected to the internet and meets the minimum recommended specifications, and
launch it.

Also, make sure your firewall is configured to allow inbound connections to port
10086 and 10085.

<a name="Installation" />

## Installing and updating

### Binaries (Windows/Linux/macOS)

According to your needs, you can refer to the following links to complete the build process

[Main-net](http://docs.drep.org/advanced/using-mainnet/) 

[Test-net](http://docs.drep.org/advanced/using-testnet/) 

[Private-net](http://docs.drep.org/advanced/using-privatenet/) 


### Build from source code(all platforms)

Building or updating from source requires the following build dependencies

[Build from source code](http://docs.drep.org/advanced/build-sourcecode/) 


## Documentation

The documentation for drep is a work-in-progress.  It is located in the
[docs](http://docs.drep.org) folder.

## License

drep is licensed under the [copyfree](http://copyfree.org) ISC License.
