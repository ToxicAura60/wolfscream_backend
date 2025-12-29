package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wolfscream/database"
	"wolfscream/validator"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)

// --------------------
// Create Table
// --------------------
func CreateTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type CreateTableBody struct {
		Name        string `json:"name" validate:"required,snakecase,min=1"`
		Description string `json:"description"`
	}

	var body CreateTableBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if err := validator.Validate.Struct(body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"errors": validator.FormatError(err),
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

	if _, err := tx.Exec(`INSERT INTO "user_defined_table" ("name", "description") VALUES ($1, $2);`, body.Name, body.Description); err != nil {
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

	query := "SELECT id, name, description FROM user_defined_table ORDER BY created_at ASC;"

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

	type Table struct {
		Id          int     `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	tables := []Table{}

	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Id, &table.Name, &table.Description); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan table: %v", err),
			})
			return
		}
		tables = append(tables, table)
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
		"status": "success",
		"data":   tables,
	})
}

// --------------------
// List Tables End
// --------------------

// --------------------
// Update Table
// --------------------
func UpdateTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tableName := chi.URLParam(r, "table-name")

	type UpdateTableBody struct {
		Name        *string `json:"name" validate:"omitempty,min=1"`
		Description *string `json:"description"`
	}

	var body UpdateTableBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if err := validator.Validate.Struct(body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"errors": validator.FormatError(err),
		})
		return
	}

	if body.Name == nil && body.Description == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "No changes applied",
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

	if body.Name != nil && *body.Name != tableName {
		if _, err := tx.Exec(fmt.Sprintf(`ALTER TABLE %s RENAME TO %s;`, pq.QuoteIdentifier(tableName), pq.QuoteIdentifier(*body.Name))); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to create table: %v", err),
			})
			return
		}
	}

	if _, err := tx.Exec(`UPDATE user_defined_table SET name = COALESCE($1, name), description = COALESCE($2, description) WHERE name = $3;`, body.Name, body.Description, tableName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to update table: %v", err),
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
		"message": "Table updated successfully",
	})
}

// --------------------
// Update Table End
// --------------------

// --------------------
// Drop Table
// --------------------
func DropTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tableName := chi.URLParam(r, "table-name")

	err := validator.Validate.Var(tableName, "required,snakecase")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"errors": validator.FormatError(err),
		})
		return
	}

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

	dropQuery := fmt.Sprintf("DROP TABLE %s", tableName)
	if _, err := tx.Exec(dropQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to drop table: %v", err),
		})
		return
	}

	if _, err := tx.Exec("DELETE FROM user_defined_table WHERE name = $1", tableName); err != nil {
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
		"data":   results,
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
