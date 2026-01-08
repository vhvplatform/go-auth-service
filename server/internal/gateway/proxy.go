package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Proxy handles reverse proxying to microservices
type Proxy struct {
	// Map of service names to their URLs
	services map[string]string
}

// NewProxy creates a new gateway proxy
func NewProxy() *Proxy {
	return &Proxy{
		services: make(map[string]string),
	}
}

// AddService adds a service to the proxy
func (p *Proxy) AddService(name, targetURL string) {
	p.services[name] = targetURL
}

// ServeHTTP handles the proxying logic
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, tenantID, internalToken string) {
	path := r.URL.Path
	var targetURL string

	// Routing rules
	if strings.HasPrefix(path, "/api/") {
		// /api/service-name/api-path
		parts := strings.SplitN(strings.TrimPrefix(path, "/api/"), "/", 2)
		if len(parts) > 0 {
			serviceName := parts[0]
			if url, ok := p.services[serviceName]; ok {
				targetURL = url
				// Rewrite path: /api/service-name/path -> /path
				if len(parts) > 1 {
					r.URL.Path = "/" + parts[1]
				} else {
					r.URL.Path = "/"
				}
			}
		}
	} else if strings.HasPrefix(path, "/page/") {
		// /page/service-name/page-path -> React page
		parts := strings.SplitN(strings.TrimPrefix(path, "/page/"), "/", 2)
		if len(parts) > 0 {
			serviceName := parts[0] + "-frontend" // Convention for frontend services
			if url, ok := p.services[serviceName]; ok {
				targetURL = url
				if len(parts) > 1 {
					r.URL.Path = "/" + parts[1]
				} else {
					r.URL.Path = "/"
				}
			}
		}
	} else if strings.HasPrefix(path, "/upload/") {
		// /upload/file-key -> file-service
		if url, ok := p.services["file-service"]; ok {
			targetURL = url
			r.URL.Path = strings.TrimPrefix(path, "/upload")
		}
	} else {
		// Others handled as slug (e.g. to a CMS service or similar)
		if url, ok := p.services["slug-service"]; ok {
			targetURL = url
		}
	}

	if targetURL == "" {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Inject headers
	r.Header.Set("X-Tenant-ID", tenantID)
	r.Header.Set("Authorization", "Bearer "+internalToken)

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}
