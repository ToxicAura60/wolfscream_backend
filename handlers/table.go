package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wolfscream/database"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)

// --------------------
// Create Table
// --------------------
type CreateTableBody struct {
	Name string `json:"name"`
	Description string `json:"description"`
}

func CreateTable(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
	var body CreateTableBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`CREATE TABLE %s ();`, pq.QuoteIdentifier(body.Name))

	if _, err := tx.Exec(query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create table: %v", err),
		})
		return
	}

	if _, err := tx.Exec(`INSERT INTO "tables" ("name", "description") VALUES ($1, $2);`, body.Name, body.Description); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to insert table: %v", err),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to commit transaction",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Table added successfully",
	})
}
// --------------------
// Create Table End
// --------------------

// --------------------
// List Tables
// --------------------
func ListTables(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name, description FROM tables;"

	rows, err := database.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query columns: %v", err),
		})
		return
	}
	defer rows.Close()

	type Data struct {
		Id          int       `json:"id"`
		Name        string    `json:"name"`
		Description *string   `json:"description"`
	}

	data := []Data{}

	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.Id, &d.Name, &d.Description); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan column: %v", err),
			})
			return
		}
		data = append(data, d)
	}

	 if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to read rows: %v", err),
		})
    return
  }

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"data": data,
	})
}
// --------------------
// List Tables End
// --------------------

// --------------------
// Drop Table
// --------------------
func DropTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	table := chi.URLParam(r, "table")

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to start transaction: %v", err),
		})
		return
	}

	defer tx.Rollback()

	dropQuery := fmt.Sprintf("DROP TABLE %s", pq.QuoteIdentifier(table))
	if _, err := tx.Exec(dropQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to drop table: %v", err),
		})
		return
	}

	deleteQuery := `DELETE FROM tables WHERE name = $1`
	if _, err := tx.Exec(deleteQuery, table); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to delete metadata: %v", err),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Table dropped successfully",
	})
}
// --------------------
// Drop Table End
// --------------------

// --------------------
// Get Data
// --------------------
func GetData(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  table := chi.URLParam(r, "table")

  rows, err := database.DB.Query(fmt.Sprintf("SELECT * FROM %s;", pq.QuoteIdentifier(table)))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]string{
      "status":  "error",
      "message": fmt.Sprintf("Failed to fetch data from table '%s': %v", table, err),
    })
    return
  }
	defer rows.Close()

  columns, err := rows.Columns()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]string{
      "status":  "error",
      "message": fmt.Sprintf("Unable to read table structure: %v", err),
    })
    return
  }

  results := []map[string]any{}

  for rows.Next() {
    values := make([]any, len(columns))
    valuePtrs := make([]any, len(columns))

    for i := range values {
      valuePtrs[i] = &values[i]
    }

		if err := rows.Scan(valuePtrs...); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan: %v", err),
			})
			return
		}

		rowMap := map[string]any{}
		for i, col := range columns {
			rowMap[col] = values[i]
		}

    results = append(results, rowMap)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]any{
        "status": "success",
        "data": results,
    })
}
// --------------------
// Get Data End
// --------------------


// --------------------
// Insert Data
// --------------------
func InsertData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	table := chi.URLParam(r, "table")

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid JSON",
		})
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Empty body",
		})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	columns := []string{}
	placeholders := []string{}
	values := []any{}
	i := 1

	for key, val := range body {
		columns = append(columns, key)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	
	if _, err := tx.Exec(query, values...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to commit transaction",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Data inserted",
	})
}

// --------------------
// Insert Data End
// --------------------


// --------------------
// Delete Data
// --------------------
func DeleteData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	columnId := chi.URLParam(r, "columnId")
	table := chi.URLParam(r, "table")

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", table)
	if _, err := database.DB.Exec(query, columnId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to delete data: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Data deleted successfully",
	})
}
// --------------------
// Delete Data End
// --------------------














