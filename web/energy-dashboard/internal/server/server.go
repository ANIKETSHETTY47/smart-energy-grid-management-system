package server

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"energy-dashboard-go/internal/api"
)

type Server struct {
	mux      *http.ServeMux
	tmpl     *template.Template
	api      *api.Client
	facility string
}

func New() *Server {
	// Register template functions
	funcMap := template.FuncMap{
		"toJSON": toJSON,
	}

	// Parse main templates
	tmpl := template.Must(template.New("base").Funcs(funcMap).ParseGlob("templates/*.html"))

	// Parse partials if they exist
	if matches, _ := filepath.Glob("templates/partials/*.html"); len(matches) > 0 {
		tmpl = template.Must(tmpl.ParseFiles(matches...))
	} else {
		log.Println("⚠️  No partial templates found — continuing without them.")
	}

	// Default facility if not set
	facility := os.Getenv("FACILITY_ID")
	if facility == "" {
		facility = "facility-001"
	}

	s := &Server{
		mux:      http.NewServeMux(),
		tmpl:     tmpl,
		api:      api.New(),
		facility: facility,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.mux.HandleFunc("/healthz", s.handleHealthz)

	s.mux.HandleFunc("/", s.handleDashboard)
	s.mux.HandleFunc("/dashboard", s.handleDashboard)

	s.mux.HandleFunc("/alerts", s.handleAlerts)
	s.mux.HandleFunc("/alerts/acknowledge", s.handleAcknowledge)

	s.mux.HandleFunc("/analytics", s.handleAnalytics)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
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
	_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	readings, _ := s.api.RecentReadings(ctx, s.facility, 24)
	alerts, _ := s.api.Alerts(ctx, s.facility, "")

	data := map[string]any{
		"Title":        "Energy Grid Dashboard",
		"FacilityID":   s.facility,
		"ReadingsJSON": toJSON(readings),
		"Alerts":       alerts,
		"APIStatus":    s.status(ctx),
	}

	s.render(w, "dashboard.html", data)
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	severity := r.URL.Query().Get("severity")
	resp, _ := s.api.Alerts(ctx, s.facility, severity)

	data := map[string]any{
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

	var report any
	if r.Method == http.MethodPost {
		date := r.FormValue("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		res, err := s.api.GenerateAnalytics(ctx, s.facility, date)
		if err != nil {
			report = map[string]any{"Error": "Failed to generate report"}
		} else {
			report = res
		}
	}

	data := map[string]any{
		"Title":      "Analytics & Reports",
		"FacilityID": s.facility,
		"Today":      time.Now().Format("2006-01-02"),
		"Report":     report,
		"APIStatus":  s.status(ctx),
	}

	s.render(w, "analytics.html", data)
}

func (s *Server) status(ctx context.Context) string {
	if h, err := s.api.Health(ctx); err == nil && h != nil {
		return "online"
	}
	return "offline"
}

// Helper: convert to JSON for use in templates
func toJSON(v any) template.JS {
	b, _ := json.Marshal(v)
	return template.JS(b)
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Println("render error:", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
