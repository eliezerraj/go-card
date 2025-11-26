package api

import (
	"context"
	"time"
	"fmt"
	"encoding/json"
	"reflect"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/go-card/internal/core/service"
	"github.com/go-card/internal/core/model"
	"github.com/go-card/internal/core/erro"

	"github.com/gorilla/mux"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_json "github.com/eliezerraj/go-core/coreJson"
)

var (
	childLogger = log.With().Str("component", "go-card").Str("package", "internal.adapter.api").Logger()
	core_json		go_core_json.CoreJson
	core_apiError 	go_core_json.APIError
	tracerProvider 	go_core_observ.TracerProvider
)

type HttpRouters struct {
	workerService 	*service.WorkerService
	ctxTimeout		time.Duration
}

// Above create routers
func NewHttpRouters(workerService *service.WorkerService,
					ctxTimeout	time.Duration) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
		ctxTimeout: ctxTimeout,
	}
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Health").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About show all context values
func (h *HttpRouters) Context(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Context").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	contextValues := reflect.ValueOf(req.Context()).Elem()
	json.NewEncoder(rw).Encode(fmt.Sprintf("%v",contextValues))
}

// About show pgx stats
func (h *HttpRouters) Stat(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	res := h.workerService.Stat(req.Context())

	json.NewEncoder(rw).Encode(res)
}

// About handle error
func (h *HttpRouters) ErrorHandler(trace_id string, err error) *go_core_json.APIError {
	if strings.Contains(err.Error(), "context deadline exceeded") {
    	err = erro.ErrTimeout
	} 
	switch err {
	case erro.ErrBadRequest:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
	case erro.ErrNotFound:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
	case erro.ErrTimeout:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusGatewayTimeout)
	default:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
	}
	return &core_apiError
}

// About add card
func (h *HttpRouters) AddCard(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","AddCard").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.api.AddCard")
	defer span.End()
	
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	card := model.Card{}
	err := json.NewDecoder(req.Body).Decode(&card)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	defer req.Body.Close()

	res, err := h.workerService.AddCard(ctx, card)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get card
func (h *HttpRouters) GetCard(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetCard").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.api.GetCard")
	defer span.End()

	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	vars := mux.Vars(req)
	varID := vars["id"]

	card := model.Card{}
	card.CardNumber = varID

	res, err := h.workerService.GetCard(ctx, card)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About update card
func (h *HttpRouters) UpdateCard(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","UpdateCard").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

    ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.api.UpdateCard")
	defer span.End()

	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	card := model.Card{}
	err := json.NewDecoder(req.Body).Decode(&card)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	defer req.Body.Close()

	res, err := h.workerService.UpdateCard(ctx, card)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About add card
func (h *HttpRouters) CreateCardToken(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.api.CreateCardToken")
	defer span.End()

	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	card := model.Card{}
	err := json.NewDecoder(req.Body).Decode(&card)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	defer req.Body.Close()

	res, err := h.workerService.CreateCardToken(ctx, card)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get card
func (h *HttpRouters) GetCardToken(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetCardToken").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	ctx, span := tracerProvider.SpanCtx(ctx, "adapter.api.GetCardToken")
	defer span.End()
	
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	vars := mux.Vars(req)
	varID := vars["id"]

	card := model.Card{}
	card.TokenData = varID

	res, err := h.workerService.GetCardToken(ctx, card)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}
