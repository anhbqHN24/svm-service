package helper

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// EncodePubKeyToString converts a 32-byte public key to its Base58 string (Address)
func EncodePubKeyToString(pubKey []byte) string {
	// 1. Convert bytes to a large integer
	x := new(big.Int).SetBytes(pubKey)

	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var result []byte

	// 2. Divide by 58 repeatedly to get characters
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// 3. Handle leading zeros (represented as '1' in Base58)
	for _, b := range pubKey {
		if b != 0 {
			break
		}
		result = append(result, base58Alphabet[0])
	}

	// 4. Reverse the result string
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func ConvertToBytes(input interface{}) ([]byte, error) {
	switch v := input.(type) {
	case string:
		return []byte(v), nil
	case float64: // JSON số thường là float64
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, int32(v))
		return buf.Bytes(), err
	case []interface{}: // Mảng JSON (Ví dụ: [1, 2, 255]) -> Program Code
		byteArray := make([]byte, len(v))
		for i, val := range v {
			if num, ok := val.(float64); ok {
				byteArray[i] = byte(num)
			} else {
				return nil, fmt.Errorf("invalid byte array content")
			}
		}
		return byteArray, nil
	default:
		return nil, fmt.Errorf("unsupported data type")
	}
}

func GetBodyInput[T any](r *http.Request) (T, error) {
	var inputType T

	if err := json.NewDecoder(r.Body).Decode(&inputType); err != nil {
		return inputType, err
	}

	return inputType, nil
}
