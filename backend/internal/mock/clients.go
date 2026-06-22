package mock

import "github.com/nfe-processor/backend/internal/domain"

// internalClients uses valid CNPJs (mod 11 verified).
var internalClients = []domain.InternalClient{
	{ID: "1", Name: "Empresa Alpha Ltda",      CNPJ: "10433218000193"},
	{ID: "2", Name: "Empresa Beta S.A.",        CNPJ: "19600133000127"},
	{ID: "3", Name: "Comércio Gama Eireli",     CNPJ: "89083863000183"},
	{ID: "4", Name: "Distribuidora Delta ME",   CNPJ: "79402654000100"},
	{ID: "5", Name: "Indústria Épsilon Ltda",   CNPJ: "23511615000188"},
}

type ClientService struct{}

func (s *ClientService) GetAll() ([]domain.InternalClient, error) {
	return internalClients, nil
}

func (s *ClientService) FindByCNPJ(cnpj string) *domain.InternalClient {
	for i, c := range internalClients {
		if c.CNPJ == cnpj {
			return &internalClients[i]
		}
	}
	return nil
}
