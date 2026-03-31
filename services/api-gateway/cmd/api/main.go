package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type gatewayConfig struct {
	port            string
	posAPIURL       string
	iamAPIURL       string
	crmAPIURL       string
	omsAPIURL       string
	scmAPIURL       string
	hrmAPIURL       string
	erpAPIURL       string
	mdmAPIURL       string
	dssAPIURL       string
	ecmAPIURL       string
	idpAPIURL       string
	jwtSecret       string
	jwtIssuer       string
	protectedRoutes []string
}

type tokenClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	cfg := loadConfig()

	posTarget := parseURLOrFatal("POS_API_URL", cfg.posAPIURL)
	iamTarget := parseURLOrFatal("IAM_API_URL", cfg.iamAPIURL)
	crmTarget := parseURLOrFatal("CRM_API_URL", cfg.crmAPIURL)
	omsTarget := parseURLOrFatal("OMS_API_URL", cfg.omsAPIURL)
	scmTarget := parseURLOrFatal("SCM_API_URL", cfg.scmAPIURL)
	hrmTarget := parseURLOrFatal("HRM_API_URL", cfg.hrmAPIURL)
	erpTarget := parseURLOrFatal("ERP_API_URL", cfg.erpAPIURL)
	mdmTarget := parseURLOrFatal("MDM_API_URL", cfg.mdmAPIURL)
	dssTarget := parseURLOrFatal("DSS_API_URL", cfg.dssAPIURL)
	ecmTarget := parseURLOrFatal("ECM_API_URL", cfg.ecmAPIURL)
	idpTarget := parseURLOrFatal("IDP_API_URL", cfg.idpAPIURL)

	posProxy := reverseProxy(posTarget, "pos-api")
	iamProxy := reverseProxy(iamTarget, "iam-api")
	crmProxy := reverseProxy(crmTarget, "crm-api")
	omsProxy := reverseProxy(omsTarget, "oms-api")
	scmProxy := reverseProxy(scmTarget, "scm-api")
	hrmProxy := reverseProxy(hrmTarget, "hrm-api")
	erpProxy := reverseProxy(erpTarget, "erp-api")
	mdmProxy := reverseProxy(mdmTarget, "mdm-api")
	dssProxy := reverseProxy(dssTarget, "dss-api")
	ecmProxy := reverseProxy(ecmTarget, "ecm-api")
	idpProxy := reverseProxy(idpTarget, "idp-api")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", gatewayHealth)

	// Login is delegated to IAM service and does not require JWT.
	mux.Handle("/login", iamProxy)

	// CRM endpoints - protected by JWT
	registerProtectedPrefix(mux, "/api/customers", crmProxy, cfg)

	// OMS endpoints - protected by JWT
	registerProtectedPrefix(mux, "/api/orders", omsProxy, cfg)

	// SCM endpoints - protected by JWT
	registerProtectedPrefix(mux, "/scm", scmProxy, cfg)

	// HRM endpoints - protected by JWT
	registerProtectedPrefix(mux, "/hrm", hrmProxy, cfg)

	// ERP endpoints - protected by JWT
	registerProtectedPrefix(mux, "/erp", erpProxy, cfg)

	// MDM endpoints - protected by JWT
	registerProtectedPrefix(mux, "/mdm", mdmProxy, cfg)

	// DSS endpoints - protected by JWT
	registerProtectedPrefix(mux, "/dss", dssProxy, cfg)

	// ECM endpoints - protected by JWT
	registerProtectedPrefix(mux, "/ecm", ecmProxy, cfg)

	// IDP endpoints - protected by JWT
	registerProtectedPrefix(mux, "/idp", idpProxy, cfg)

	// POS endpoints - protected by JWT for specific routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if isProtectedRoute(r.URL.Path, cfg.protectedRoutes) {
			tokenString, ok := extractBearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"error":   "missing or invalid Authorization header",
				})
				return
			}

			if err := validateJWT(tokenString, cfg.jwtSecret, cfg.jwtIssuer); err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"error":   "invalid token",
				})
				return
			}
		}

		posProxy.ServeHTTP(w, r)
	})

	log.Printf("🚀 API Gateway starting on :%s", cfg.port)
	log.Printf("-> POS target: %s", cfg.posAPIURL)
	log.Printf("-> IAM target: %s", cfg.iamAPIURL)
	log.Printf("-> CRM target: %s", cfg.crmAPIURL)
	log.Printf("-> OMS target: %s", cfg.omsAPIURL)
	log.Printf("-> SCM target: %s", cfg.scmAPIURL)
	log.Printf("-> HRM target: %s", cfg.hrmAPIURL)
	log.Printf("-> ERP target: %s", cfg.erpAPIURL)
	log.Printf("-> MDM target: %s", cfg.mdmAPIURL)
	log.Printf("-> DSS target: %s", cfg.dssAPIURL)
	log.Printf("-> ECM target: %s", cfg.ecmAPIURL)
	log.Printf("-> IDP target: %s", cfg.idpAPIURL)
	log.Printf("-> Protected routes: %s", strings.Join(cfg.protectedRoutes, ","))

	if err := http.ListenAndServe(":"+cfg.port, mux); err != nil {
		log.Fatalf("gateway failed to start: %v", err)
	}
}

