package parser

import (
	"os"
	"strings"
	"testing"
)

// examples/*.xml are the same real, hand-annotated fixtures the Python
// smoke test (tests/smoke/smoke_test.py) uploads against a live server --
// reused here so unit tests exercise the exact same documents, not
// separately hand-rolled ones that could drift from what SEFAZ actually
// sends. SchemaDir defaults to a path relative to the compiled binary's
// own directory (see cmd/server/main.go), which doesn't hold under `go
// test` -- point it at the schemas/ directory checked into this package.
func init() {
	SchemaDir = "schemas"
}

func mustReadExample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../../examples/" + name)
	if err != nil {
		t.Fatalf("reading example fixture %s: %v", name, err)
	}
	return data
}

func TestParseNFe_Venda(t *testing.T) {
	data := mustReadExample(t, "nfe_venda.xml")

	nfe, err := ParseNFe(data)
	if err != nil {
		t.Fatalf("ParseNFe(nfe_venda.xml) = %v, want success", err)
	}

	if want := "35240119600133000127550010000000021000000020"; nfe.AccessKey != want {
		t.Errorf("AccessKey = %q, want %q", nfe.AccessKey, want)
	}
	if want := "19600133000127"; nfe.IssuerCNPJ != want {
		t.Errorf("IssuerCNPJ = %q, want %q", nfe.IssuerCNPJ, want)
	}
	if want := "18495931000173"; nfe.RecipientCNPJ != want {
		t.Errorf("RecipientCNPJ = %q, want %q", nfe.RecipientCNPJ, want)
	}
	if want := 3200.50; nfe.TotalAmount != want {
		t.Errorf("TotalAmount = %v, want %v", nfe.TotalAmount, want)
	}
}

func TestParseNFe_Compra(t *testing.T) {
	data := mustReadExample(t, "nfe_compra.xml")

	nfe, err := ParseNFe(data)
	if err != nil {
		t.Fatalf("ParseNFe(nfe_compra.xml) = %v, want success", err)
	}

	if want := "59407816000187"; nfe.IssuerCNPJ != want {
		t.Errorf("IssuerCNPJ = %q, want %q", nfe.IssuerCNPJ, want)
	}
	if want := "10433218000193"; nfe.RecipientCNPJ != want {
		t.Errorf("RecipientCNPJ = %q, want %q", nfe.RecipientCNPJ, want)
	}
	if want := 1500.00; nfe.TotalAmount != want {
		t.Errorf("TotalAmount = %v, want %v", nfe.TotalAmount, want)
	}
}

func TestParseNFe_NaoIdentificada(t *testing.T) {
	data := mustReadExample(t, "nfe_nao_identificada.xml")

	nfe, err := ParseNFe(data)
	if err != nil {
		t.Fatalf("ParseNFe(nfe_nao_identificada.xml) = %v, want success", err)
	}

	// This fixture's whole point is that neither party is an internal
	// client -- ParseNFe itself doesn't classify (that's NFeService.classify),
	// but both CNPJs must still parse out correctly for classification to
	// have a chance of running against them later.
	if want := "03413164000139"; nfe.IssuerCNPJ != want {
		t.Errorf("IssuerCNPJ = %q, want %q", nfe.IssuerCNPJ, want)
	}
	if want := "75255341000107"; nfe.RecipientCNPJ != want {
		t.Errorf("RecipientCNPJ = %q, want %q", nfe.RecipientCNPJ, want)
	}
}

func TestParseNFe_RejectsGarbage(t *testing.T) {
	cases := map[string][]byte{
		"empty input":       []byte{},
		"not XML at all":    []byte("hello world"),
		"unrelated XML":     []byte("<root><child>hi</child></root>"),
		"truncated nfeProc": []byte("<nfeProc><NFe>"),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseNFe(data); err == nil {
				t.Errorf("ParseNFe(%s) = nil error, want rejection", name)
			}
		})
	}
}

func TestParseNFe_RejectsTamperedAccessKey(t *testing.T) {
	data := mustReadExample(t, "nfe_venda.xml")
	tampered := []byte(strings.Replace(string(data),
		`Id="NFe35240119600133000127550010000000021000000020"`,
		`Id="NFe35240119600133000127550010000000021000000021"`, // last digit flipped -> bad mod 11
		1,
	))

	if _, err := ParseNFe(tampered); err == nil {
		t.Errorf("ParseNFe with a tampered access key check digit = nil error, want rejection")
	}
}
