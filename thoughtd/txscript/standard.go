// Copyright (c) 2013-2020 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"fmt"

	"github.com/thoughtnetwork/rosetta-thought/thoughtd/util"
	"github.com/thoughtnetwork/rosetta-thought/thoughtd/chaincfg"
	"github.com/thoughtnetwork/rosetta-thought/thoughtd/wire"
)

const (
	// MaxDataCarrierSize is the maximum number of bytes allowed in pushed
	// data to be considered a nulldata transaction
	MaxDataCarrierSize = 80

	// StandardVerifyFlags are the script flags which are used when
	// executing transaction scripts to enforce additional checks which
	// are required for the script to be considered standard.  These checks
	// help reduce issues related to transaction malleability as well as
	// allow pay-to-script hash transactions.  Note these flags are
	// different than what is required for the consensus rules in that they
	// are more strict.
	//
	// TODO: This definition does not belong here.  It belongs in a policy
	// package.
	StandardVerifyFlags = ScriptBip16 |
		ScriptVerifyDERSignatures |
		ScriptVerifyStrictEncoding |
		ScriptVerifyMinimalData |
		ScriptStrictMultiSig |
		ScriptDiscourageUpgradableNops |
		ScriptVerifyCleanStack |
		ScriptVerifyNullFail |
		ScriptVerifyCheckLockTimeVerify |
		ScriptVerifyCheckSequenceVerify |
		ScriptVerifyLowS |
		ScriptStrictMultiSig |
		ScriptVerifyMinimalIf |
		ScriptVerifyDiscourageOpSuccess |
		ScriptVerifyDiscourageUpgradeablePubkeyType
)

// ScriptClass is an enumeration for the list of standard types of script.
type ScriptClass byte

// Classes of script payment known about in the blockchain.
const (
	NonStandardTy         ScriptClass = iota // None of the recognized forms.
	PubKeyTy                                 // Pay pubkey.
	PubKeyHashTy                             // Pay pubkey hash.
	ScriptHashTy                             // Pay to script hash.
	MultiSigTy                               // Multi signature.
	NullDataTy                               // Empty data-only (provably prunable).
)

// scriptClassToName houses the human-readable strings which describe each
// script class.
var scriptClassToName = []string{
	NonStandardTy:         "nonstandard",
	PubKeyTy:              "pubkey",
	PubKeyHashTy:          "pubkeyhash",
	ScriptHashTy:          "scripthash",
	MultiSigTy:            "multisig",
	NullDataTy:            "nulldata",
}

// String implements the Stringer interface by returning the name of
// the enum script class. If the enum is invalid then "Invalid" will be
// returned.
func (t ScriptClass) String() string {
	if int(t) > len(scriptClassToName) || int(t) < 0 {
		return "Invalid"
	}
	return scriptClassToName[t]
}

// extractCompressedPubKey extracts a compressed public key from the passed
// script if it is a standard pay-to-compressed-secp256k1-pubkey script.  It
// will return nil otherwise.
func extractCompressedPubKey(script []byte) []byte {
	// A pay-to-compressed-pubkey script is of the form:
	//  OP_DATA_33 <33-byte compressed pubkey> OP_CHECKSIG

	// All compressed secp256k1 public keys must start with 0x02 or 0x03.
	if len(script) == 35 &&
		script[34] == OP_CHECKSIG &&
		script[0] == OP_DATA_33 &&
		(script[1] == 0x02 || script[1] == 0x03) {

		return script[1:34]
	}

	return nil
}

// extractUncompressedPubKey extracts an uncompressed public key from the
// passed script if it is a standard pay-to-uncompressed-secp256k1-pubkey
// script.  It will return nil otherwise.
func extractUncompressedPubKey(script []byte) []byte {
	// A pay-to-uncompressed-pubkey script is of the form:
	//   OP_DATA_65 <65-byte uncompressed pubkey> OP_CHECKSIG
	//
	// All non-hybrid uncompressed secp256k1 public keys must start with 0x04.
	// Hybrid uncompressed secp256k1 public keys start with 0x06 or 0x07:
	//   - 0x06 => hybrid format for even Y coords
	//   - 0x07 => hybrid format for odd Y coords
	if len(script) == 67 &&
		script[66] == OP_CHECKSIG &&
		script[0] == OP_DATA_65 &&
		(script[1] == 0x04 || script[1] == 0x06 || script[1] == 0x07) {

		return script[1:66]
	}
	return nil
}

