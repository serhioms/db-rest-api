package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// DB connection
var db *sql.DB

// healthHandler handles the GET /health request
func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// selectHandler handles the GET /select/{from} request
func selectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	from := vars["from"]
	where := r.URL.Query().Get("where")

	// Dummy response for now
	var query string
	if where == "" {
		query = fmt.Sprintf("SELECT * FROM %s", from)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s WHERE %s", from, where)
	}

	fmt.Println("query:", query)

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		results = append(results, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// insertHandler handles the POST /insert/{from} request
func insertHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	from := vars["from"]

	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var columns, values []string
	for k := range data {
		columns = append(columns, k)
		values = append(values, fmt.Sprintf("'%v'", data[k]))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", from, string(Join(columns, ",")), string(Join(values, ",")))
	_, err = db.Exec(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Successfully inserted into %s", from)
}

func main() {
	// Database connection
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	fmt.Println("connStr:", connStr)

	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to database (attempt %d): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database after 10 attempts: %v", err)
	}
	defer db.Close()

	fmt.Println("Successfully connected to the database")

	// Create tables
	if err := createTables(db); err != nil {
		log.Fatalf("Could not create tables: %v", err)
	}

	// Router
	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/select/{from}", selectHandler).Methods("GET")
	r.HandleFunc("/insert/{from}", insertHandler).Methods("POST")

	// Start server
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createTables(db *sql.DB) error {
	// Check if table exists
	var exists bool
	checkQuery := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')"
	err := db.QueryRow(checkQuery).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Println("Table 'users' already exists, skipping creation.")
		return nil
	}

	log.Println("Creating table 'users'...")
	query := `
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		email VARCHAR(100) UNIQUE
	);`
	_, err = db.Exec(query)
	return err
}

func Join(s []string, sep string) []byte {
	var b []byte
	for i, x := range s {
		if i > 0 {
			b = append(b, sep...)
		}
		b = append(b, x...)
	}
	return b
}
