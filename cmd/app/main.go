package main

import (
	"log"
	"net/http"

	"avito-shop/internal/config"
	"avito-shop/internal/handler"
	"avito-shop/internal/repository"
	"avito-shop/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, cfg)

	r := mux.NewRouter()

	h := handler.NewHandler(svc, cfg)

	authRouter := r.PathPrefix("/api").Subrouter()
	authRouter.HandleFunc("/auth", h.Auth).Methods("POST")

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(handler.JwtMiddleware(cfg.JWTSecret)) 
	apiRouter.HandleFunc("/info", h.GetInfo).Methods("GET")
	apiRouter.HandleFunc("/sendCoin", h.SendCoin).Methods("POST")
	apiRouter.HandleFunc("/buy/{item}", h.BuyItem).Methods("GET")

	log.Printf("Server starting at :%d\n", cfg.AppPort)
	if err := http.ListenAndServe(cfg.Address(), r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
