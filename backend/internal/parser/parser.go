package parser

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nfe-processor/backend/internal/domain"
)

// nfeProc is the authorized NF-e wrapper returned by SEFAZ.
// Only this format is accepted — bare <NFe> documents are not authorized.
type nfeProc struct {
	XMLName xml.Name `xml:"nfeProc"`
	NFe     struct {
		InfNFe struct {
			ID  string `xml:"Id,attr"`
			Ide struct {
				DhEmi string `xml:"dhEmi"`
			} `xml:"ide"`
			Emit struct {
				CNPJ  string `xml:"CNPJ"`
				XNome string `xml:"xNome"`
			} `xml:"emit"`
			Dest struct {
				CNPJ  string `xml:"CNPJ"`
				CPF   string `xml:"CPF"`
				XNome string `xml:"xNome"`
			} `xml:"dest"`
			Total struct {
				ICMSTot struct {
					VNF string `xml:"vNF"`
				} `xml:"ICMSTot"`
			} `xml:"total"`
		} `xml:"infNFe"`
	} `xml:"NFe"`
	ProtNFe struct {
		InfProt struct {
			ChNFe    string `xml:"chNFe"`
			DhRecbto string `xml:"dhRecbto"`
			NProt    string `xml:"nProt"`
			CStat    string `xml:"cStat"`
		} `xml:"infProt"`
	} `xml:"protNFe"`
}

// ParseNFe parses and validates an authorized NF-e XML (nfeProc format).
//
// Validation pipeline:
//  1. XSD schema          — structural completeness against procNFe_v4.00.xsd
//  2. Access key format   — length, digits only, model code (55/65)
//  3. Access key checksum — mod 11 check digit (SEFAZ MOC section 5.4)
//  4. Issuer CNPJ mod 11  — validates both check digits (Receita Federal)
//  5. CNPJ consistency    — CNPJ in access key must match issuer CNPJ in XML body
func ParseNFe(data []byte) (*domain.NFe, error) {
	// 1. XSD validation — rejects non-nfeProc and structurally invalid documents
	if err := validateXSD(data); err != nil {
		return nil, fmt.Errorf("schema validation: %w", err)
	}

	var doc nfeProc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse XML: %w", err)
	}

	inf := doc.NFe.InfNFe
	accessKey := strings.TrimPrefix(inf.ID, "NFe")
	issuerCNPJ := sanitize(inf.Emit.CNPJ)

	// 2 & 3. Access key format + mod 11 check digit
	if err := validateAccessKey(accessKey); err != nil {
		return nil, fmt.Errorf("invalid access key: %w", err)
	}

	// 4. Issuer CNPJ mod 11
	if issuerCNPJ == "" {
		return nil, fmt.Errorf("issuer CNPJ not found")
	}
	if err := validateCNPJ(issuerCNPJ); err != nil {
		return nil, fmt.Errorf("invalid issuer CNPJ: %w", err)
	}

	// 5. CNPJ in access key must match issuer CNPJ in XML body
	if err := validateCNPJConsistency(accessKey, issuerCNPJ); err != nil {
		return nil, fmt.Errorf("inconsistent NF-e: %w", err)
	}

	issuedAt, err := parseDate(inf.Ide.DhEmi)
	if err != nil {
		issuedAt = time.Now()
	}

	totalAmount, _ := strconv.ParseFloat(inf.Total.ICMSTot.VNF, 64)

	recipientCNPJ := inf.Dest.CNPJ
	if recipientCNPJ == "" {
		recipientCNPJ = inf.Dest.CPF
	}

	return &domain.NFe{
		AccessKey:     accessKey,
		IssuerName:    inf.Emit.XNome,
		IssuerCNPJ:    issuerCNPJ,
		RecipientName: inf.Dest.XNome,
		RecipientCNPJ: sanitize(recipientCNPJ),
		TotalAmount:   totalAmount,
		IssuedAt:      issuedAt,
	}, nil
}

func parseDate(raw string) (time.Time, error) {
	for _, f := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(f, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date format: %s", raw)
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