// extractPubKey extracts either compressed or uncompressed public key from the
// passed script if it is a either a standard pay-to-compressed-secp256k1-pubkey
// or pay-to-uncompressed-secp256k1-pubkey script, respectively.  It will return
// nil otherwise.
func extractPubKey(script []byte) []byte {
	if pubKey := extractCompressedPubKey(script); pubKey != nil {
		return pubKey
	}
	return extractUncompressedPubKey(script)
}

// isPubKeyScript returns whether or not the passed script is either a standard
// pay-to-compressed-secp256k1-pubkey or pay-to-uncompressed-secp256k1-pubkey
// script.
func isPubKeyScript(script []byte) bool {
	return extractPubKey(script) != nil
}

// extractPubKeyHash extracts the public key hash from the passed script if it
// is a standard pay-to-pubkey-hash script.  It will return nil otherwise.
func extractPubKeyHash(script []byte) []byte {
	// A pay-to-pubkey-hash script is of the form:
	//  OP_DUP OP_HASH160 <20-byte hash> OP_EQUALVERIFY OP_CHECKSIG
	if len(script) == 25 &&
		script[0] == OP_DUP &&
		script[1] == OP_HASH160 &&
		script[2] == OP_DATA_20 &&
		script[23] == OP_EQUALVERIFY &&
		script[24] == OP_CHECKSIG {

		return script[3:23]
	}

	return nil
}

// isPubKeyHashScript returns whether or not the passed script is a standard
// pay-to-pubkey-hash script.
func isPubKeyHashScript(script []byte) bool {
	return extractPubKeyHash(script) != nil
}

// extractScriptHash extracts the script hash from the passed script if it is a
// standard pay-to-script-hash script.  It will return nil otherwise.
//
// NOTE: This function is only valid for version 0 opcodes.  Since the function
// does not accept a script version, the results are undefined for other script
// versions.
func extractScriptHash(script []byte) []byte {
	// A pay-to-script-hash script is of the form:
	//  OP_HASH160 <20-byte scripthash> OP_EQUAL
	if len(script) == 23 &&
		script[0] == OP_HASH160 &&
		script[1] == OP_DATA_20 &&
		script[22] == OP_EQUAL {

		return script[2:22]
	}

	return nil
}

// isScriptHashScript returns whether or not the passed script is a standard
// pay-to-script-hash script.
func isScriptHashScript(script []byte) bool {
	return extractScriptHash(script) != nil
}

// multiSigDetails houses details extracted from a standard multisig script.
type multiSigDetails struct {
	requiredSigs int
	numPubKeys   int
	pubKeys      [][]byte
	valid        bool
}

// extractMultisigScriptDetails attempts to extract details from the passed
// script if it is a standard multisig script.  The returned details struct will
// have the valid flag set to false otherwise.
//
// The extract pubkeys flag indicates whether or not the pubkeys themselves
// should also be extracted and is provided because extracting them results in
// an allocation that the caller might wish to avoid.  The pubKeys member of
// the returned details struct will be nil when the flag is false.
//
// NOTE: This function is only valid for version 0 scripts.  The returned
// details struct will always be empty and have the valid flag set to false for
// other script versions.
func extractMultisigScriptDetails(script []byte, extractPubKeys bool) multiSigDetails {
	// The only currently supported script version is 0.

	// A multi-signature script is of the form:
	//  NUM_SIGS PUBKEY PUBKEY PUBKEY ... NUM_PUBKEYS OP_CHECKMULTISIG

	// The script can't possibly be a multisig script if it doesn't end with
	// OP_CHECKMULTISIG or have at least two small integer pushes preceding it.
	// Fail fast to avoid more work below.
	if len(script) < 3 || script[len(script)-1] != OP_CHECKMULTISIG {
		return multiSigDetails{}
	}

	// The first opcode must be a small integer specifying the number of
	// signatures required.
	tokenizer := MakeScriptTokenizer(script)
	if !tokenizer.Next() || !isSmallInt(tokenizer.Opcode()) {
		return multiSigDetails{}
	}
	requiredSigs := asSmallInt(tokenizer.Opcode())

	// The next series of opcodes must either push public keys or be a small
	// integer specifying the number of public keys.
	var numPubKeys int
	var pubKeys [][]byte
	if extractPubKeys {
		pubKeys = make([][]byte, 0, MaxPubKeysPerMultiSig)
	}
	for tokenizer.Next() {
		if isSmallInt(tokenizer.Opcode()) {
			break
		}

		data := tokenizer.Data()
		numPubKeys++
		if !isStrictPubKeyEncoding(data) {
			continue
		}
		if extractPubKeys {
			pubKeys = append(pubKeys, data)
		}
	}
	if tokenizer.Done() {
		return multiSigDetails{}
	}

	// The next opcode must be a small integer specifying the number of public
	// keys required.
	op := tokenizer.Opcode()
	if !isSmallInt(op) || asSmallInt(op) != numPubKeys {
		return multiSigDetails{}
	}

	// There must only be a single opcode left unparsed which will be
	// OP_CHECKMULTISIG per the check above.
	if int32(len(tokenizer.Script()))-tokenizer.ByteIndex() != 1 {
		return multiSigDetails{}
	}

	return multiSigDetails{
		requiredSigs: requiredSigs,
		numPubKeys:   numPubKeys,
		pubKeys:      pubKeys,
		valid:        true,
	}
}

