package server

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"energy-dashboard-go/internal/api"
	"energy-dashboard-go/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	mux       *http.ServeMux
	tmpl      *template.Template
	api       *api.Client
	facility  string
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	broadcast chan interface{}
}

func New() *Server {
	funcMap := template.FuncMap{
		"toJSON": toJSON,
		"formatTime": func(ts int64) string {
			return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
		},
	}

	tmpl := template.Must(template.New("base").Funcs(funcMap).ParseGlob("templates/*.html"))

	if matches, _ := filepath.Glob("templates/partials/*.html"); len(matches) > 0 {
		tmpl = template.Must(tmpl.ParseFiles(matches...))
	}

	facility := os.Getenv("FACILITY_ID")
	if facility == "" {
		facility = "facility-001"
	}

	s := &Server{
		mux:       http.NewServeMux(),
		tmpl:      tmpl,
		api:       api.New(),
		facility:  facility,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan interface{}, 256),
	}

	s.routes()
	go s.handleBroadcast()
	go s.periodicUpdate()

	return s
}

func (s *Server) routes() {
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.HandleFunc("/", s.handleDashboard)
	s.mux.HandleFunc("/dashboard", s.handleDashboard)
	s.mux.HandleFunc("/alerts", s.handleAlerts)
	s.mux.HandleFunc("/alerts/acknowledge", s.handleAcknowledge)
	s.mux.HandleFunc("/analytics", s.handleAnalytics)
	s.mux.HandleFunc("/equipment", s.handleEquipment)
	s.mux.HandleFunc("/api/stats", s.handleAPIStats)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()

	ctx := context.Background()
	stats, _ := s.getStats(ctx)
	conn.WriteJSON(map[string]interface{}{
		"type": "init",
		"data": stats,
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *Server) handleBroadcast() {
	for msg := range s.broadcast {
		s.clientsMu.RLock()
		for conn := range s.clients {
			if err := conn.WriteJSON(msg); err != nil {
				conn.Close()
				delete(s.clients, conn)
			}
		}
		s.clientsMu.RUnlock()
	}
}

func (s *Server) periodicUpdate() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		stats, err := s.getStats(ctx)
		if err != nil {
			continue
		}

		s.broadcast <- map[string]interface{}{
			"type": "update",
			"data": stats,
		}
	}
}

func (s *Server) getStats(ctx context.Context) (map[string]interface{}, error) {
	readings, _ := s.api.RecentReadings(ctx, s.facility, 24)
	alerts, _ := s.api.Alerts(ctx, s.facility, "")

	stats := map[string]interface{}{
		"readings":  readings,
		"alerts":    alerts,
		"timestamp": time.Now().Unix(),
	}

	return stats, nil
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health, err := s.api.Health(ctx)
	status := "offline"
	if err == nil && health != nil {
		status = "online"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	readings, _ := s.api.RecentReadings(ctx, s.facility, 24)
	alerts, _ := s.api.Alerts(ctx, s.facility, "")

	data := map[string]interface{}{
		"Title":        "Energy Grid Dashboard",
		"FacilityID":   s.facility,
		"ReadingsJSON": toJSON(readings),
		"Alerts":       alerts,
		"APIStatus":    s.status(ctx),
	}

	s.render(w, "dashboard.html", data)
}

func (s *Server) handleEquipment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	equipment := []models.Equipment{
		{ID: "eq-001", Type: "Transformer", Status: "operational", Health: 95.5},
		{ID: "eq-002", Type: "Generator", Status: "operational", Health: 88.2},
		{ID: "eq-003", Type: "Meter", Status: "warning", Health: 72.8},
		{ID: "eq-004", Type: "Switch", Status: "operational", Health: 98.1},
	}

	data := map[string]interface{}{
		"Title":      "Equipment Monitoring",
		"FacilityID": s.facility,
		"Equipment":  equipment,
		"APIStatus":  s.status(ctx),
	}

	s.render(w, "equipment.html", data)
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	severity := r.URL.Query().Get("severity")
	resp, _ := s.api.Alerts(ctx, s.facility, severity)

	data := map[string]interface{}{
		"Title":      "System Alerts",
		"FacilityID": s.facility,
		"Severity":   severity,
		"Alerts":     resp,
		"APIStatus":  s.status(ctx),
	}

	s.render(w, "alerts.html", data)
}

func (s *Server) handleAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := s.api.AcknowledgeAlert(ctx, id); err != nil {
		log.Println("ack error:", err)
		http.Redirect(w, r, "/alerts?ack=fail", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/alerts?ack=ok", http.StatusSeeOther)
}

func (s *Server) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	var report interface{}
	if r.Method == http.MethodPost {
		date := r.FormValue("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		res, err := s.api.GenerateAnalytics(ctx, s.facility, date)
		if err != nil {
			report = map[string]interface{}{"Error": "Failed to generate report"}
		} else {
			report = res
		}
	}

	data := map[string]interface{}{
		"Title":      "Analytics & Reports",
		"FacilityID": s.facility,
		"Today":      time.Now().Format("2006-01-02"),
		"Report":     report,
		"APIStatus":  s.status(ctx),
	}

	s.render(w, "analytics.html", data)
}

func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats, err := s.getStats(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) status(ctx context.Context) string {
	if h, err := s.api.Health(ctx); err == nil && h != nil {
		return "online"
	}
	return "offline"
}

func toJSON(v interface{}) template.JS {
	b, _ := json.Marshal(v)
	return template.JS(b)
}

func (s *Server) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Println("render error:", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
