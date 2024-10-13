package normalize

import (
	"encoding/hex"
	"regexp"
	"strings"

	tonaddress "github.com/openweb3-io/crosschain/blockchain/ton/address"
	tontx "github.com/openweb3-io/crosschain/blockchain/ton/tx"
	xc_types "github.com/openweb3-io/crosschain/types"
)

// Given an address like coin::Coin<0x11AAbbCCdd::coin::NAME>,
// we only want to normalize the 0x11AAbbCCdd part, and remove the coin::Coin::<> part.
func NormalizeMoveAddress(address string) string {
	// find a hexadecimal string
	r, err := regexp.Compile("0[xX][0-9a-fA-F]+")
	if err != nil {
		panic(err)
	}
	if strings.Contains(address, "coin::Coin<") {
		address = strings.Replace(address, "coin::Coin<", "", 1)
		address = strings.Replace(address, ">", "", 1)
	}
	if !strings.HasPrefix(address, "0x") {
		rHex, err := regexp.Compile("^[0-9a-fA-F]+")
		if err != nil {
			panic(err)
		}
		match := rHex.FindString(address)
		if len(match) > 0 {
			address = "0x" + address
		}
	}

	match := r.FindString(address)
	if match != "" {
		// replace the hexadeciaml portion of the string with lowercase
		matchLower := strings.ToLower(match)
		address = strings.Replace(address, match, matchLower, 1)

		return address
	} else {
		// check if it's valid hex
		_, err := hex.DecodeString(address)
		if err == nil {
			address = strings.ToLower(address)
		}

		return address
	}
}

// Zero pad hex string with prefix.
// Target Length should be the length of the hex string (without prefix), not the represented bytes.
func zeroPadHex(prefix string, addr string, targetLength int) string {
	addr = strings.TrimPrefix(addr, prefix)
	for len(addr) < targetLength {
		addr = "0" + addr
	}
	return prefix + strings.ToLower(addr)
}

type NormalizeOptions struct {
	NoPrefix bool
	ZeroPad  bool
	// is this a transaction hash instead of an address?
	// TransactionHash bool
}

// NormalizeAddressString normalizes an address or hash
// If possible (if it's hex), it will be lowercased.
// You may specify if you want to remove or ensure the common prefix (if there is one).
func Normalize(address string, nativeAsset xc_types.NativeAsset) string {
	if address == "" {
		return ""
	}
	if address == string(nativeAsset) {
		// In some cases e.g. ("ETH", "ETH") is passed, and we should not normalize anything.
		return address
	}
	if nativeAsset == "" && strings.HasPrefix(address, "0x") {
		nativeAsset = xc_types.ETH
	}
	options := &NormalizeOptions{
		NoPrefix: false,
		ZeroPad:  false,
	}

	address = strings.TrimSpace(address)
	switch blockchain := xc_types.NativeAsset(nativeAsset).Blockchain(); blockchain {
	case xc_types.BlockchainEVM, xc_types.BlockchainEVMLegacy:
		prefix := "0x"
		if nativeAsset == xc_types.XDC {
			// XDC chain uses a different prefix
			address = strings.TrimPrefix(address, prefix)
			prefix = "xdc"
		}
		if options.ZeroPad {
			address = zeroPadHex(prefix, address, 40)
		}
		address = strings.TrimPrefix(address, prefix)
		if !options.NoPrefix {
			address = prefix + address
		}
		address = strings.ToLower(address)

	case xc_types.BlockchainBtcCash, xc_types.BlockchainBtc:
		// remove bitcoincash: prefix
		if strings.Contains(address, ":") {
			address = strings.Split(address, ":")[1]
		}
	/*
		case xc_types.BlockchainAptos, xc_types.BlockchainSui:
			address = NormalizeMoveAddress(address)
			if options.NoPrefix {
				address = strings.TrimPrefix(address, "0x")
			}
	*/
	case xc_types.BlockchainCosmos:
		// nothing to do, bech32

	case xc_types.BlockchainSolana:
		// nothing to do, base58
	case xc_types.BlockchainTron:
		// TODO
	case xc_types.BlockchainTon:
		// convert the "0:1234" format to base64 if needed
		address, _ = tonaddress.Normalize(address)
	default:
	}
	return address
}

// Normalize a transaction hash
func TransactionHash(hash string, nativeAsset xc_types.NativeAsset) string {
	if hash == "" {
		return ""
	}

	hash = strings.TrimSpace(hash)

	switch blockchain := xc_types.NativeAsset(nativeAsset).Blockchain(); blockchain {
	case xc_types.BlockchainEVM, xc_types.BlockchainEVMLegacy:
		prefix := "0x"
		if nativeAsset == xc_types.XDC {
			// XDC chain uses a different prefix
			hash = strings.TrimPrefix(hash, prefix)
			prefix = "xdc"
		}
		hash = zeroPadHex(prefix, hash, 64)

		// TODO should we include 0x prefix?
		hash = strings.TrimPrefix(hash, prefix)
		hash = prefix + hash
		hash = strings.ToLower(hash)

	case xc_types.BlockchainBtcCash, xc_types.BlockchainBtc:
		hash = strings.TrimPrefix(hash, "0x")
		hash = strings.ToLower(hash)
	/*
		case xc_types.BlockchainAptos, xc_types.BlockchainSui:
			if blockchain == xc_types.BlockchainSui {
				// Sui transaction hashes are not hex
				return hash
			}
			hash = NormalizeMoveAddress(hash)
	*/
	case xc_types.BlockchainCosmos:
		// cosmos hash tx do not prefix 0x, so we always remove.
		hash = strings.TrimPrefix(hash, "0x")
		hash = zeroPadHex("", hash, 64)
		hash = strings.ToLower(hash)

	case xc_types.BlockchainSolana:
		// nothing to do, base58
	case xc_types.BlockchainTron:
		// TODO
	case xc_types.BlockchainTon:
		return tontx.Normalize(hash)
	default:
	}
	return hash
}

// deprecated, use Normalize
func NormalizeAddressString(address string, nativeAsset xc_types.NativeAsset) string {
	return Normalize(address, nativeAsset)
}

func AddressEqual(address1 string, address2 string, nativeAsset xc_types.NativeAsset) bool {
	addr1 := NormalizeAddressString(address1, nativeAsset)
	addr2 := NormalizeAddressString(address2, nativeAsset)
	return addr1 == addr2
}
