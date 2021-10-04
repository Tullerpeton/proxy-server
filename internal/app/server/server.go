package server

import (
	"log"
	"net/http"

	"github.com/proxy-server/internal/pkg/request/handler"
	"github.com/proxy-server/internal/pkg/request/repository"
	"github.com/proxy-server/internal/pkg/request/usecase"
	"github.com/proxy-server/pkg/proxy"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Start() {
	// Connect to postgreSql db
	postgreSqlConn, err := sqlx.Open(
		"postgres",
		"user=proxy_user "+
			"password=proxy_user "+
			"dbname=proxy_db "+
			"host=localhost "+
			"port=5432 "+
			"sslmode=disable",
	)

	if err != nil {
		log.Fatal(err)
	}
	defer postgreSqlConn.Close()
	if err = postgreSqlConn.Ping(); err != nil {
		log.Fatal(err)
	}

	requestRepo := repository.NewRepository(postgreSqlConn)
	proxyManager := proxy.NewProxyManager(requestRepo)
	requestUseCase := usecase.NewProxyUseCase(requestRepo)
	requestHandler := handler.NewHandler(requestUseCase, proxyManager)

	serverProxy := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(requestHandler.ProxyRequest),
	}

	go func() {
		log.Fatal(serverProxy.ListenAndServe())
	}()

	router := mux.NewRouter()
	router.HandleFunc("/requests", requestHandler.GetRequests)
	router.HandleFunc("/requests/{id:[0-9]+}", requestHandler.GetRequest)
	router.HandleFunc("/scan/{id:[0-9]+}", requestHandler.ScanRequest)
	router.HandleFunc("/repeat/{id:[0-9]+}", requestHandler.RepeatRequest)

	serverRequest := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	log.Fatal(serverRequest.ListenAndServe())
}
