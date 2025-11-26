package database

import (
	"context"
	"time"
	"errors"
	
	"github.com/go-card/internal/core/model"
	"github.com/go-card/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var (
	tracerProvider go_core_observ.TracerProvider
	childLogger = log.With().Str("component","go-card").Str("package","internal.adapter.database").Logger()
)

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

// Above new worker
func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Info().Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

// Above get stats from database
func (w WorkerRepository) Stat(ctx context.Context) (go_core_pg.PoolStats){
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	stats := w.DatabasePGServer.Stat()

	resPoolStats := go_core_pg.PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}

	return resPoolStats
}

// Above add card
func (w WorkerRepository) AddCard(ctx context.Context, tx pgx.Tx, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","AddCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "database.AddCard")
	defer span.End()

	// prepare
	card.CreatedAt = time.Now()
	card.ExpiredAt = time.Now().AddDate(5, 0, 0) // add 5 year
	card.Atc = 0

	//query
	query := `INSERT INTO card (fk_account_id,
								card_number, 
								card_type,
								holder,
								card_model, 
								status,
								atc, 
								expired_at, 
								created_at, 
								tenant_id) 
								VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	
	// execute	
	row := tx.QueryRow(ctx, query,  card.FkAccountID,  
									card.CardNumber,
									card.Type,
									card.Holder,
									card.Model,
									card.Status,
									card.Atc,
									card.ExpiredAt,
									card.CreatedAt,
									card.TenantID,
									)

	var id int
	
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Send()			
		return nil, errors.New(err.Error())
	}

	card.ID = id
	
	return &card, nil
}

// Above get card
func (w WorkerRepository) GetCard(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","GetCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetCard")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()	
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_card := model.Card{}

	query := `SELECT  	cc.id,
						cc.fk_account_id,
						cc.card_number, 
						cc.card_type,
						cc.holder,
						cc.card_model, 
						cc.status,
						cc.atc, 
						cc.expired_at, 
						cc.created_at,
						cc.updated_at, 
						cc.tenant_id
				FROM card cc
				WHERE card_number = $1`

	// execute			
	rows, err := conn.Query(ctx, query, card.CardNumber)
	if err != nil {
		childLogger.Error().Err(err).Send()	
		return nil, errors.New(err.Error())
	}
	defer rows.Close()
    if err := rows.Err(); err != nil {
		childLogger.Error().Err(err).Msg("fatal error closing rows")
        return nil, errors.New(err.Error())
    }

	for rows.Next() {
		err := rows.Scan( 	&res_card.ID,
							&res_card.FkAccountID,
							&res_card.CardNumber, 
							&res_card.Type,
							&res_card.Holder,
							&res_card.Model,
							&res_card.Status,	
							&res_card.Atc,
							&res_card.ExpiredAt,
							&res_card.CreatedAt,
							&res_card.UpdatedAt,
							&res_card.TenantID,
						)
		if err != nil {
			childLogger.Error().Err(err).Send()	
			return nil, errors.New(err.Error())
        }
		return &res_card, nil
	}
	
	return nil, erro.ErrNotFound
}

// Above update atc
func (w WorkerRepository) UpdateCard(ctx context.Context, tx pgx.Tx, card model.Card) (int64, error){
	childLogger.Info().Str("func","UpdateCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.UpdateCard")
	defer span.End()

	// prepare query
	t_updateAt := time.Now()
	card.UpdatedAt = &t_updateAt

	query := `Update public.card
				set atc = atc + 1, 
					updated_at = $2
				where card_number = $1`

	// execute
	row, err := tx.Exec(ctx, query, card.CardNumber,  
									card.UpdatedAt)
	if err != nil {
		childLogger.Error().Err(err).Send()			
		return 0, errors.New(err.Error())
	}

	if int(row.RowsAffected()) == 0 {
		return 0, erro.ErrUpdateRows
	}
	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")
	
	return row.RowsAffected(), nil
}

// About add token card 
func (w *WorkerRepository) CreateCardToken(ctx context.Context, tx pgx.Tx, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	//trace
	span := tracerProvider.Span(ctx, "database.CreateCardToken")
	defer span.End()

	// Query e Execute
	query := `INSERT INTO card_token(fk_id_card, 
									token,
									status,
									created_at,
									expired_at,
									tenant_id) 
			 VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	row := tx.QueryRow(	ctx, 
						query, 
						card.ID, 
						card.TokenData, 
						card.Status, 
						card.CreatedAt, 
						card.ExpiredAt, 
						card.TenantID)								
	var id int
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Send()	
		return nil, errors.New(err.Error())
	}

	card.ID = id

	return &card , nil
}

// About add token card 
func (w *WorkerRepository) GetCardToken(ctx context.Context, card model.Card) (*[]model.Card, error){
	childLogger.Info().Str("func","GetCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	//trace
	span := tracerProvider.Span(ctx, "database.GetCardToken")
	defer span.End()

	// Prepare
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()	
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	res_card := model.Card{}
	res_card_list := []model.Card{}
	
	// Query e Execute
	query := `SELECT ct.id, 
					ca.card_number,
					ca.card_model, 
					ct.token,
					ct.status,
					ct.expired_at,
					ct.created_at,
					ct.updated_at,																									
					ct.tenant_id	
				FROM card_token ct,
					card ca
				WHERE ct.token = $1
				and ca.id = ct.fk_id_card 
				order by ct.created_at desc`

	rows, err := conn.Query(ctx, query, string(card.TokenData))
	if err != nil {
		childLogger.Error().Err(err).Send()	
		return nil, errors.New(err.Error())
	}
	defer rows.Close()
    if err := rows.Err(); err != nil {
		childLogger.Error().Err(err).Msg("fatal error closing rows")
        return nil, errors.New(err.Error())
    }
	
	for rows.Next() {
		err := rows.Scan( 	&res_card.ID, 
							&res_card.CardNumber,
							&res_card.Model, 
							&res_card.TokenData, 
							&res_card.Status,
							&res_card.ExpiredAt,
							&res_card.CreatedAt,
							&res_card.UpdatedAt,
							&res_card.TenantID)
		if err != nil {
			childLogger.Error().Err(err).Send()	
			return nil, errors.New(err.Error())
        }
		res_card_list = append(res_card_list, res_card)
	}

	return &res_card_list , nil
}