// isMultisigScript returns whether or not the passed script is a standard
// multisig script.
//
// NOTE: This function is only valid for version 0 scripts.  It will always
// return false for other script versions.
func isMultisigScript(script []byte) bool {
	// Since this is only checking the form of the script, don't extract the
	// public keys to avoid the allocation.
	details := extractMultisigScriptDetails(script, false)
	return details.valid
}

// IsMultisigScript returns whether or not the passed script is a standard
// multisignature script.
//
// NOTE: This function is only valid for version 0 scripts.  Since the function
// does not accept a script version, the results are undefined for other script
// versions.
//
// The error is DEPRECATED and will be removed in the major version bump.
func IsMultisigScript(script []byte) (bool, error) {
	return isMultisigScript(script), nil
}

// IsMultisigSigScript returns whether or not the passed script appears to be a
// signature script which consists of a pay-to-script-hash multi-signature
// redeem script.  Determining if a signature script is actually a redemption of
// pay-to-script-hash requires the associated public key script which is often
// expensive to obtain.  Therefore, this makes a fast best effort guess that has
// a high probability of being correct by checking if the signature script ends
// with a data push and treating that data push as if it were a p2sh redeem
// script
//
// NOTE: This function is only valid for version 0 scripts.  Since the function
// does not accept a script version, the results are undefined for other script
// versions.
func IsMultisigSigScript(script []byte) bool {

	// The script can't possibly be a multisig signature script if it doesn't
	// end with OP_CHECKMULTISIG in the redeem script or have at least two small
	// integers preceding it, and the redeem script itself must be preceded by
	// at least a data push opcode.  Fail fast to avoid more work below.
	if len(script) < 4 || script[len(script)-1] != OP_CHECKMULTISIG {
		return false
	}

	// Parse through the script to find the last opcode and any data it might
	// push and treat it as a p2sh redeem script even though it might not
	// actually be one.
	possibleRedeemScript := finalOpcodeData(script)
	if possibleRedeemScript == nil {
		return false
	}

	// Finally, return if that possible redeem script is a multisig script.
	return isMultisigScript(possibleRedeemScript)
}

// isNullDataScript returns whether or not the passed script is a standard
// null data script.
//
// NOTE: This function is only valid for version 0 scripts.  It will always
// return false for other script versions.
func isNullDataScript(script []byte) bool {
	// A null script is of the form:
	//  OP_RETURN <optional data>
	//
	// Thus, it can either be a single OP_RETURN or an OP_RETURN followed by a
	// data push up to MaxDataCarrierSize bytes.

	// The script can't possibly be a null data script if it doesn't start
	// with OP_RETURN.  Fail fast to avoid more work below.
	if len(script) < 1 || script[0] != OP_RETURN {
		return false
	}

	// Single OP_RETURN.
	if len(script) == 1 {
		return true
	}

	// OP_RETURN followed by data push up to MaxDataCarrierSize bytes.
	tokenizer := MakeScriptTokenizer(script[1:])
	return tokenizer.Next() && tokenizer.Done() &&
		(isSmallInt(tokenizer.Opcode()) || tokenizer.Opcode() <= OP_PUSHDATA4) &&
		len(tokenizer.Data()) <= MaxDataCarrierSize
}

