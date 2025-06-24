package psqldb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Entity struct {
	ID string `json:"id"`
}

type DLTXReceipt struct {
	TransactionHash string `json:"transactionHash"`
	BlockNumberRaw  string `json:"blockNumberRaw"`
}

func getEntityID(jsonData []byte) string {
	var entity Entity
	if err := json.Unmarshal(jsonData, &entity); err != nil {
		entity.ID = ""
	}
	return entity.ID
}

func getEthReceipt(jsonData []byte) DLTXReceipt {
	var receipt DLTXReceipt
	if err := json.Unmarshal(jsonData, &receipt); err != nil {
		receipt.TransactionHash = ""
		receipt.BlockNumberRaw = ""
	}
	return receipt
}

// Operation struct will keep track of the data from and to PostgreSQL
type Operation struct {
	InData       json.RawMessage `json:"inData"`
	OutData      json.RawMessage `json:"outData"`
	CMStatus     uint16          `json:"cmStatus"`
	LDStatus     uint16          `json:"ldStatus"`
	Method       string          `json:"method"`
	Tenant       string          `json:"tenant"`
	LinkHdr      string          `json:"string"`
	EntityID     string          `json:"entityId"`
	TicketId     string          `json:"ticketId"`
	TicketNumber string          `json:"ticketNumber"`
	CreatedAt    string          `json:"createdAt"`
}

// OperationRepository handles database operations
type OperationRepository struct {
	pool *pgxpool.Pool
}

func CreateOperation(repo *OperationRepository, op *Operation) {
	ethTicket := getEthReceipt(op.OutData)
	entityId := getEntityID(op.InData)
	op.EntityID = entityId
	op.TicketId = ethTicket.TransactionHash
	op.TicketNumber = ethTicket.BlockNumberRaw

	ctx := context.Background()
	repo.CreateOperation(ctx, op)
}

func NewOperationRepository(connString string) (*OperationRepository, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("Unable to create connection pool: %w", err)
	}

	return &OperationRepository{pool: pool}, nil
}

// Close - Function to close the connection pool
func (r *OperationRepository) Close() {
	if r != nil {
		r.pool.Close()
	}
}

// CreateOperation inserts a new operation record in the database
func (r *OperationRepository) CreateOperation(ctx context.Context, op *Operation) error {
	query := `
	   INSERT INTO operations (
	     in_data, out_data, method, tenant, entity_id, ticket_id, ticket_number, cm_status, ld_status, link_header)
	   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		op.InData,
		op.OutData,
		op.Method,
		op.Tenant,
		op.EntityID,
		op.TicketId,
		op.TicketNumber,
		op.CMStatus,
		op.LDStatus,
		op.LinkHdr,
	)

	if err != nil {
		log.Println(err)
	}

	return err
}

func (r *OperationRepository) GetOpertationsByURI(ctx context.Context, uri string) ([]Operation, error) {
	query := `
	   SELECT in_data, out_data, method, tenant, entity_id, ticket_id, ticket_number, created_at, cm_status, ld_status, link_header
	   from operations
	   where uri = $1
	   order by created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, uri)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return scanOperations(rows)
}

func scanOperations(rows pgx.Rows) ([]Operation, error) {
	var operations []Operation

	for rows.Next() {
		var op Operation
		err := rows.Scan(
			&op.InData,
			&op.OutData,
			&op.Method,
			&op.Tenant,
			&op.EntityID,
			&op.TicketId,
			&op.TicketNumber,
			&op.CreatedAt,
			&op.CMStatus,
			&op.LDStatus,
			&op.LinkHdr,
		)
		if err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return operations, nil
}

/*
Postgresql -

-- Create the table
CREATE TABLE operations (
    in_data JSONB,
    out_data JSONB,
	cm_status SMALLINT,
	ld_status SMALLINT,
    method TEXT,
    tenant TEXT,
    entity_id TEXT,
    ticket_id TEXT,
    ticket_number TEXT,
	link_header TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for your query patterns
CREATE INDEX idx_operations_uri_created_at ON operations (entity_id, created_at);
CREATE INDEX idx_operations_ticket_id_created_at ON operations (ticket_id, created_at);
CREATE INDEX idx_operations_ticket_number_created_at ON operations (ticket_number, created_at);
*/
