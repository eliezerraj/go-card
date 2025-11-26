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
	go_core_midleware "github.com/eliezerraj/go-core/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	// Trace
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	 sdktrace "go.opentelemetry.io/otel/sdk/trace"
	 
	// Metrics
	"runtime/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	childLogger 	= log.With().Str("component","go-card").Str("package","internal.infra.server").Logger()
	core_middleware go_core_midleware.ToolsMiddleware
	tracerProvider	go_core_observ.TracerProvider
	infoTrace 		go_core_observ.InfoTrace
	tracer			trace.Tracer
)

type HttpServer struct {
	httpServer	*model.Server
}

// About create new http server
func NewHttpAppServer(httpServer *model.Server) HttpServer {
	childLogger.Info().Str("func","NewHttpAppServer").Send()
	return HttpServer{httpServer: httpServer }
}

// About initialize MeterProvider with Prometheus exporter
func initMeterProvider(ctx context.Context, serviceName string) (*sdkmetric.MeterProvider, error) {
	childLogger.Info().Str("func","initMeterProvider").Send()

	// 1. Configurar o Recurso OTel
	res, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", "production"),
		),
	)
	if err != nil {
		return nil, err
	}

	// 2. Criar o Prometheus Exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// 3. Criar o MeterProvider, usando o Prometheus Exporter como Reader.
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	return provider, nil
}

// About setup Go runtime metrics
func setupGoRuntimeMetrics(meter metric.Meter) {
	// Padrões de métricas do runtime do Go que queremos coletar.
	// O go_memstats_heap_alloc_bytes e o go_cpu_seconds_total são ótimos para começar.
	childLogger.Info().Str("func","setupGoRuntimeMetrics").Send()

	patterns := []string{
		"go.cpu.seconds_total",
		"go.memory.heap.alloc_bytes",
		"go.memory.gc.cpu.seconds_total",
	}

	for _, pattern := range patterns {
		_, err := meter.Int64ObservableGauge(
			pattern,
			metric.WithDescription("Métrica de runtime do Go"),
			metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
				// Usamos o pacote runtime/metrics para obter o valor.
				sample := make([]metrics.Sample, 1)
				sample[0].Name = pattern
				metrics.Read(sample)

				if sample[0].Value.Kind() == metrics.KindUint64 {
					// O valor do runtime/metrics é um contador/gauge, transformamos em int64 para o OTel.
					observer.Observe(int64(sample[0].Value.Uint64()))
				}
				return nil
			}),
		)
		if err != nil {
			childLogger.Error().Err(err).Msg("erro setupGoRuntimeMetrics")
			log.Printf("Erro ao configurar métrica %s: %v", pattern, err)
		}
	}
}

// About start http server
func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpRouters *api.HttpRouters,
										appServer *model.AppServer) {
	childLogger.Info().Str("func","StartHttpAppServer").Send()
			
	// --------- OTEL traces ---------------
	var initTracerProvider *sdktrace.TracerProvider
	
	if appServer.InfoPod.OtelTraces {
		infoTrace.PodName = appServer.InfoPod.PodName
		infoTrace.PodVersion = appServer.InfoPod.ApiVersion
		infoTrace.ServiceType = "k8-workload"
		infoTrace.Env = appServer.InfoPod.Env
		infoTrace.AccountID = appServer.InfoPod.AccountID

		initTracerProvider = tracerProvider.NewTracerProvider(	ctx, 
																appServer.ConfigOTEL, 
																&infoTrace)

		otel.SetTextMapPropagator(propagation.TraceContext{})
		otel.SetTracerProvider(initTracerProvider)
		tracer = initTracerProvider.Tracer(appServer.InfoPod.PodName)
	}

	// --------- OTEL metrics ---------------
	var meterProvider *sdkmetric.MeterProvider
	
	if appServer.InfoPod.OtelMetrics {
		meterProvider, err := initMeterProvider(ctx, infoTrace.PodName)
		if err != nil {
			childLogger.Error().Err(err).Msg("Error start Otel Metrics Provider")
		} else {
			meter := meterProvider.Meter(infoTrace.PodName)
			setupGoRuntimeMetrics(meter)
			childLogger.Info().Msg("Otel Metrics Provider started SUCCESSFULL")
		}
	}

	defer func() {
		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				childLogger.Error().Err(err).Msg("failed to stop instrumentation")
			}
		}

		if initTracerProvider != nil {
			err := initTracerProvider.Shutdown(ctx)
			if err != nil{
				childLogger.Error().Err(err).Send()
			}
		}
		childLogger.Info().Msg("stop done !!!")
	}()
	
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(core_middleware.MiddleWareHandlerHeader)

	myRouter.Handle("/metrics", promhttp.Handler())

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