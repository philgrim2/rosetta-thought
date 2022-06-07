// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"fmt"
)

// NetMagic represents which Thought network a message belongs to.
type NetMagic uint32

// Constants used to indicate the message Thought network.  They can also be
// used to seek to the next message when a stream's state is unknown, but
// this package does not provide that functionality since it's generally a
// better idea to simply disconnect clients that are misbehaving over TCP.
const (
  // MainNet represents the main Thought network.
  MainNet NetMagic = 0x59472ee4
  // TestNet3 represents the test network (version 3).
  TestNet3 NetMagic = 0x2b9939bf
)

// bnStrings is a map of thought networks back to their constant names for
// pretty printing.
var bnStrings = map[NetMagic]string{
	MainNet:  "MainNet",
	TestNet3: "TestNet3",
}

// String returns the NetMagic in human-readable form.
func (n NetMagic) String() string {
	if s, ok := bnStrings[n]; ok {
		return s
	}

	return fmt.Sprintf("Unknown network (%d)", uint32(n))
}

// Params defines a Bitcoin network by its parameters. These parameters may be
// used by Bitcoin applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
    // Name defines a human-readable identifier for the network.
    Name string

    // Net defines the magic bytes used to identify the network.
    Net NetMagic

    // DefaultPort defines the default peer-to-peer port for the network.
    DefaultPort string

    // Address encoding magics
    PubKeyHashAddrID  byte // First byte of a P2PKH address
    ScriptHashAddrID  byte // First byte of a P2SH address
    PrivateKeyID   byte // First byte of a WIF private key

    // BIP32 hierarchical deterministic extended key magics
    HDPrivateKeyID [4]byte
    HDPublicKeyID [4]byte

    // BIP44 coin type used in the hierarchical deterministic path for
    // address generation.
    HDCoinType uint32
}

// MainNetParams defines the network parameters for the main Bitcoin network.
var MainNetParams = Params{
  Name:  "main",
  Net:   MainNet,
  DefaultPort: "10618",

  // Address encoding magics
  PubKeyHashAddrID:  0x07, 
  ScriptHashAddrID:  0x09, 
  PrivateKeyID:   0x7b, 

  // BIP32 hierarchical deterministic extended key magics
  HDPublicKeyID: [4]byte{0xfb, 0xc6, 0xa0, 0x0d}, // starts with xpub
  HDPrivateKeyID: [4]byte{0x5a, 0xeb, 0xd8, 0xc6}, // starts with xprv

  // BIP44 coin type used in the hierarchical deterministic path for
  // address generation.
  HDCoinType: 5,
}


// TestNet3Params defines the network parameters for the test Thought network
// (version 3). Not to be confused with the regression test network, this
// network is sometimes simply called "testnet".
var TestNet3Params = Params{
  Name:  "test",
  Net:   TestNet3,
  DefaultPort: "11618",

  // Address encoding magics
  PubKeyHashAddrID:  0x6d,
  ScriptHashAddrID:  0xc1,
  PrivateKeyID:   0xeb,

  // BIP32 hierarchical deterministic extended key magics
  HDPublicKeyID: [4]byte{0x5d, 0x40, 0x5f, 0x7a}, // starts with tpub
  HDPrivateKeyID: [4]byte{0xb6, 0xf1, 0x3f, 0x50}, // starts with tprv

  // BIP44 coin type used in the hierarchical deterministic path for
  // address generation.
  HDCoinType: 1,
}
