package service

import(
	"fmt"
	"time"
	"context"
	"errors"
	"net/http"	
	"encoding/json"
	"github.com/zeebo/blake3"

	"github.com/rs/zerolog/log"

	"github.com/go-card/internal/core/model"
	"github.com/go-card/internal/core/erro"
	"github.com/go-card/internal/adapter/database"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("component","go-card").Str("package","internal.core.service").Logger()
var apiService go_core_api.ApiService

type WorkerService struct {
	workerRepository 	*database.WorkerRepository
	apiService			[]model.ApiService
}

// About create a new worker service
func NewWorkerService(	workerRepository 	*database.WorkerRepository,
						apiService			[]model.ApiService,) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		apiService: apiService,
		workerRepository: workerRepository,
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int) error{
	var err error
	switch statusCode {
		case http.StatusUnauthorized:
			err = erro.ErrUnauthorized
		case http.StatusForbidden:
			err = erro.ErrHTTPForbiden
		case http.StatusNotFound:
			err = erro.ErrNotFound
		default:
			err = erro.ErrServer
		}
	return err
}

// About create a card
func (s *WorkerService) AddCard(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","AddCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.AddCard")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()
	
	// prepare database
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// Get the Account ID (PK) from Account-service
	// Set headers
	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[0].XApigwApiId,
		"Host": s.apiService[0].HostName,
	}
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + card.AccountID,
		Method: s.apiService[0].Method,
		Timeout: 15,
		Headers: &headers,
	}

	res_payload, statusCode, err := apiService.CallRestApi(	ctx,
															httpClient, 
															nil)
	
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	card.FkAccountID = account_parsed.ID
	// add card
	res, err := s.workerRepository.AddCard(ctx, tx, card)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About get a card
func (s *WorkerService) GetCard(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","GetCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.GetCard")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	defer span.End()
	
	// get card
	res_card, err := s.workerRepository.GetCard(ctx, card)
	if err != nil {
		return nil, err
	}

	// Get the Account ID (PK) from Account-service
	// Set headers
	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[0].XApigwApiId,
		"Host": s.apiService[0].HostName,
	}
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/getId/" + fmt.Sprintf("%v",res_card.FkAccountID),
		Method: s.apiService[0].Method,
		Timeout: 15,
		Headers: &headers,
	}
	// get account_if from id (PK)
	res_payload, statusCode, err := apiService.CallRestApi(	ctx,
															httpClient, 
															nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)
	res_card.AccountID = account_parsed.AccountID

	return res_card, nil
}

// About update a update
func (s *WorkerService) UpdateCard(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","UpdateCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.UpdateCard")
	defer span.End()

	// prepare database
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	//Check data exists
	_, err = s.workerRepository.GetCard(ctx, card)
	if err != nil {
		return nil, err
	}

	// Do update atc
	res_update, err := s.workerRepository.UpdateCard(ctx, tx, card)
	if err != nil {
		return nil, err
	}
	if (res_update == 0) {
		return nil, erro.ErrUpdate
	}

	//Get atc data
	res_card, err := s.workerRepository.GetCard(ctx, card)
	if err != nil {
		return nil, err
	}

	return res_card, nil
}

// About create a tokenization data
func (s * WorkerService) CreateCardToken(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.CreateCardToken")

	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}

	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// Get cards info from token (FkAccountID)
	res_card, err := s.workerRepository.GetCard(ctx, card)
	if err != nil {
		return nil, err
	}

	// prepare data
	hasher := blake3.New()
	hasher.Write([]byte(card.CardNumber))
	card.TokenData = fmt.Sprintf("%x", (hasher.Sum(nil)) )
	card.Status = "ACTIVE"

	card.CreatedAt = time.Now()
	card.ExpiredAt = time.Now().AddDate(0, 3, 0) // Add 3 months
	card.ID = res_card.ID

	// Call a service
	res, err := s.workerRepository.CreateCardToken(ctx, tx, card)
	if err != nil {
		return nil, err
	}

	// Setting PK
	card.ID = res.ID

	return &card, nil
}

// About get the card from token
func (s * WorkerService) GetCardToken(ctx context.Context, card model.Card) (*[]model.Card, error){
	childLogger.Info().Str("func","GetCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.GetCardToken")
	defer span.End()

	// Call a service
	res, err := s.workerRepository.GetCardToken(ctx, card)
	if err != nil {
		return nil, err
	}

	return res, nil
}
