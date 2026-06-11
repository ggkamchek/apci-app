package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/black/apci-app/security-center/internal/graph"
	"github.com/black/apci-app/security-center/internal/ingest"
)

type Handler struct {
	service    *ingest.Service
	adminToken string
}

func NewHandler(service *ingest.Service, adminToken string) *Handler {
	return &Handler{service: service, adminToken: adminToken}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/admin")
	path = strings.TrimPrefix(path, "/")

	switch {
	case path == "links" && r.Method == http.MethodGet:
		h.links(w, r)
	case path == "stats" && r.Method == http.MethodGet:
		h.stats(w, r)
	case path == "stats/top-clones" && r.Method == http.MethodGet:
		h.topClones(w, r)
	case path == "stats/activity-series" && r.Method == http.MethodGet:
		h.activitySeries(w, r)
	case path == "devices" && r.Method == http.MethodGet:
		h.devices(w, r)
	case strings.HasPrefix(path, "account/") && r.Method == http.MethodGet:
		h.accountRoutes(w, r, strings.TrimPrefix(path, "account/"))
	case path == "diff" && r.Method == http.MethodGet:
		h.diff(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) links(w http.ResponseWriter, r *http.Request) {
	links, err := h.service.ListLinks(r.Context())
	if err != nil {
		log.Printf("admin list links: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if links == nil {
		links = []graph.Link{}
	}
	writeJSON(w, links)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		log.Printf("admin stats: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

func (h *Handler) topClones(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 10)
	clones, err := h.service.TopDeviceClones(r.Context(), limit)
	if err != nil {
		log.Printf("admin top clones: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if clones == nil {
		clones = []graph.DeviceClone{}
	}
	writeJSON(w, clones)
}

func (h *Handler) activitySeries(w http.ResponseWriter, r *http.Request) {
	hours := queryInt(r, "hours", 24)
	series, err := h.service.ActivitySeries(r.Context(), hours)
	if err != nil {
		log.Printf("admin activity series: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if series == nil {
		series = []graph.ActivityPoint{}
	}
	writeJSON(w, series)
}

func (h *Handler) devices(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "clone"
	}
	limit := queryInt(r, "limit", 100)
	items, err := h.service.ListDeviceRegistry(r.Context(), filter, limit)
	if err != nil {
		log.Printf("admin devices: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []graph.DeviceRegistryItem{}
	}
	writeJSON(w, map[string]any{
		"filter": filter,
		"items":  items,
	})
}

func (h *Handler) accountRoutes(w http.ResponseWriter, r *http.Request, rest string) {
	rest = strings.TrimSuffix(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	userID := parts[0]
	if !isLookupAccountID(userID) {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if len(parts) == 1 {
		profile, err := h.service.GetAccount(r.Context(), userID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, profile)
		return
	}

	if len(parts) == 2 && parts[1] == "related" {
		result, err := h.service.GetRelatedAccounts(r.Context(), userID)
		if err != nil {
			log.Printf("admin related: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if result.Related == nil {
			result.Related = []graph.RelatedAccount{}
		}
		writeJSON(w, result)
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) diff(w http.ResponseWriter, r *http.Request) {
	userA := strings.TrimSpace(r.URL.Query().Get("a"))
	userB := strings.TrimSpace(r.URL.Query().Get("b"))
	if userA == "" || userB == "" {
		http.Error(w, "a and b query params required", http.StatusBadRequest)
		return
	}
	if !isLookupAccountID(userA) {
		http.Error(w, "invalid user id a", http.StatusBadRequest)
		return
	}
	if !isLookupAccountID(userB) {
		http.Error(w, "invalid user id b", http.StatusBadRequest)
		return
	}

	result, err := h.service.DiffAccounts(r.Context(), userA, userB)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, result)
}

func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.authorize(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "q is required", http.StatusBadRequest)
		return
	}

	if isLookupAccountID(query) {
		profile, err := h.service.GetAccount(r.Context(), query)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, profile)
		return
	}

	result, err := h.service.FindAccountsByDeviceQuery(r.Context(), query)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if result.Accounts == nil {
		result.Accounts = []graph.AccountProfile{}
	}
	writeJSON(w, result)
}

func (h *Handler) authorize(r *http.Request) bool {
	if h.adminToken == "" {
		return true
	}

	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	return strings.TrimPrefix(auth, prefix) == h.adminToken
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("admin encode: %v", err)
	}
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func (h *Handler) Links(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}
