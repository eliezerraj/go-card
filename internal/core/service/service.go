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

	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var (
	tracerProvider go_core_observ.TracerProvider
	childLogger = log.With().Str("component","go-card").Str("package","internal.core.service").Logger()
	apiService go_core_api.ApiService
)

type WorkerService struct {
	goCoreRestApiService	go_core_api.ApiService
	workerRepository 		*database.WorkerRepository
	apiService				[]model.ApiService
}

// About create a new worker service
func NewWorkerService(	goCoreRestApiService	go_core_api.ApiService,	
						workerRepository 		*database.WorkerRepository,
						apiService				[]model.ApiService,) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		goCoreRestApiService: 	goCoreRestApiService,
		apiService: 			apiService,
		workerRepository: 		workerRepository,
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int, serviceName string, msg_err error) error{
	childLogger.Info().Str("func","errorStatusCode").Interface("serviceName", serviceName).Interface("statusCode", statusCode).Send()

	var err error
	switch statusCode {
		case http.StatusUnauthorized:
			err = erro.ErrUnauthorized
		case http.StatusForbidden:
			err = erro.ErrHTTPForbiden
		case http.StatusNotFound:
			err = erro.ErrNotFound
		default:
			err = errors.New(fmt.Sprintf("service %s in outage => cause error: %s", serviceName, msg_err.Error() ))
		}
	return err
}

// About handle/convert http status code
func (s *WorkerService) Stat(ctx context.Context) (go_core_pg.PoolStats){
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	return s.workerRepository.Stat(ctx)
}

// About create a card
func (s *WorkerService) AddCard(ctx context.Context, card model.Card) (*model.Card, error){
	childLogger.Info().Str("func","AddCard").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("card", card).Send()

	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.AddCard")
	defer span.End()
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	// prepare database
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)

	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
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
	// Set client http	
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + card.AccountID,
		Method: s.apiService[0].Method,
		Timeout: s.apiService[0].HttpTimeout,
		Headers: &headers,
	}

	res_payload, statusCode, err := apiService.CallRestApiV1(	ctx,
																s.goCoreRestApiService.Client,
																httpClient, 
																nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[0].Name, err)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// prepare data, set ID (PK_)
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

	// span and trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetCard")
	defer span.End()

	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

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

	// Set client http	
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/getId/" + fmt.Sprintf("%v",res_card.FkAccountID),
		Method: s.apiService[0].Method,
		Timeout: s.apiService[0].HttpTimeout,
		Headers: &headers,
	}
	// get account_if from id (PK)
	res_payload, statusCode, err := apiService.CallRestApiV1(ctx,
															s.goCoreRestApiService.Client,
															httpClient, 
															nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[0].Name, err)
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
	ctx, span := tracerProvider.SpanCtx(ctx, "service.UpdateCard")
	defer span.End()

	// prepare database
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)

	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
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

	tx.Commit(ctx) // commit to get the new values updated below

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
	ctx, span := tracerProvider.SpanCtx(ctx, "service.CreateCardToken")

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
	ctx, span := tracerProvider.SpanCtx(ctx, "service.GetCardToken")
	defer span.End()

	// Call a service
	res, err := s.workerRepository.GetCardToken(ctx, card)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About check health service
func (s * WorkerService) HealthCheck(ctx context.Context) error{
	childLogger.Info().Str("func","HealthCheck").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// Trace
	ctx, span := tracerProvider.SpanCtx(ctx, "service.HealthCheck")
	defer span.End()
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	// Check database health
	err := s.workerRepository.DatabasePGServer.Ping()
	if err != nil {
		log.Error().Err(err).Msg("*** Database HEALTH FAILED ***")
		return erro.ErrHealthCheck
	}
	childLogger.Info().Str("func","HealthCheck").Msg("*** Database HEALTH SUCCESSFULL ***")

	// Set headers
	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"X-Request-Id": trace_id,
		"Host": s.apiService[0].HostName,
	}

	// Set client http	
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/health",
		Method: s.apiService[0].Method,
		Timeout: s.apiService[0].HttpTimeout,
		Headers: &headers,
	}

	// get account_if from id (PK)
	_, _, err = apiService.CallRestApiV1(	ctx,
											s.goCoreRestApiService.Client,
											httpClient, 
											nil)
	if err != nil {
		log.Error().Err(err).Msg("*** Service ACCOUNT HEALTH FAILED ***")
		return erro.ErrHealthCheck
	}
	childLogger.Info().Str("func","HealthCheck").Msg("*** Service ACCOUNT HEALTH SUCCESSFULL ***")
	
	return nil
}