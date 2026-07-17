package parser

import "testing"

// Real, mod-11-valid CNPJs already used as fixtures elsewhere in this repo
// (internal/mock/clients.go and examples/*.xml) -- reusing them here keeps
// the test data consistent with what the rest of the system actually uses,
// rather than inventing new numbers that happen to pass mod 11.
func TestValidateCNPJ(t *testing.T) {
	cases := []struct {
		name    string
		cnpj    string
		wantErr bool
	}{
		{"valid: Empresa Alpha", "10433218000193", false},
		{"valid: Empresa Beta", "19600133000127", false},
		{"valid: Comercio Gama", "89083863000183", false},
		{"valid: external issuer (nfe_compra fixture)", "59407816000187", false},
		{"valid: external recipient (nfe_venda fixture)", "18495931000173", false},
		{"invalid: wrong first check digit", "10433218000103", true},
		{"invalid: wrong second check digit", "10433218000192", true},
		{"invalid: too short", "1043321800019", true},
		{"invalid: too long", "104332180001933", true},
		{"invalid: contains a letter", "1043321800019A", true},
		{"invalid: all identical digits", "11111111111111", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateCNPJ(c.cnpj)
			if c.wantErr && err == nil {
				t.Errorf("validateCNPJ(%q) = nil, want error", c.cnpj)
			}
			if !c.wantErr && err != nil {
				t.Errorf("validateCNPJ(%q) = %v, want nil", c.cnpj, err)
			}
		})
	}
}

func TestValidateAccessKey(t *testing.T) {
	cases := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid: nfe_venda fixture key", "NFe35240119600133000127550010000000021000000020", false},
		{"valid: nfe_compra fixture key", "NFe35240159407816000187550010000000011000000014", false},
		{"valid: nfe_nao_identificada fixture key", "NFe35240103413164000139550010000000031000000030", false},
		{"valid: no NFe prefix", "35240119600133000127550010000000021000000020", false},
		{"invalid: wrong length", "3524011960013300012755001000000002100000002", true},
		{"invalid: contains a letter", "3524011960013300012755001000000002100000002X", true},
		{"invalid: bad model code", "35240119600133000127990010000000021000000020", true},
		{"invalid: corrupted check digit", "35240119600133000127550010000000021000000021", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateAccessKey(c.key)
			if c.wantErr && err == nil {
				t.Errorf("validateAccessKey(%q) = nil, want error", c.key)
			}
			if !c.wantErr && err != nil {
				t.Errorf("validateAccessKey(%q) = %v, want nil", c.key, err)
			}
		})
	}
}

func TestValidateCNPJConsistency(t *testing.T) {
	cases := []struct {
		name       string
		accessKey  string
		issuerCNPJ string
		wantErr    bool
	}{
		{
			"matching CNPJ",
			"NFe35240119600133000127550010000000021000000020",
			"19600133000127",
			false,
		},
		{
			"mismatched CNPJ",
			"NFe35240119600133000127550010000000021000000020",
			"59407816000187",
			true,
		},
		{
			"access key too short",
			"NFe123",
			"19600133000127",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateCNPJConsistency(c.accessKey, c.issuerCNPJ)
			if c.wantErr && err == nil {
				t.Errorf("validateCNPJConsistency(%q, %q) = nil, want error", c.accessKey, c.issuerCNPJ)
			}
			if !c.wantErr && err != nil {
				t.Errorf("validateCNPJConsistency(%q, %q) = %v, want nil", c.accessKey, c.issuerCNPJ, err)
			}
		})
	}
}

func TestValidateProtocol(t *testing.T) {
	if err := validateProtocol(true, "some-key"); err != nil {
		t.Errorf("validateProtocol(true, ...) = %v, want nil", err)
	}
	if err := validateProtocol(false, "some-key"); err == nil {
		t.Errorf("validateProtocol(false, ...) = nil, want error")
	}
}

// reverse is a tiny internal helper, but it's easy to get an off-by-one
// wrong in and validateMod11NFeKey depends on it being exactly right.
func TestReverse(t *testing.T) {
	cases := map[string]string{
		"":      "",
		"a":     "a",
		"ab":    "ba",
		"abc":   "cba",
		"12345": "54321",
	}
	for in, want := range cases {
		if got := reverse(in); got != want {
			t.Errorf("reverse(%q) = %q, want %q", in, got, want)
		}
	}
}
