package parser

import (
	"fmt"
	"strings"
)

// validateMod11NFeKey validates the mod 11 check digit of a NF-e access key.
//
// Algorithm (SEFAZ MOC section 5.4):
//   - Weights 2..9 cyclically applied right to left over the first 43 digits
//   - remainder = sum % 11
//   - remainder 0 or 1 → DV = 0; otherwise DV = 11 - remainder
func validateMod11NFeKey(key string) bool {
	weights := [8]int{2, 3, 4, 5, 6, 7, 8, 9}
	total := 0
	for i, ch := range reverse(key[:43]) {
		total += int(ch-'0') * weights[i%8]
	}
	remainder := total % 11
	var expected int
	if remainder == 0 || remainder == 1 {
		expected = 0
	} else {
		expected = 11 - remainder
	}
	return int(key[43]-'0') == expected
}

// validateCNPJ validates a 14-digit CNPJ using mod 11.
//
// Algorithm (Receita Federal):
//   1st DV: weights 5,4,3,2,9,8,7,6,5,4,3,2 over first 12 digits
//   2nd DV: weights 6,5,4,3,2,9,8,7,6,5,4,3,2 over first 13 digits (including 1st DV)
//   remainder < 2 → DV = 0; otherwise DV = 11 - remainder
func validateCNPJ(cnpj string) error {
	if len(cnpj) != 14 {
		return fmt.Errorf("CNPJ must be 14 digits, got %d", len(cnpj))
	}
	for _, c := range cnpj {
		if c < '0' || c > '9' {
			return fmt.Errorf("CNPJ must contain only digits")
		}
	}

	// All same digits are mathematically valid but not registered
	allSame := true
	for _, c := range cnpj[1:] {
		if c != rune(cnpj[0]) {
			allSame = false
			break
		}
	}
	if allSame {
		return fmt.Errorf("CNPJ with all identical digits is invalid")
	}

	w1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	w2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	dv1 := mod11DV(cnpj[:12], w1)
	if dv1 != int(cnpj[12]-'0') {
		return fmt.Errorf("CNPJ %s: invalid first check digit", cnpj)
	}

	dv2 := mod11DV(cnpj[:13], w2)
	if dv2 != int(cnpj[13]-'0') {
		return fmt.Errorf("CNPJ %s: invalid second check digit", cnpj)
	}

	return nil
}

// mod11DV computes a mod 11 check digit given digits and weights.
func mod11DV(digits string, weights []int) int {
	total := 0
	for i, ch := range digits {
		total += int(ch-'0') * weights[i]
	}
	remainder := total % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// validateAccessKey validates the NF-e access key format and mod 11 check digit.
//
// Key structure (44 digits):
//   [0:2]  cUF     state code
//   [2:6]  AAMM    year and month of issue
//   [6:20] CNPJ    issuer CNPJ (14 digits)
//   [20:22] mod    NF-e model: 55 = NF-e, 65 = NFC-e
//   [22:25] serie  series
//   [25:34] nNF    document number
//   [34:35] tpEmis emission type
//   [35:43] cNF    random numeric code
//   [43]   cDV     mod 11 check digit
func validateAccessKey(key string) error {
	key = strings.TrimPrefix(key, "NFe")

	if len(key) != 44 {
		return fmt.Errorf("must be 44 digits, got %d", len(key))
	}
	for _, c := range key {
		if c < '0' || c > '9' {
			return fmt.Errorf("must contain only digits")
		}
	}

	model := key[20:22]
	if model != "55" && model != "65" {
		return fmt.Errorf("invalid model %q: expected 55 (NF-e) or 65 (NFC-e)", model)
	}

	if !validateMod11NFeKey(key) {
		return fmt.Errorf("invalid mod 11 check digit")
	}

	return nil
}

// validateCNPJConsistency ensures the CNPJ embedded in the access key
// matches the issuer CNPJ extracted from the XML body.
func validateCNPJConsistency(accessKey, issuerCNPJ string) error {
	accessKey = strings.TrimPrefix(accessKey, "NFe")
	if len(accessKey) < 20 {
		return fmt.Errorf("access key too short to extract CNPJ")
	}
	keyCNPJ := accessKey[6:20]
	if keyCNPJ != issuerCNPJ {
		return fmt.Errorf("access key contains CNPJ %s but issuer CNPJ is %s", keyCNPJ, issuerCNPJ)
	}
	return nil
}

// validateProtocol checks that the <protNFe> authorization block is present.
func validateProtocol(hasProtocol bool, accessKey string) error {
	if !hasProtocol {
		return fmt.Errorf("NF-e %s has no protNFe: document may not be SEFAZ-authorized", accessKey)
	}
	return nil
}
