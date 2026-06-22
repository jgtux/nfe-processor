package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
)

var (
	schema     *xsd.Schema
	schemaOnce sync.Once
	schemaErr  error
)

// SchemaDir must be set at startup to the absolute path of the schemas/ directory.
var SchemaDir string

func loadSchema() (*xsd.Schema, error) {
	schemaOnce.Do(func() {
		if SchemaDir == "" {
			cwd, _ := os.Getwd()
			SchemaDir = filepath.Join(cwd, "internal", "parser", "schemas")
		}

		path := filepath.Join(SchemaDir, "procNFe_v4.00.xsd")
		if _, err := os.Stat(path); err != nil {
			schemaErr = fmt.Errorf("XSD schema not found at %s: %w", path, err)
			return
		}

		schema, schemaErr = xsd.ParseFromFile(path)
		if schemaErr != nil {
			schemaErr = fmt.Errorf("parse procNFe_v4.00.xsd: %w", schemaErr)
		}
	})
	return schema, schemaErr
}

// extractNFe extracts the <NFe> element from a <nfeProc> document,
// preserving the <Signature> block which is required by the schema.
// The signature values are structurally valid but not cryptographically
// verified here — that requires XMLDSig which is out of scope.
func extractNFe(data []byte) ([]byte, error) {
	type nfeProc struct {
		NFe struct {
			Inner []byte `xml:",innerxml"`
		} `xml:"NFe"`
	}

	var proc nfeProc
	if err := xml.Unmarshal(data, &proc); err != nil {
		return nil, fmt.Errorf("extract NFe: %w", err)
	}

	nfe := append(
		[]byte(`<NFe xmlns="http://www.portalfiscal.inf.br/nfe" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">`),
		append(proc.NFe.Inner, []byte(`</NFe>`)...)...,
	)
	return nfe, nil
}

// validateXSD validates a NF-e XML against the official procNFe_v4.00.xsd schema.
// Structural validation only — cryptographic signature verification (XMLDSig)
// is out of scope and requires C14N canonicalization.
func validateXSD(data []byte) error {
	s, err := loadSchema()
	if err != nil {
		return err
	}

	doc, err := libxml2.ParseString(string(data))
	if err != nil {
		return fmt.Errorf("parse XML: %w", err)
	}
	defer doc.Free()

	if err := s.Validate(doc); err != nil {
		if verr, ok := err.(xsd.SchemaValidationError); ok {
			var msgs []string
			for _, e := range verr.Errors() {
				msgs = append(msgs, e.Error())
			}
			return fmt.Errorf("XSD validation errors:\n%s", strings.Join(msgs, "\n"))
		}
		return fmt.Errorf("XSD validation: %w", err)
	}

	return nil
}