// scriptType returns the type of the script being inspected from the known
// standard types. The version version should be 0 if the script is segwit v0
// or prior, and 1 for segwit v1 (taproot) scripts.
func typeOfScript(script []byte) ScriptClass {
		switch {
		case isPubKeyScript(script):
			return PubKeyTy
		case isPubKeyHashScript(script):
			return PubKeyHashTy
		case isScriptHashScript(script):
			return ScriptHashTy
		case isMultisigScript(script):
			return MultiSigTy
		case isNullDataScript(script):
			return NullDataTy
		}

	return NonStandardTy
}

// GetScriptClass returns the class of the script passed.
//
// NonStandardTy will be returned when the script does not parse.
func GetScriptClass(script []byte) ScriptClass {
	return typeOfScript(script)
}

// NewScriptClass returns the ScriptClass corresponding to the string name
// provided as argument. ErrUnsupportedScriptType error is returned if the
// name doesn't correspond to any known ScriptClass.
//
// Not to be confused with GetScriptClass.
func NewScriptClass(name string) (*ScriptClass, error) {
	for i, n := range scriptClassToName {
		if n == name {
			value := ScriptClass(i)
			return &value, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedScriptType, name)
}

// expectedInputs returns the number of arguments required by a script.
// If the script is of unknown type such that the number can not be determined
// then -1 is returned. We are an internal function and thus assume that class
// is the real class of pops (and we can thus assume things that were determined
// while finding out the type).
//
// NOTE: This function is only valid for version 0 scripts.  Since the function
// does not accept a script version, the results are undefined for other script
// versions.
func expectedInputs(script []byte, class ScriptClass) int {
	switch class {
	case PubKeyTy:
		return 1

	case PubKeyHashTy:
		return 2

	case ScriptHashTy:
		// Not including script.  That is handled by the caller.
		return 1

	case MultiSigTy:
		// Standard multisig has a push a small number for the number
		// of sigs and number of keys.  Check the first push instruction
		// to see how many arguments are expected. typeOfScript already
		// checked this so we know it'll be a small int.  Also, due to
		// the original bitcoind bug where OP_CHECKMULTISIG pops an
		// additional item from the stack, add an extra expected input
		// for the extra push that is required to compensate.
		return asSmallInt(script[0]) + 1

	case NullDataTy:
		fallthrough
	default:
		return -1
	}
}

// ScriptInfo houses information about a script pair that is determined by
// CalcScriptInfo.
type ScriptInfo struct {
	// PkScriptClass is the class of the public key script and is equivalent
	// to calling GetScriptClass on it.
	PkScriptClass ScriptClass

	// NumInputs is the number of inputs provided by the public key script.
	NumInputs int

	// ExpectedInputs is the number of outputs required by the signature
	// script and any pay-to-script-hash scripts. The number will be -1 if
	// unknown.
	ExpectedInputs int

	// SigOps is the number of signature operations in the script pair.
	SigOps int
}

// CalcScriptInfo returns a structure providing data about the provided script
// pair.  It will error if the pair is in someway invalid such that they can not
// be analysed, i.e. if they do not parse or the pkScript is not a push-only
// script
//
// NOTE: This function is only valid for version 0 scripts.  Since the function
// does not accept a script version, the results are undefined for other script
// versions.
//
// DEPRECATED.  This will be removed in the next major version bump.
func CalcScriptInfo(sigScript, pkScript []byte, witness wire.TxWitness,
	bip16, segwit bool) (*ScriptInfo, error) {

	// Count the number of opcodes in the signature script while also ensuring
	// that successfully parses.  Since there is a check below to ensure the
	// script is push only, this equates to the number of inputs to the public
	// key script.
	var numInputs int
	tokenizer := MakeScriptTokenizer(sigScript)
	for tokenizer.Next() {
		numInputs++
	}
	if err := tokenizer.Err(); err != nil {
		return nil, err
	}

	if err := checkScriptParses(pkScript); err != nil {
		return nil, err
	}

	// Can't have a signature script that doesn't just push data.
	if !IsPushOnlyScript(sigScript) {
		return nil, scriptError(ErrNotPushOnly,
			"signature script is not push only")
	}

	si := new(ScriptInfo)
	si.PkScriptClass = typeOfScript(pkScript)

	si.ExpectedInputs = expectedInputs(pkScript, si.PkScriptClass)

	switch {
	// Count sigops taking into account pay-to-script-hash.
	case si.PkScriptClass == ScriptHashTy && bip16 && !segwit:
		// The redeem script is the final data push of the signature script.
		redeemScript := finalOpcodeData(sigScript)
		reedeemClass := typeOfScript(redeemScript)
		rsInputs := expectedInputs(redeemScript, reedeemClass)
		if rsInputs == -1 {
			si.ExpectedInputs = -1
		} else {
			si.ExpectedInputs += rsInputs
		}
		si.SigOps = countSigOpsV0(redeemScript, true)

		// All entries pushed to stack (or are OP_RESERVED and exec
		// will fail).
		si.NumInputs = numInputs

	default:
		si.SigOps = countSigOpsV0(pkScript, true)

		// All entries pushed to stack (or are OP_RESERVED and exec
		// will fail).
		si.NumInputs = numInputs
	}

	return si, nil
}

// CalcMultiSigStats returns the number of public keys and signatures from
// a multi-signature transaction script.  The passed script MUST already be
// known to be a multi-signature script.
//
// NOTE: This function is only valid for version 0 scripts.  Since the function
// does not accept a script version, the results are undefined for other script
// versions.
func CalcMultiSigStats(script []byte) (int, int, error) {
	// The public keys are not needed here, so pass false to avoid the extra
	// allocation.
	details := extractMultisigScriptDetails(script, false)
	if !details.valid {
		str := fmt.Sprintf("script %x is not a multisig script", script)
		return 0, 0, scriptError(ErrNotMultisigScript, str)
	}

	return details.numPubKeys, details.requiredSigs, nil
}

// payToPubKeyHashScript creates a new script to pay a transaction
// output to a 20-byte pubkey hash. It is expected that the input is a valid
// hash.
func payToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return NewScriptBuilder().AddOp(OP_DUP).AddOp(OP_HASH160).
		AddData(pubKeyHash).AddOp(OP_EQUALVERIFY).AddOp(OP_CHECKSIG).
		Script()
}

