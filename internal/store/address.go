package store

import "github.com/ethereum/go-ethereum/common"

// normalizeWalletAddress returns nil for an empty input (so the column is
// stored as NULL) and the EIP-55 checksum form for any non-empty input.
// Callers should validate that the address is well-formed before reaching
// this helper; common.HexToAddress silently coerces invalid hex to the zero
// address.
func normalizeWalletAddress(walletAddress string) any {
	if walletAddress == "" {
		return nil
	}
	return common.HexToAddress(walletAddress).Hex()
}
