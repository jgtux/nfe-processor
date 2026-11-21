package parser

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lestrrat-go/helium"
	"github.com/lestrrat-go/helium/xsd"
)

var (
	schema     *xsd.Schema
	schemaOnce sync.Once
	schemaErr  error
)

// SchemaDir must be set at startup to the absolute path of the schemas/ directory.
var SchemaDir string

// schemaFS serves xs:include/xs:import only from inside root, so a schema
// can never pull a file from elsewhere on the host.
type schemaFS struct{ root string }

func (s schemaFS) Open(name string) (fs.File, error) {
	rel, err := filepath.Rel(s.root, filepath.Clean(name))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}
	return os.Open(filepath.Join(s.root, rel))
}

func loadSchema() (*xsd.Schema, error) {
	schemaOnce.Do(func() {
		if SchemaDir == "" {
			cwd, _ := os.Getwd()
			SchemaDir = filepath.Join(cwd, "internal", "parser", "schemas")
		}
		dir, err := filepath.Abs(SchemaDir)
		if err != nil {
			schemaErr = fmt.Errorf("schema dir: %w", err)
			return
		}

		path := filepath.Join(dir, "procNFe_v4.00.xsd")
		if _, err := os.Stat(path); err != nil {
			schemaErr = fmt.Errorf("XSD schema not found at %s: %w", path, err)
			return
		}

		schema, schemaErr = xsd.NewCompiler().
			BaseDir(dir).
			FS(schemaFS{root: dir}).
			CompileFile(context.Background(), path)
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

	ctx := context.Background()
	doc, err := helium.NewParser().Parse(ctx, data)
	if err != nil {
		return fmt.Errorf("parse XML: %w", err)
	}

	errs := helium.NewErrorCollector(ctx, helium.ErrorLevelNone)
	if err := xsd.NewValidator(s).ErrorHandler(errs).Validate(ctx, doc); err != nil {
		if errors.Is(err, xsd.ErrValidationFailed) {
			var msgs []string
			for _, e := range errs.Errors() {
				msgs = append(msgs, e.Error())
			}
			return fmt.Errorf("XSD validation errors:\n%s", strings.Join(msgs, "\n"))
		}
		return fmt.Errorf("XSD validation: %w", err)
	}

	return nil
}
