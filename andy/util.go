package andy

import "math/big"

func base58Encode(input []byte) string {
	var result []byte
	x := new(big.Int).SetBytes(input)
	radix := big.NewInt(58)
	mod := new(big.Int)

	for x.Sign() > 0 {
		x.DivMod(x, radix, mod)
		result = append([]byte{base58Alphabet[mod.Int64()]}, result...)
	}

	// Leading zero bytes
	for _, b := range input {
		if b == 0x00 {
			result = append([]byte{base58Alphabet[0]}, result...)
		} else {
			break
		}
	}

	return string(result)
}
