// Copyright (c) 2013-2022 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"github.com/thoughtcore/rosetta-thought/thoughtd/chaincfg/chainhash"
	"github.com/thoughtcore/rosetta-thought/thoughtd/thtec"
	"github.com/thoughtcore/rosetta-thought/thoughtd/thtec/ecdsa"
)

// signatureVerifier is an abstract interface that allows the op code execution
// to abstract over the _type_ of signature validation being executed. At this
// point in Bitcoin's history, there're four possible sig validation contexts:
// pre-segwit, segwit v0, segwit v1 (taproot key spend validation), and the
// base tapscript verification.
type signatureVerifier interface {
	// Verify returns true if the signature verifier context deems the
	// signature to be valid for the given context.
	Verify() bool
}

// baseSigVerifier is used to verify signatures for the _base_ system, meaning
// ECDSA signatures encoded in DER or BER encoding.
type baseSigVerifier struct {
	vm *Engine

	pubKey *thtec.PublicKey

	sig *ecdsa.Signature

	fullSigBytes []byte

	sigBytes []byte
	pkBytes  []byte

	subScript []byte

	hashType SigHashType
}

// parseBaseSigAndPubkey attempts to parse a signature and public key according
// to the base consensus rules, which expect an 33-byte public key and DER or
// BER encoded signature.
func parseBaseSigAndPubkey(pkBytes, fullSigBytes []byte,
	vm *Engine) (*thtec.PublicKey, *ecdsa.Signature, SigHashType, error) {

	strictEncoding := vm.hasFlag(ScriptVerifyStrictEncoding) ||
		vm.hasFlag(ScriptVerifyDERSignatures)

	// Trim off hashtype from the signature string and check if the
	// signature and pubkey conform to the strict encoding requirements
	// depending on the flags.
	//
	// NOTE: When the strict encoding flags are set, any errors in the
	// signature or public encoding here result in an immediate script error
	// (and thus no result bool is pushed to the data stack).  This differs
	// from the logic below where any errors in parsing the signature is
	// treated as the signature failure resulting in false being pushed to
	// the data stack.  This is required because the more general script
	// validation consensus rules do not have the new strict encoding
	// requirements enabled by the flags.
	hashType := SigHashType(fullSigBytes[len(fullSigBytes)-1])
	sigBytes := fullSigBytes[:len(fullSigBytes)-1]
	if err := vm.checkHashTypeEncoding(hashType); err != nil {
		return nil, nil, 0, err
	}
	if err := vm.checkSignatureEncoding(sigBytes); err != nil {
		return nil, nil, 0, err
	}
	if err := vm.checkPubKeyEncoding(pkBytes); err != nil {
		return nil, nil, 0, err
	}

	// First, parse the public key, which we expect to be in the proper
	// encoding.
	pubKey, err := thtec.ParsePubKey(pkBytes)
	if err != nil {
		return nil, nil, 0, err
	}

	// Next, parse the signature which should be in DER or BER depending on
	// the active script flags.
	var signature *ecdsa.Signature
	if strictEncoding {
		signature, err = ecdsa.ParseDERSignature(sigBytes)
	} else {
		signature, err = ecdsa.ParseSignature(sigBytes)
	}
	if err != nil {
		return nil, nil, 0, err
	}

	return pubKey, signature, hashType, nil
}

// newBaseSigVerifier returns a new instance of the base signature verifier. An
// error is returned if the signature, sighash, or public key aren't correctly
// encoded.
func newBaseSigVerifier(pkBytes, fullSigBytes []byte,
	vm *Engine) (*baseSigVerifier, error) {

	pubKey, sig, hashType, err := parseBaseSigAndPubkey(
		pkBytes, fullSigBytes, vm,
	)
	if err != nil {
		return nil, err
	}

	// Get script starting from the most recent OP_CODESEPARATOR.
	subScript := vm.subScript()

	return &baseSigVerifier{
		vm:           vm,
		pubKey:       pubKey,
		pkBytes:      pkBytes,
		sig:          sig,
		sigBytes:     fullSigBytes[:len(fullSigBytes)-1],
		subScript:    subScript,
		hashType:     hashType,
		fullSigBytes: fullSigBytes,
	}, nil
}

// verifySig attempts to verify the signature given the computed sighash. A nil
// error is returned if the signature is valid.
func (b *baseSigVerifier) verifySig(sigHash []byte) bool {
	var valid bool
	if b.vm.sigCache != nil {
		var sigHashBytes chainhash.Hash
		copy(sigHashBytes[:], sigHash[:])

		valid = b.vm.sigCache.Exists(sigHashBytes, b.sigBytes, b.pkBytes)
		if !valid && b.sig.Verify(sigHash, b.pubKey) {
			b.vm.sigCache.Add(sigHashBytes, b.sigBytes, b.pkBytes)
			valid = true
		}
	} else {
		valid = b.sig.Verify(sigHash, b.pubKey)
	}

	return valid
}

// Verify returns true if the signature verifier context deems the signature to
// be valid for the given context.
//
// NOTE: This is part of the baseSigVerifier interface.
func (b *baseSigVerifier) Verify() bool {
	// Remove the signature since there is no way for a signature
	// to sign itself.
	subScript := removeOpcodeByData(b.subScript, b.fullSigBytes)

	sigHash := calcSignatureHash(
		subScript, b.hashType, &b.vm.tx, b.vm.txIdx,
	)

	return b.verifySig(sigHash)
}

// A compile-time assertion to ensure baseSigVerifier implements the
// signatureVerifier interface.
var _ signatureVerifier = (*baseSigVerifier)(nil)
