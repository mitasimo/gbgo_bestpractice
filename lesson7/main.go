package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"time"
)

func main() {
	loadConfig()

	// connect to db
	db, err := sql.Open("postgres", dbConnectSting())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create http server
	server := http.Server{
		Addr:        svcPort,
		Handler:     &Service{db},
		ReadTimeout: 10 * time.Second,
		//WriteTimeout: 10 * time.Second,
	}

	// os signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// process os signals
	go func() {
		<-sigs
		//Error return value of `server.Shutdown` is not checked (errcheck)
		//        server.Shutdown(context.Background())
		err := server.Shutdown(context.Background())
		if err != nil {
			os.Exit(-1)
		}
	}()

	log.Println("service started...")
	err = server.ListenAndServe() // start server
	if err != nil {
		log.Println("service stopped with error: ", err)
	} else {
		log.Println("service stoped")
	}
}

// Service -
type Service struct {
	*sql.DB
}

func (svc *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// check auth
	usr, pass, ok := r.BasicAuth()
	if !ok {
		http.Error(w, errAuth, http.StatusUnauthorized)
		return
	}
	if usr != svcUser || pass != svcPassword {
		http.Error(w, errAuth, http.StatusUnauthorized)
		return
	}

	// check http method
	if r.Method != http.MethodPost {
		http.Error(w, "Only post method", http.StatusBadRequest)
		return
	}

	// read sql query from http request's body
	sqlQuery, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// execute http query
	res, err := svc.ExecContext(r.Context(), string(sqlQuery))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ra, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Rows affected: %d", ra)
	w.WriteHeader(http.StatusOK)
}
