package helper

import (
	"encoding/json"
	"math/big"
	"net/http"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// EncodePubKeyToString converts a 32-byte public key to a Base58 string
func EncodePubKeyToString(pubKey []byte) string {
	// Convert bytes to a big integer
	x := new(big.Int).SetBytes(pubKey)

	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var result []byte

	// Repeatedly divide by 58 to map values to the alphabet
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// Handle leading zeros (represented as '1' in Base58)
	for _, b := range pubKey {
		if b != 0 {
			break
		}
		result = append(result, base58Alphabet[0])
	}

	// Reverse the slice to get the correct order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// Generic helper to parse JSON body into a struct
func GetBodyInput[T any](r *http.Request) (T, error) {
	var inputType T

	if err := json.NewDecoder(r.Body).Decode(&inputType); err != nil {
		return inputType, err
	}

	return inputType, nil
}
