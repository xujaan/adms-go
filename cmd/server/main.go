package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"adms-go/internal/handler"
	"adms-go/internal/store"
	"adms-go/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Port      string
	DBAdmsDSN string
}

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	dbAdmsDSN := os.Getenv("DB_ADMS_DSN")
	if dbAdmsDSN == "" {
		dbAdmsDSN = "root:@tcp(127.0.0.1:3306)/java_adms?parseTime=true"
	}

	return Config{
		Port:      port,
		DBAdmsDSN: dbAdmsDSN,
	}
}

func main() {
	cfg := loadConfig()

	db, err := store.New(cfg.DBAdmsDSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Templates
	tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"or": func(a, b interface{}) interface{} {
				if a == nil || a == "" {
					return b
				}
				return a
			},
			"until": func(n int) []int {
				out := make([]int, n)
				for i := range out {
					out[i] = i
				}
				return out
			},
			"add": func(a, b int) int { return a + b },
			"sub": func(a, b int) int {
				if a-b < 0 {
					return 0
				}
				return a - b
			},
			"eq": func(a, b int) bool { return a == b },
			"le": func(a, b int) bool { return a <= b },
			"ge": func(a, b int) bool { return a >= b },
		}).
		ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("templates: %v", err)
	}

	// Webhook dispatcher
	whDispatcher := webhook.NewDispatcher(db)

	// Handlers
	iclockH := &handler.IClockHandler{Store: db, Dispatcher: whDispatcher}
	dashH := handler.NewDashboardHandler(db, tmpl)
	whHandler := &handler.WebhookHandler{Store: db}

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// IClock routes — no middleware
	r.Route("/iclock", func(r chi.Router) {
		r.Get("/cdata", iclockH.Handshake)
		r.Post("/cdata", iclockH.ReceiveRecords)
		r.Get("/test", iclockH.TestHandler)
		r.Get("/getrequest", iclockH.GetRequestHandler)
	})

	// Web dashboard routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/devices", http.StatusFound)
	})
	r.Get("/devices", dashH.Devices)
	r.Get("/devices-log", dashH.DeviceLog)
	r.Get("/finger-log", dashH.FingerLog)
	r.Get("/attendance", dashH.Attendance)

	// Webhook API routes
	r.Route("/api/webhooks", func(r chi.Router) {
		r.Get("/", whHandler.ListWebhooks)
		r.Post("/", whHandler.CreateWebhook)
		r.Delete("/{id}", whHandler.DeleteWebhook)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("ADMS server starting on http://0.0.0.0%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
