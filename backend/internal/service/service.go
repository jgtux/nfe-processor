package service

import (
	"fmt"
	"log"

	"github.com/nfe-processor/backend/internal/domain"
	"github.com/nfe-processor/backend/internal/mock"
	"github.com/nfe-processor/backend/internal/parser"
	"github.com/nfe-processor/backend/internal/queue"
	"github.com/nfe-processor/backend/internal/repository"
)

type NFeService struct {
	repo    *repository.NFeRepository
	queue   *queue.RabbitMQ
	clients *mock.ClientService
}

func New(repo *repository.NFeRepository, q *queue.RabbitMQ) *NFeService {
	svc := &NFeService{
		repo:    repo,
		queue:   q,
		clients: &mock.ClientService{},
	}
	if err := q.Consume(svc.processMessage); err != nil {
		log.Fatalf("[service] failed to start consumer: %v", err)
	}
	return svc
}

// EnqueueXML persists the raw XML in the database and publishes only the upload ID to the queue.
func (s *NFeService) EnqueueXML(xmlData []byte) (string, error) {
	if len(xmlData) == 0 {
		return "", fmt.Errorf("empty XML file")
	}

	uploadID, err := s.repo.SaveUpload(xmlData)
	if err != nil {
		return "", fmt.Errorf("save upload: %w", err)
	}

	if err := s.queue.Publish(domain.QueueMessage{UploadID: uploadID}); err != nil {
		return "", fmt.Errorf("enqueue: %w", err)
	}

	log.Printf("[service] enqueued upload_id=%s", uploadID)
	return uploadID, nil
}

// processMessage is called by the RabbitMQ consumer for each queued message.
func (s *NFeService) processMessage(msg domain.QueueMessage) error {
	log.Printf("[consumer] processing upload_id=%s", msg.UploadID)

	xmlData, err := s.repo.GetUpload(msg.UploadID)
	if err != nil {
		return fmt.Errorf("get upload %s: %w", msg.UploadID, err)
	}

	nfe, err := parser.ParseNFe(xmlData)
	if err != nil {
		log.Printf("[consumer] parse error upload_id=%s: %v", msg.UploadID, err)
		_ = s.repo.UpsertNFe(&domain.NFe{
			UploadID:  msg.UploadID,
			AccessKey: msg.UploadID, // fallback key to satisfy UNIQUE constraint
			Status:    domain.StatusError,
			ErrorMsg:  err.Error(),
			Operation: domain.OperationUnidentified,
		})
		return nil // ack — permanently broken XML should not be requeued
	}

	nfe.UploadID = msg.UploadID
	s.classify(nfe)
	nfe.Status = domain.StatusProcessed

	if err := s.repo.UpsertNFe(nfe); err != nil {
		return fmt.Errorf("persist nfe access_key=%s: %w", nfe.AccessKey, err)
	}

	log.Printf("[consumer] processed access_key=%s operation=%s", nfe.AccessKey, nfe.Operation)
	return nil
}

func (s *NFeService) classify(nfe *domain.NFe) {
	if c := s.clients.FindByCNPJ(nfe.RecipientCNPJ); c != nil {
		nfe.Operation = domain.OperationPurchase
		nfe.LinkedClient = c.Name
		return
	}
	if c := s.clients.FindByCNPJ(nfe.IssuerCNPJ); c != nil {
		nfe.Operation = domain.OperationSale
		nfe.LinkedClient = c.Name
		return
	}
	nfe.Operation = domain.OperationUnidentified
	nfe.UnidentifiedNote = "neither issuer nor recipient CNPJ matches an internal client"
}

func (s *NFeService) ListAll() ([]domain.NFe, error)              { return s.repo.ListAll() }
func (s *NFeService) ListUnidentified() ([]domain.NFe, error)     { return s.repo.ListUnidentified() }
func (s *NFeService) ClientSummary() ([]domain.ClientSummary, error) { return s.repo.ClientSummary() }
func (s *NFeService) InternalClients() ([]domain.InternalClient, error) {
	return s.clients.GetAll()
}
