package model

import (
	"time"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

type AppServer struct {
	InfoPod 		*InfoPod 					`json:"info_pod"`
	Server     		*Server     				`json:"server"`
	ConfigOTEL		*go_core_observ.ConfigOTEL	`json:"otel_config"`
	DatabaseConfig	*go_core_pg.DatabaseConfig  `json:"database"`
	ApiService 		*[]ApiService				`json:"api_endpoints"`
}

type InfoPod struct {
	PodName				string 	`json:"pod_name"`
	ApiVersion			string 	`json:"version"`
	OSPID				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	AvailabilityZone 	string 	`json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	OtelTraces			bool   	`json:"otel_traces"`
	OtelMetrics			bool   	`json:"otel_metrics"`
	OtelLogs			bool   	`json:"otel_logs"`
	AccountID			string `json:"account_id,omitempty"`
}

type Server struct {
	Port 					int `json:"port"`
	ReadTimeout				int `json:"readTimeout"`
	WriteTimeout			int `json:"writeTimeout"`
	IdleTimeout				int `json:"idleTimeout"`
	CtxTimeout				int `json:"ctxTimeout"`
}

type ApiService struct {
	Name			string `json:"name_service"`
	Url				string `json:"url"`
	Method			string `json:"method"`
	XApigwApiId		string `json:"x-apigw-api-id,omitempty"`
	HostName		string `json:"host_name"`
	HttpTimeout		time.Duration `json:"httpTimeout"`
}

type MessageRouter struct {
	Message			string `json:"message"`
}

type Account struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
	TenantId		string  	`json:"tenant_id,omitempty"`
}

type Card struct {
	ID				int			`json:"id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`	
	CardNumber		string  	`json:"card_number,omitempty"`
	TokenData		string  	`json:"token_data,omitempty"`
	Holder			string  	`json:"holder,omitempty"`
	Type			string  	`json:"type,omitempty"`
	Model			string  	`json:"model,omitempty"`
	Atc				int			`json:"atc,omitempty"`
	Status			string  	`json:"status,omitempty"`
	ExpiredAt		time.Time 	`json:"expired_at,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
}