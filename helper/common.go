package helper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/near/borsh-go"
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

func SerializePrimitive(value any) ([]byte, error) {
	var tag uint8
	var body []byte
	var err error

	switch v := value.(type) {
	case int64:
		tag = SERIALIZE_TAG_INT
		body, err = borsh.Serialize(v)
	case int: // Hỗ trợ thêm kiểu int mặc định của Go
		tag = SERIALIZE_TAG_INT
		body, err = borsh.Serialize(int64(v))
	case string:
		tag = SERIALIZE_TAG_STRING
		body, err = borsh.Serialize(v)
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}

	if err != nil {
		return nil, err
	}

	// merge tag and body
	result := append([]byte{tag}, body...)
	return result, nil
}

func DeserializePrimitive(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	// Read the tag
	tag := data[0]

	// Notice: data[1:] is the actual serialized int bytes
	switch tag {
	case SERIALIZE_TAG_INT:
		var val int64
		err := borsh.Deserialize(&val, data[1:])
		if err != nil {
			return nil, err
		}
		return val, nil

	case SERIALIZE_TAG_STRING:
		var val string
		err := borsh.Deserialize(&val, data[1:])
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, fmt.Errorf("unsupported data type (tag: %d)", tag)
	}
}
