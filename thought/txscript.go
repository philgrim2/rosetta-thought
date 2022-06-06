package thought

import (
	"fmt"
)

// ScriptClass is an enumeration for the list of standard types of script.
type ScriptClass byte

const (
	MaxOpsPerScript       = 201 // Max number of non-push operations.
	MaxPubKeysPerMultiSig = 20  // Multisig can't have more sigs than this.
	MaxScriptElementSize  = 520 // Max bytes pushable to the stack.
	// MaxDataCarrierSize is the maximum number of bytes allowed in pushed
	// data to be considered a nulldata transaction
	MaxDataCarrierSize = 80
)

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


// finalOpcodeData returns the data associated with the final opcode in the
// script.  It will return nil if the script fails to parse.
func finalOpcodeData(script []byte) []byte {
	// Avoid unnecessary work.
	if len(script) == 0 {
		return nil
	}

	var data []byte
	tokenizer := MakeScriptTokenizer(script)
	for tokenizer.Next() {
		data = tokenizer.Data()
	}
	if tokenizer.Err() != nil {
		return nil
	}
	return data
}
// isStrictPubKeyEncoding returns whether or not the passed public key adheres
// to the strict encoding requirements.
func isStrictPubKeyEncoding(pubKey []byte) bool {
	if len(pubKey) == 33 && (pubKey[0] == 0x02 || pubKey[0] == 0x03) {
		// Compressed
		return true
	}
	if len(pubKey) == 65 {
		switch pubKey[0] {
		case 0x04:
			// Uncompressed
			return true

		case 0x06, 0x07:
			// Hybrid
			return true
		}
	}
	return false
}


// pubKeyHashToAddrs is a convenience function to attempt to convert the
// passed hash to a pay-to-pubkey-hash address housed within an address
// slice.  It is used to consolidate common code.
func pubKeyHashToAddrs(hash []byte, params *Params) []Address {
	// Skip the pubkey hash if it's invalid for some reason.
	var addrs []Address
	addr, err := NewAddressPubKeyHash(hash, params)
	if err == nil {
		addrs = append(addrs, addr)
	}
	return addrs
}

// scriptHashToAddrs is a convenience function to attempt to convert the passed
// hash to a pay-to-script-hash address housed within an address slice.  It is
// used to consolidate common code.
func scriptHashToAddrs(hash []byte, params *Params) []Address {
	// Skip the hash if it's invalid for some reason.
	var addrs []Address
	addr, err := NewAddressScriptHashFromHash(hash, params)
	if err == nil {
		addrs = append(addrs, addr)
	}
	return addrs
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


// isPubKeyScript returns whether or not the passed script is either a standard
// pay-to-compressed-secp256k1-pubkey or pay-to-uncompressed-secp256k1-pubkey
// script.
func isPubKeyScript(script []byte) bool {
	return extractPubKey(script) != nil
}

// scriptType returns the type of the script being inspected from the known
// standard types. 
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
// ExtractPkScriptAddrs returns the type of script, addresses and required
// signatures associated with the passed PkScript.  Note that it only works for
// 'standard' transaction script types.  Any data such as public keys which are
// invalid are omitted from the results.
func ExtractPkScriptAddrs(pkScript []byte,
	chainParams *Params) (ScriptClass, []Address, int, error) {

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
		var addrs []Address
		addr, err := NewAddressPubKey(data, chainParams)
		if err == nil {
			addrs = append(addrs, addr)
		}
		return PubKeyTy, addrs, 1, nil
	}

	// Check for multi-signature script.
	details := extractMultisigScriptDetails(pkScript, true)
	if details.valid {
		// Convert the public keys while skipping any that are invalid.
		addrs := make([]Address, 0, len(details.pubKeys))
		for _, pubkey := range details.pubKeys {
			addr, err := NewAddressPubKey(pubkey, chainParams)
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
