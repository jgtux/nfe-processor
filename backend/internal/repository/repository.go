package repository

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nfe-processor/backend/internal/config"
	"github.com/nfe-processor/backend/internal/domain"
)

type NFeRepository struct {
	db *sqlx.DB
}

func New(cfg *config.DBConfig) (*NFeRepository, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	repo := &NFeRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return repo, nil
}

func (r *NFeRepository) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS nfe_uploads (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			xml_data   BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS nfes (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			upload_id         UUID NOT NULL REFERENCES nfe_uploads(id),
			access_key        VARCHAR(44) UNIQUE NOT NULL,
			issuer_name       TEXT NOT NULL,
			issuer_cnpj       VARCHAR(14) NOT NULL,
			recipient_name    TEXT NOT NULL,
			recipient_cnpj    VARCHAR(14) NOT NULL,
			total_amount      NUMERIC(15,2) NOT NULL DEFAULT 0,
			issued_at         TIMESTAMPTZ NOT NULL,
			operation         VARCHAR(20) NOT NULL DEFAULT 'unidentified',
			linked_client     TEXT,
			unidentified_note TEXT,
			status            VARCHAR(20) NOT NULL DEFAULT 'pending',
			error_msg         TEXT,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_nfes_operation ON nfes(operation);
		CREATE INDEX IF NOT EXISTS idx_nfes_status    ON nfes(status);
		CREATE INDEX IF NOT EXISTS idx_nfes_upload_id ON nfes(upload_id);
	`)
	if err != nil {
		return err
	}
	log.Println("[repo] migrations applied")
	return nil
}

// SaveUpload persists the raw XML and returns the generated upload ID.
func (r *NFeRepository) SaveUpload(xmlData []byte) (string, error) {
	id := uuid.New().String()
	_, err := r.db.Exec(
		`INSERT INTO nfe_uploads (id, xml_data) VALUES ($1, $2)`,
		id, xmlData,
	)
	return id, err
}

// GetUpload retrieves the raw XML for a given upload ID.
func (r *NFeRepository) GetUpload(uploadID string) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(`SELECT xml_data FROM nfe_uploads WHERE id = $1`, uploadID).Scan(&data)
	return data, err
}

// UpsertNFe inserts or updates a processed NF-e record.
func (r *NFeRepository) UpsertNFe(nfe *domain.NFe) error {
	if nfe.ID == "" {
		nfe.ID = uuid.New().String()
	}
	_, err := r.db.Exec(`
		INSERT INTO nfes (
			id, upload_id, access_key, issuer_name, issuer_cnpj,
			recipient_name, recipient_cnpj, total_amount, issued_at,
			operation, linked_client, unidentified_note, status, error_msg,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW(),NOW()
		)
		ON CONFLICT (access_key) DO UPDATE SET
			issuer_name       = EXCLUDED.issuer_name,
			issuer_cnpj       = EXCLUDED.issuer_cnpj,
			recipient_name    = EXCLUDED.recipient_name,
			recipient_cnpj    = EXCLUDED.recipient_cnpj,
			total_amount      = EXCLUDED.total_amount,
			issued_at         = EXCLUDED.issued_at,
			operation         = EXCLUDED.operation,
			linked_client     = EXCLUDED.linked_client,
			unidentified_note = EXCLUDED.unidentified_note,
			status            = EXCLUDED.status,
			error_msg         = EXCLUDED.error_msg,
			updated_at        = NOW()
	`,
		nfe.ID, nfe.UploadID, nfe.AccessKey, nfe.IssuerName, nfe.IssuerCNPJ,
		nfe.RecipientName, nfe.RecipientCNPJ, nfe.TotalAmount, nfe.IssuedAt,
		nfe.Operation, nfe.LinkedClient, nfe.UnidentifiedNote, nfe.Status, nfe.ErrorMsg,
	)
	return err
}

func (r *NFeRepository) ListAll() ([]domain.NFe, error) {
	var list []domain.NFe
	err := r.db.Select(&list, `SELECT * FROM nfes ORDER BY created_at DESC`)
	return list, err
}

func (r *NFeRepository) ListUnidentified() ([]domain.NFe, error) {
	var list []domain.NFe
	err := r.db.Select(&list, `
		SELECT * FROM nfes WHERE operation = 'unidentified' ORDER BY created_at DESC
	`)
	return list, err
}

func (r *NFeRepository) ClientSummary() ([]domain.ClientSummary, error) {
	var rows []domain.ClientSummary
	err := r.db.Select(&rows, `
		SELECT
			linked_client                                          AS client,
			COUNT(*) FILTER (WHERE operation = 'purchase')::int   AS purchases,
			COUNT(*) FILTER (WHERE operation = 'sale')::int       AS sales
		FROM nfes
		WHERE operation IN ('purchase','sale')
		  AND linked_client IS NOT NULL
		GROUP BY linked_client
		ORDER BY linked_client
	`)
	return rows, err
}

func (r *NFeRepository) DB() *sqlx.DB { return r.db }
