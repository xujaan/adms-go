package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"adms-go/internal/handler"
	"adms-go/internal/middleware"
	"adms-go/internal/store"
	"adms-go/internal/webhook"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Port      string
	DBAdmsDSN string
	AdminUser string
	AdminPass string
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

	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}

	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = "admin123"
	}

	return Config{
		Port:      port,
		DBAdmsDSN: dbAdmsDSN,
		AdminUser: adminUser,
		AdminPass: adminPass,
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
	tmpl := template.New("").Funcs(template.FuncMap{
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
		"eq": func(a, b interface{}) bool { return a == b },
		"le": func(a, b int) bool { return a <= b },
		"ge": func(a, b int) bool { return a >= b },
	})

	for _, pattern := range []string{"templates/*.html"} {
		if _, err := tmpl.ParseGlob(pattern); err != nil {
			log.Fatalf("templates %s: %v", pattern, err)
		}
	}

	// Webhook dispatcher
	whDispatcher := webhook.NewDispatcher(db)

	// Handlers
	iclockH := &handler.IClockHandler{Store: db, Dispatcher: whDispatcher}
	dashH := handler.NewDashboardHandler(db, tmpl)
	whHandler := &handler.WebhookHandler{Store: db}
	authH := &handler.AuthHandler{User: cfg.AdminUser, Pass: cfg.AdminPass, Template: tmpl}

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)

	// IClock routes — no middleware, no auth (ZKTeco devices)
	r.Route("/iclock", func(r chi.Router) {
		r.Get("/cdata", iclockH.Handshake)
		r.Post("/cdata", iclockH.ReceiveRecords)
		r.Get("/test", iclockH.TestHandler)
		r.Get("/getrequest", iclockH.GetRequestHandler)
	})

	// Auth routes — public
	r.Get("/login", authH.LoginPage)
	r.Post("/login", authH.Login)
	r.Get("/logout", authH.Logout)

	// Protected dashboard routes
	authMw := middleware.BasicAuth(cfg.AdminUser, cfg.AdminPass)
	r.Group(func(r chi.Router) {
		r.Use(authMw)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/devices", http.StatusFound)
		})
		r.Get("/devices", dashH.Devices)
		r.Get("/devices-log", dashH.DeviceLog)
		r.Get("/finger-log", dashH.FingerLog)
		r.Get("/webhooks", dashH.Webhooks)
		r.Get("/attendance", dashH.Attendance)

		// Webhook API routes
		r.Route("/api/webhooks", func(r chi.Router) {
			r.Get("/", whHandler.ListWebhooks)
			r.Post("/", whHandler.CreateWebhook)
			r.Delete("/{id}", whHandler.DeleteWebhook)
		})
	})

	// Health check — public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("ADMS server starting on http://0.0.0.0%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
