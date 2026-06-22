package domain

import "time"

type OperationType string

const (
	OperationPurchase     OperationType = "purchase"
	OperationSale         OperationType = "sale"
	OperationUnidentified OperationType = "unidentified"
)

type ProcessingStatus string

const (
	StatusPending   ProcessingStatus = "pending"
	StatusProcessed ProcessingStatus = "processed"
	StatusError     ProcessingStatus = "error"
)

type NFe struct {
	ID               string           `db:"id"                json:"id"`
	UploadID         string           `db:"upload_id"         json:"upload_id"`
	AccessKey        string           `db:"access_key"        json:"access_key"`
	IssuerName       string           `db:"issuer_name"       json:"issuer_name"`
	IssuerCNPJ       string           `db:"issuer_cnpj"       json:"issuer_cnpj"`
	RecipientName    string           `db:"recipient_name"    json:"recipient_name"`
	RecipientCNPJ    string           `db:"recipient_cnpj"    json:"recipient_cnpj"`
	TotalAmount      float64          `db:"total_amount"      json:"total_amount"`
	IssuedAt         time.Time        `db:"issued_at"         json:"issued_at"`
	Operation        OperationType    `db:"operation"         json:"operation"`
	LinkedClient     string           `db:"linked_client"     json:"linked_client,omitempty"`
	UnidentifiedNote string           `db:"unidentified_note" json:"unidentified_note,omitempty"`
	Status           ProcessingStatus `db:"status"            json:"status"`
	ErrorMsg         string           `db:"error_msg"         json:"error_msg,omitempty"`
	CreatedAt        time.Time        `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at"        json:"updated_at"`
}

type Upload struct {
	ID        string    `db:"id"         json:"id"`
	XMLData   []byte    `db:"xml_data"   json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type InternalClient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CNPJ string `json:"cnpj"`
}

type ClientSummary struct {
	Client    string `db:"client"    json:"client"`
	Purchases int    `db:"purchases" json:"purchases"`
	Sales     int    `db:"sales"     json:"sales"`
}

type QueueMessage struct {
	UploadID string `json:"upload_id"`
}
