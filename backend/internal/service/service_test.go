package service

import (
	"testing"

	"github.com/nfe-processor/backend/internal/domain"
	"github.com/nfe-processor/backend/internal/mock"
)

// classify only touches s.clients, so a bare struct literal is enough --
// no repository/queue/config needed, which would otherwise require a live
// Postgres and RabbitMQ just to test a pure classification decision.
func newTestService() *NFeService {
	return &NFeService{clients: &mock.ClientService{}}
}

func TestClassify_RecipientIsInternalClient_Purchase(t *testing.T) {
	// Empresa Alpha (internal/mock/clients.go) as recipient -- same fixture
	// relationship as examples/nfe_compra.xml.
	nfe := &domain.NFe{
		IssuerCNPJ:    "59407816000187",
		RecipientCNPJ: "10433218000193",
	}

	newTestService().classify(nfe)

	if nfe.Operation != domain.OperationPurchase {
		t.Errorf("Operation = %q, want %q", nfe.Operation, domain.OperationPurchase)
	}
	if nfe.LinkedClient != "Empresa Alpha Ltda" {
		t.Errorf("LinkedClient = %q, want %q", nfe.LinkedClient, "Empresa Alpha Ltda")
	}
}

func TestClassify_IssuerIsInternalClient_Sale(t *testing.T) {
	// Empresa Beta as issuer -- same fixture relationship as
	// examples/nfe_venda.xml.
	nfe := &domain.NFe{
		IssuerCNPJ:    "19600133000127",
		RecipientCNPJ: "18495931000173",
	}

	newTestService().classify(nfe)

	if nfe.Operation != domain.OperationSale {
		t.Errorf("Operation = %q, want %q", nfe.Operation, domain.OperationSale)
	}
	if nfe.LinkedClient != "Empresa Beta S.A." {
		t.Errorf("LinkedClient = %q, want %q", nfe.LinkedClient, "Empresa Beta S.A.")
	}
}

func TestClassify_NeitherPartyIsInternalClient_Unidentified(t *testing.T) {
	// Same fixture relationship as examples/nfe_nao_identificada.xml.
	nfe := &domain.NFe{
		IssuerCNPJ:    "03413164000139",
		RecipientCNPJ: "75255341000107",
	}

	newTestService().classify(nfe)

	if nfe.Operation != domain.OperationUnidentified {
		t.Errorf("Operation = %q, want %q", nfe.Operation, domain.OperationUnidentified)
	}
	if nfe.LinkedClient != "" {
		t.Errorf("LinkedClient = %q, want empty", nfe.LinkedClient)
	}
	if nfe.UnidentifiedNote == "" {
		t.Errorf("UnidentifiedNote is empty, want an explanation")
	}
}

func TestClassify_RecipientMatchTakesPriorityOverIssuer(t *testing.T) {
	// A degenerate case the real client list can't produce (every CNPJ in
	// internal/mock/clients.go is distinct), but classify() checks
	// recipient before issuer -- pin that ordering down explicitly so a
	// future reordering is caught even though no current fixture exercises it.
	nfe := &domain.NFe{
		IssuerCNPJ:    "19600133000127", // Empresa Beta
		RecipientCNPJ: "10433218000193", // Empresa Alpha
	}

	newTestService().classify(nfe)

	if nfe.Operation != domain.OperationPurchase {
		t.Errorf("Operation = %q, want %q (recipient check should win)", nfe.Operation, domain.OperationPurchase)
	}
}
