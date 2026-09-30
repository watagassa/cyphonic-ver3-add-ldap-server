package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AuthenticationService      AuthenticationService
	SelfDesiredFQDN            []byte
	ServiceType                string
	ErrorLogFilePath           string
	DebugLogFilePath           string
	TunInterfaceName           string
	TunDnsInterfaceName        string
	TunDnsInterfaceIPv4        string
	TunDnsInterfaceIPv6        string
	APIPort                    string
	NodePort                   string
	VirtualIPType              int
	DebugMode                  bool
	RouteOptimizationMode      bool
	KeepAliveIntervalSeconds   int
	KeepAliveACKTimeoutSeconds int
}

type AuthenticationService struct {
	Host   string
	Port   string
	CaFile string
}

var (
	ErrUnsupportedFieldType = errors.New("unsupported field type")
	ErrMissingEnvVars       = errors.New("missing required environment variables")
)

func Get() (*Config, error) {
	var (
		cfg                        Config
		missed                     []string
		debugMode                  string
		routeOptimizationMode      string
		virtualIPType              string
		keepAliveIntervalSeconds   string
		keepAliveACKTimeoutSeconds string
	)

	for _, prop := range []struct {
		name     string
		field    any
		optional bool
	}{
		{name: "AUTHENTICATION_SERVICE_HOST", field: &cfg.AuthenticationService.Host, optional: false},
		{name: "AUTHENTICATION_SERVICE_PORT", field: &cfg.AuthenticationService.Port, optional: false},
		{name: "AUTHENTICATION_SERVICE_CA_FILE_PATH", field: &cfg.AuthenticationService.CaFile, optional: false},
		{name: "SELF_DESIRED_FQDN", field: &cfg.SelfDesiredFQDN, optional: true},
		{name: "DEBUG_MODE", field: &debugMode, optional: false},
		{name: "ROUTE_OPTIMIZATION_MODE", field: &routeOptimizationMode, optional: false},
		{name: "SERVICE_TYPE", field: &cfg.ServiceType, optional: false},
		{name: "ERROR_LOG_FILE_PATH", field: &cfg.ErrorLogFilePath, optional: false},
		{name: "DEBUG_LOG_FILE_PATH", field: &cfg.DebugLogFilePath, optional: false},
		{name: "TUN_INTERFACE_NAME", field: &cfg.TunInterfaceName, optional: false},
		{name: "TUN_DNS_INTERFACE_NAME", field: &cfg.TunDnsInterfaceName, optional: false},
		{name: "VIRTUAL_IP_TYPE", field: &virtualIPType, optional: false},
		{name: "TUN_DNS_INTERFACE_IPv4", field: &cfg.TunDnsInterfaceIPv4, optional: false},
		{name: "TUN_DNS_INTERFACE_IPv6", field: &cfg.TunDnsInterfaceIPv6, optional: false},
		{name: "API_PORT", field: &cfg.APIPort, optional: false},
		{name: "NODE_PORT", field: &cfg.NodePort, optional: false},
		{name: "KEEP_ALIVE_INTERVAL_SECONDS", field: &keepAliveIntervalSeconds, optional: false},
		{name: "KEEP_ALIVE_ACK_TIMEOUT_SECONDS", field: &keepAliveACKTimeoutSeconds, optional: false},
	} {
		val := os.Getenv(prop.name)
		if val == "" {
			if !prop.optional {
				missed = append(missed, prop.name)
			}

			continue
		}

		switch field := prop.field.(type) {
		case *string:
			*field = val
		case *[]byte:
			*field = []byte(val)
		default:
			return nil, fmt.Errorf("%w: %T", ErrUnsupportedFieldType, field)
		}
	}

	if len(missed) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrMissingEnvVars, missed)
	}

	isDebugMode, err := strconv.ParseBool(debugMode)
	if err != nil {
		return nil, fmt.Errorf("parse error debugMode: %w", err)
	}
	cfg.DebugMode = isDebugMode

	isRouteOptimizationMode, err := strconv.ParseBool(routeOptimizationMode)
	if err != nil {
		return nil, fmt.Errorf("parse error routeOptimizationMode: %w", err)
	}
	cfg.RouteOptimizationMode = isRouteOptimizationMode

	virtualIPTypeInt, err := strconv.Atoi(virtualIPType)
	if err != nil {
		return nil, fmt.Errorf("parse error virtualIPType: %w", err)
	}
	cfg.VirtualIPType = virtualIPTypeInt

	keepAliveIntervalSecondsInt, err := strconv.Atoi(keepAliveIntervalSeconds)
	if err != nil {
		return nil, fmt.Errorf("parse error keepAliveIntervalSeconds: %w", err)
	}
	cfg.KeepAliveIntervalSeconds = keepAliveIntervalSecondsInt

	keepAliveACKTimeoutSecondsInt, err := strconv.Atoi(keepAliveACKTimeoutSeconds)
	if err != nil {
		return nil, fmt.Errorf("parse error keepAliveACKTimeoutSeconds: %w", err)
	}
	cfg.KeepAliveACKTimeoutSeconds = keepAliveACKTimeoutSecondsInt

	return &cfg, nil
}
