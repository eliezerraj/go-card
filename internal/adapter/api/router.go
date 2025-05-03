package api

import (
	"fmt"
	"encoding/json"
	"reflect"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-card/internal/core/service"
	"github.com/go-card/internal/core/model"
	"github.com/go-card/internal/core/erro"

	"github.com/eliezerraj/go-core/coreJson"
	"github.com/gorilla/mux"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component", "go-card").Str("package", "internal.adapter.api").Logger()

var core_json coreJson.CoreJson
var core_apiError coreJson.APIError
var tracerProvider go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

// Above create routers
func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
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

// About add card
func (h *HttpRouters) AddCard(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","AddCard").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	span := tracerProvider.Span(req.Context(), "adapter.api.AddCard")
	defer span.End()
	
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	card := model.Card{}
	err := json.NewDecoder(req.Body).Decode(&card)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.AddCard(req.Context(), card)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get card
func (h *HttpRouters) GetCard(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetCard").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	span := tracerProvider.Span(req.Context(), "adapter.api.GetCard")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	vars := mux.Vars(req)
	varID := vars["id"]

	card := model.Card{}
	card.CardNumber = varID

	res, err := h.workerService.GetCard(req.Context(), card)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About update card
func (h *HttpRouters) UpdateCard(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","UpdateCard").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	span := tracerProvider.Span(req.Context(), "adapter.api.UpdateCard")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	card := model.Card{}
	err := json.NewDecoder(req.Body).Decode(&card)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.UpdateCard(req.Context(), card)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About add card
func (h *HttpRouters) CreateCardToken(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	span := tracerProvider.Span(req.Context(), "adapter.api.CreateCardToken")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	card := model.Card{}
	err := json.NewDecoder(req.Body).Decode(&card)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.CreateCardToken(req.Context(), card)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get card
func (h *HttpRouters) GetCardToken(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetCardToken").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	span := tracerProvider.Span(req.Context(), "adapter.api.GetCardToken")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	vars := mux.Vars(req)
	varID := vars["id"]

	card := model.Card{}
	card.TokenData = varID

	res, err := h.workerService.GetCardToken(req.Context(), card)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id ,http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}