// payToScriptHashScript creates a new script to pay a transaction output to a
// script hash. It is expected that the input is a valid hash.
func payToScriptHashScript(scriptHash []byte) ([]byte, error) {
	return NewScriptBuilder().AddOp(OP_HASH160).AddData(scriptHash).
		AddOp(OP_EQUAL).Script()
}

// payToPubkeyScript creates a new script to pay a transaction output to a
// public key. It is expected that the input is a valid pubkey.
func payToPubKeyScript(serializedPubKey []byte) ([]byte, error) {
	return NewScriptBuilder().AddData(serializedPubKey).
		AddOp(OP_CHECKSIG).Script()
}

// PayToAddrScript creates a new script to pay a transaction output to a the
// specified address.
func PayToAddrScript(addr util.Address) ([]byte, error) {
	const nilAddrErrStr = "unable to generate payment script for nil address"

	switch addr := addr.(type) {
	case *util.AddressPubKeyHash:
		if addr == nil {
			return nil, scriptError(ErrUnsupportedAddress,
				nilAddrErrStr)
		}
		return payToPubKeyHashScript(addr.ScriptAddress())

	case *util.AddressScriptHash:
		if addr == nil {
			return nil, scriptError(ErrUnsupportedAddress,
				nilAddrErrStr)
		}
		return payToScriptHashScript(addr.ScriptAddress())

	case *util.AddressPubKey:
		if addr == nil {
			return nil, scriptError(ErrUnsupportedAddress,
				nilAddrErrStr)
		}
		return payToPubKeyScript(addr.ScriptAddress())

	}

	str := fmt.Sprintf("unable to generate payment script for unsupported "+
		"address type %T", addr)
	return nil, scriptError(ErrUnsupportedAddress, str)
}

// NullDataScript creates a provably-prunable script containing OP_RETURN
// followed by the passed data.  An Error with the error code ErrTooMuchNullData
// will be returned if the length of the passed data exceeds MaxDataCarrierSize.
func NullDataScript(data []byte) ([]byte, error) {
	if len(data) > MaxDataCarrierSize {
		str := fmt.Sprintf("data size %d is larger than max "+
			"allowed size %d", len(data), MaxDataCarrierSize)
		return nil, scriptError(ErrTooMuchNullData, str)
	}

	return NewScriptBuilder().AddOp(OP_RETURN).AddData(data).Script()
}

