package server

import (
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"

	"github.com/go-card/internal/adapter/api"	
	"github.com/go-card/internal/core/model"
	go_core_observ "github.com/eliezerraj/go-core/observability"  

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/eliezerraj/go-core/middleware"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	//"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

var (
	childLogger = log.With().Str("component","go-card").Str("package","internal.infra.server").Logger()
	core_middleware middleware.ToolsMiddleware
	tracerProvider go_core_observ.TracerProvider
	infoTrace go_core_observ.InfoTrace
)

type HttpServer struct {
	httpServer	*model.Server
}

// About create new http server
func NewHttpAppServer(httpServer *model.Server) HttpServer {
	childLogger.Info().Str("func","NewHttpAppServer").Send()
	return HttpServer{httpServer: httpServer }
}

// About start http server
func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpRouters *api.HttpRouters,
										appServer *model.AppServer) {
	childLogger.Info().Str("func","StartHttpAppServer").Send()
			
	// ---------------------- OTEL ---------------
	infoTrace.PodName = appServer.InfoPod.PodName
	infoTrace.PodVersion = appServer.InfoPod.ApiVersion
	infoTrace.ServiceType = "k8-workload"
	infoTrace.Env = appServer.InfoPod.Env
	infoTrace.AccountID = appServer.InfoPod.AccountID

	tp := tracerProvider.NewTracerProvider(	ctx, 
											appServer.ConfigOTEL, 
											&infoTrace)
	
	if tp != nil {
		otel.SetTextMapPropagator(xray.Propagator{} // propagation.TraceContext{} xray.Propagator{}
		otel.SetTracerProvider(tp)
	}

	defer func() { 
		if tp != nil {
			err := tp.Shutdown(ctx)
			if err != nil{
				childLogger.Error().Err(err).Send()
			}
		}
		childLogger.Info().Msg("stop done !!!")
	}()
	
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(core_middleware.MiddleWareHandlerHeader)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(appServer)
	})

	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpRouters.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpRouters.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpRouters.Header)

	stat := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    stat.HandleFunc("/stat", httpRouters.Stat)

	wk_ctx := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    wk_ctx.HandleFunc("/context", httpRouters.Context)

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Info().Str("HandleFunc","/info").Send()

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(appServer)
	})
	
	addCard := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addCard.HandleFunc("/card", core_middleware.MiddleWareErrorHandler(httpRouters.AddCard))		
	addCard.Use(otelmux.Middleware("go-card"))

	getCard := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getCard.HandleFunc("/card/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetCard))		
	getCard.Use(otelmux.Middleware("go-card"))

	updateCard := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	updateCard.HandleFunc("/atc", core_middleware.MiddleWareErrorHandler(httpRouters.UpdateCard))		
	updateCard.Use(otelmux.Middleware("go-card"))

	getCardToken := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getCardToken.HandleFunc("/cardToken/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetCardToken))		
	getCardToken.Use(otelmux.Middleware("go-card"))
	
	createCardToken := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	createCardToken.HandleFunc("/cardToken", core_middleware.MiddleWareErrorHandler(httpRouters.CreateCardToken))		
	createCardToken.Use(otelmux.Middleware("go-card"))

	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpServer.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpServer.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpServer.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpServer.IdleTimeout) * time.Second, 
	}

	childLogger.Info().Str("Service Port", strconv.Itoa(h.httpServer.Port)).Send()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Error().Err(err).Msg("canceling http mux server !!!")
		}
	}()

	// Get SIGNALS
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-ch

		switch sig {
		case syscall.SIGHUP:
			childLogger.Info().Msg("Received SIGHUP: reloading configuration...")
		case syscall.SIGINT, syscall.SIGTERM:
			childLogger.Info().Msg("Received SIGINT/SIGTERM termination signal. Exiting")
			return
		default:
			childLogger.Info().Interface("Received signal:", sig).Send()
		}
	}

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("warning dirty shutdown !!!")
		return
	}
}