func loadConfig() gatewayConfig {
	protected := strings.Split(getEnv("PROTECTED_PATHS", "/api/orders,/api/customers,/scm,/hrm,/erp,/mdm,/dss,/ecm,/idp"), ",")
	cleaned := make([]string, 0, len(protected))
	for _, p := range protected {
		value := strings.TrimSpace(p)
		if value == "" {
			continue
		}
		if !strings.HasPrefix(value, "/") {
			value = "/" + value
		}
		cleaned = append(cleaned, value)
	}
	if len(cleaned) == 0 {
		cleaned = []string{"/api/orders", "/api/customers", "/scm", "/hrm", "/erp", "/mdm", "/dss", "/ecm", "/idp"}
	}

	return gatewayConfig{
		port:            getEnv("GATEWAY_PORT", "8080"),
		posAPIURL:       getEnv("POS_API_URL", "http://localhost:3000"),
		iamAPIURL:       getEnv("IAM_API_URL", "http://localhost:4001"),
		crmAPIURL:       getEnv("CRM_API_URL", "http://localhost:4002"),
		omsAPIURL:       getEnv("OMS_API_URL", "http://localhost:4003"),
		scmAPIURL:       getEnv("SCM_API_URL", "http://localhost:4004"),
		hrmAPIURL:       getEnv("HRM_API_URL", "http://localhost:4006"),
		erpAPIURL:       getEnv("ERP_API_URL", "http://localhost:4007"),
		mdmAPIURL:       getEnv("MDM_API_URL", "http://localhost:4008"),
		dssAPIURL:       getEnv("DSS_API_URL", "http://localhost:4009"),
		ecmAPIURL:       getEnv("ECM_API_URL", "http://localhost:4010"),
		idpAPIURL:       getEnv("IDP_API_URL", "http://localhost:4011"),
		jwtSecret:       getEnv("JWT_SECRET", "dev-secret-change-me"),
		jwtIssuer:       getEnv("JWT_ISSUER", "iam-api"),
		protectedRoutes: cleaned,
	}
}

func reverseProxy(target *url.URL, serviceName string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director

	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		r.Host = target.Host
		r.Header.Set("X-Forwarded-Host", r.Host)
		r.Header.Set("X-Forwarded-Proto", "http")
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error (%s): %v", serviceName, err)
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("%s unavailable", serviceName),
		})
	}

	return proxy
}

func parseURLOrFatal(name, value string) *url.URL {
	target, err := url.Parse(value)
	if err != nil {
		log.Fatalf("invalid %s: %v", name, err)
	}
	return target
}

func registerProtectedPrefix(mux *http.ServeMux, prefix string, proxy *httputil.ReverseProxy, cfg gatewayConfig) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		handleProtectedRoute(w, r, proxy, cfg)
	})
	mux.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) {
		handleProtectedRoute(w, r, proxy, cfg)
	})
}

func isProtectedRoute(path string, protectedRoutes []string) bool {
	for _, protected := range protectedRoutes {
		if strings.HasPrefix(path, protected) {
			return true
		}
	}
	return false
}

// handleProtectedRoute validates JWT token and forwards request to the target proxy
func handleProtectedRoute(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy, cfg gatewayConfig) {
	tokenString, ok := extractBearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "missing or invalid Authorization header",
		})
		return
	}

	if err := validateJWT(tokenString, cfg.jwtSecret, cfg.jwtIssuer); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "invalid token",
		})
		return
	}

	// Forward to the target service
	proxy.ServeHTTP(w, r)
}

func extractBearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	if strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func validateJWT(tokenString, secret, expectedIssuer string) error {
	claims := &tokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return []byte(secret), nil
		},
	)
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("token is invalid")
	}
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return fmt.Errorf("unexpected issuer: %s", claims.Issuer)
	}

	return nil
}

func gatewayHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "api-gateway",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