// MultiSigScript returns a valid script for a multisignature redemption where
// nrequired of the keys in pubkeys are required to have signed the transaction
// for success.  An Error with the error code ErrTooManyRequiredSigs will be
// returned if nrequired is larger than the number of keys provided.
func MultiSigScript(pubkeys []*util.AddressPubKey, nrequired int) ([]byte, error) {
	if len(pubkeys) < nrequired {
		str := fmt.Sprintf("unable to generate multisig script with "+
			"%d required signatures when there are only %d public "+
			"keys available", nrequired, len(pubkeys))
		return nil, scriptError(ErrTooManyRequiredSigs, str)
	}

	builder := NewScriptBuilder().AddInt64(int64(nrequired))
	for _, key := range pubkeys {
		builder.AddData(key.ScriptAddress())
	}
	builder.AddInt64(int64(len(pubkeys)))
	builder.AddOp(OP_CHECKMULTISIG)

	return builder.Script()
}

// PushedData returns an array of byte slices containing any pushed data found
// in the passed script.  This includes OP_0, but not OP_1 - OP_16.
func PushedData(script []byte) ([][]byte, error) {
	var data [][]byte
	tokenizer := MakeScriptTokenizer(script)
	for tokenizer.Next() {
		if tokenizer.Data() != nil {
			data = append(data, tokenizer.Data())
		} else if tokenizer.Opcode() == OP_0 {
			data = append(data, nil)
		}
	}
	if err := tokenizer.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// pubKeyHashToAddrs is a convenience function to attempt to convert the
// passed hash to a pay-to-pubkey-hash address housed within an address
// slice.  It is used to consolidate common code.
func pubKeyHashToAddrs(hash []byte, params *chaincfg.Params) []util.Address {
	// Skip the pubkey hash if it's invalid for some reason.
	var addrs []util.Address
	addr, err := util.NewAddressPubKeyHash(hash, params)
	if err == nil {
		addrs = append(addrs, addr)
	}
	return addrs
}

// scriptHashToAddrs is a convenience function to attempt to convert the passed
// hash to a pay-to-script-hash address housed within an address slice.  It is
// used to consolidate common code.
func scriptHashToAddrs(hash []byte, params *chaincfg.Params) []util.Address {
	// Skip the hash if it's invalid for some reason.
	var addrs []util.Address
	addr, err := util.NewAddressScriptHashFromHash(hash, params)
	if err == nil {
		addrs = append(addrs, addr)
	}
	return addrs
}

// ExtractPkScriptAddrs returns the type of script, addresses and required
// signatures associated with the passed PkScript.  Note that it only works for
// 'standard' transaction script types.  Any data such as public keys which are
// invalid are omitted from the results.
func ExtractPkScriptAddrs(pkScript []byte,
	chainParams *chaincfg.Params) (ScriptClass, []util.Address, int, error) {

	// Check for pay-to-pubkey-hash script.
	if hash := extractPubKeyHash(pkScript); hash != nil {
		return PubKeyHashTy, pubKeyHashToAddrs(hash, chainParams), 1, nil
	}

	// Check for pay-to-script-hash.
	if hash := extractScriptHash(pkScript); hash != nil {
		return ScriptHashTy, scriptHashToAddrs(hash, chainParams), 1, nil
	}

	// Check for pay-to-pubkey script.
	if data := extractPubKey(pkScript); data != nil {
		var addrs []util.Address
		addr, err := util.NewAddressPubKey(data, chainParams)
		if err == nil {
			addrs = append(addrs, addr)
		}
		return PubKeyTy, addrs, 1, nil
	}

	// Check for multi-signature script.
	details := extractMultisigScriptDetails(pkScript, true)
	if details.valid {
		// Convert the public keys while skipping any that are invalid.
		addrs := make([]util.Address, 0, len(details.pubKeys))
		for _, pubkey := range details.pubKeys {
			addr, err := util.NewAddressPubKey(pubkey, chainParams)
			if err == nil {
				addrs = append(addrs, addr)
			}
		}
		return MultiSigTy, addrs, details.requiredSigs, nil
	}

	// Check for null data script.
	if isNullDataScript(pkScript) {
		// Null data transactions have no addresses or required signatures.
		return NullDataTy, nil, 0, nil
	}

	// If none of the above passed, then the address must be non-standard.
	return NonStandardTy, nil, 0, nil
}

// AtomicSwapDataPushes houses the data pushes found in atomic swap contracts.
type AtomicSwapDataPushes struct {
	RecipientHash160 [20]byte
	RefundHash160    [20]byte
	SecretHash       [32]byte
	SecretSize       int64
	LockTime         int64
}

// ExtractAtomicSwapDataPushes returns the data pushes from an atomic swap
// contract.  If the script is not an atomic swap contract,
// ExtractAtomicSwapDataPushes returns (nil, nil).  Non-nil errors are returned
// for unparsable scripts.
//
// NOTE: Atomic swaps are not considered standard script types by the dcrd
// mempool policy and should be used with P2SH.  The atomic swap format is also
// expected to change to use a more secure hash function in the future.
//
// This function is only defined in the txscript package due to API limitations
// which prevent callers using txscript to parse nonstandard scripts.
//
// DEPRECATED.  This will be removed in the next major version bump.  The error
// should also likely be removed if the code is reimplemented by any callers
// since any errors result in a nil result anyway.
func ExtractAtomicSwapDataPushes(version uint16, pkScript []byte) (*AtomicSwapDataPushes, error) {
	// An atomic swap is of the form:
	//  IF
	//   SIZE <secret size> EQUALVERIFY SHA256 <32-byte secret> EQUALVERIFY DUP
	//   HASH160 <20-byte recipient hash>
	//  ELSE
	//   <locktime> CHECKLOCKTIMEVERIFY DROP DUP HASH160 <20-byte refund hash>
	//  ENDIF
	//  EQUALVERIFY CHECKSIG
	type templateMatch struct {
		expectCanonicalInt bool
		maxIntBytes        int
		opcode             byte
		extractedInt       int64
		extractedData      []byte
	}
	var template = [20]templateMatch{
		{opcode: OP_IF},
		{opcode: OP_SIZE},
		{expectCanonicalInt: true, maxIntBytes: maxScriptNumLen},
		{opcode: OP_EQUALVERIFY},
		{opcode: OP_SHA256},
		{opcode: OP_DATA_32},
		{opcode: OP_EQUALVERIFY},
		{opcode: OP_DUP},
		{opcode: OP_HASH160},
		{opcode: OP_DATA_20},
		{opcode: OP_ELSE},
		{expectCanonicalInt: true, maxIntBytes: cltvMaxScriptNumLen},
		{opcode: OP_CHECKLOCKTIMEVERIFY},
		{opcode: OP_DROP},
		{opcode: OP_DUP},
		{opcode: OP_HASH160},
		{opcode: OP_DATA_20},
		{opcode: OP_ENDIF},
		{opcode: OP_EQUALVERIFY},
		{opcode: OP_CHECKSIG},
	}

	var templateOffset int
	tokenizer := MakeScriptTokenizer(pkScript)
	for tokenizer.Next() {
		// Not an atomic swap script if it has more opcodes than expected in the
		// template.
		if templateOffset >= len(template) {
			return nil, nil
		}

		op := tokenizer.Opcode()
		data := tokenizer.Data()
		tplEntry := &template[templateOffset]
		if tplEntry.expectCanonicalInt {
			switch {
			case data != nil:
				val, err := makeScriptNum(data, true, tplEntry.maxIntBytes)
				if err != nil {
					return nil, err
				}
				tplEntry.extractedInt = int64(val)

			case isSmallInt(op):
				tplEntry.extractedInt = int64(asSmallInt(op))

			// Not an atomic swap script if the opcode does not push an int.
			default:
				return nil, nil
			}
		} else {
			if op != tplEntry.opcode {
				return nil, nil
			}

			tplEntry.extractedData = data
		}

		templateOffset++
	}
	if err := tokenizer.Err(); err != nil {
		return nil, err
	}
	if !tokenizer.Done() || templateOffset != len(template) {
		return nil, nil
	}

	// At this point, the script appears to be an atomic swap, so populate and
	// return the extacted data.
	pushes := AtomicSwapDataPushes{
		SecretSize: template[2].extractedInt,
		LockTime:   template[11].extractedInt,
	}
	copy(pushes.SecretHash[:], template[5].extractedData)
	copy(pushes.RecipientHash160[:], template[9].extractedData)
	copy(pushes.RefundHash160[:], template[16].extractedData)
	return &pushes, nil
}
