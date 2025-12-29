package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wolfscream/database"

	"github.com/go-chi/chi/v5"
)

// --------------------
// Add Column
// --------------------
func AddColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tableName := chi.URLParam(r, "table-name")

	type AddColumnBody struct {
		Name         string `json:"name" validate:"required,snakecase,min=1"`
		Type         string `json:"type" validate:"required"`
		Length       *int   `json:"length" validate:"omitempty,gte=1"`
		DefaultValue any    `json:"default"`
	}

	var body AddColumnBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	var tableId int

	err := database.DB.QueryRow("SELECT id FROM user_defined_table WHERE name = $1", tableName).Scan(&tableId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Table %s does not exist", tableName),
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query table: %v", err),
		})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	queryParts := []string{
		fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s`, tableName, body.Name),
	}

	switch body.Type {
	case "int4":
		queryParts = append(queryParts, "INTEGER")
	case "int8":
		queryParts = append(queryParts, "BIGINT")
	case "varchar":
		if body.Length != nil {
			queryParts = append(queryParts, fmt.Sprintf("VARCHAR(%d)", *body.Length))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Length is required for type varchar",
			})
			return
		}
	case "text":
		queryParts = append(queryParts, "TEXT")
	case "bool":
		queryParts = append(queryParts, "BOOLEAN")
	case "float4":
		queryParts = append(queryParts, "REAL")
	case "float8":
		queryParts = append(queryParts, "DOUBLE PRECISION")
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Unknown column type: %s", body.Type),
		})
		return
	}

	if body.DefaultValue != nil {
		switch body.Type {
		case "varchar", "text":
			switch value := body.DefaultValue.(type) {
			case string:
				queryParts = append(queryParts, fmt.Sprintf(`DEFAULT '%s'`, strings.ReplaceAll(value, "'", "''")))
			default:
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("default value for column %s must be string", body.Name),
				})
				return
			}
		case "int4", "int8", "float4", "float8":
			switch value := body.DefaultValue.(type) {
			case float64:
				queryParts = append(queryParts, fmt.Sprintf("DEFAULT %.0f", value))
			default:
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("default value for column %s must be numeric", body.Name),
				})
				return
			}
		case "bool":
			switch value := body.DefaultValue.(type) {
			case bool:
				if value {
					queryParts = append(queryParts, "DEFAULT true")
				} else {
					queryParts = append(queryParts, "DEFAULT false")
				}
			default:
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("default value for column %s must be boolean", body.Name),
				})
				return
			}
		}
	}

	if _, err := tx.Exec(strings.Join(queryParts, " ")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add column: %v", err),
		})
		return
	}

	if _, err := tx.Exec("INSERT INTO user_defined_column (user_defined_table_id, name, type, length, default_value) VALUES ($1, $2, $3, $4, $5)", tableId, body.Name, body.Type, *body.Length, body.DefaultValue); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to insert into columns: %v", err),
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
		"message": "Column added successfully",
	})
}

// --------------------
// Add Column End
// --------------------

// --------------------
// List Columns
// --------------------
func ListColumns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tableName := chi.URLParam(r, "table-name")

	query := `
		SELECT 
    		udc.name,
    		udc.type,
			udc.length,
    		udc.default_value
		FROM user_defined_column udc
			JOIN user_defined_table udt ON udc.user_defined_table_id = udt.id
		WHERE udt.name = $1;
	`

	rows, err := database.DB.Query(query, tableName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query columns: %v", err),
		})
		return
	}
	defer rows.Close()

	type Column struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Length       *int   `json:"length"`
		DefaultValue any    `json:"default"`
	}

	columns := []Column{}

	for rows.Next() {
		var column Column
		if err := rows.Scan(&column.Name, &column.Type, &column.Length, &column.DefaultValue); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan column: %v", err),
			})
			return
		}
		columns = append(columns, column)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   columns,
	})
}

// --------------------
// List Columns End
// --------------------

// *
func UpdateColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tableName := chi.URLParam(r, "table-name")
	columnName := chi.URLParam(r, "column-name")

	type UpdateColumnBody struct {
		Name         *string `json:"name"`
		DefaultValue any     `json:"default_value"`
	}

	var body UpdateColumnBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	var columnType string

	err := database.DB.
		QueryRow("SELECT type FROM user_defined_columns WHERE name = $1", columnName).
		Scan(
			&columnType,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Column not found",
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query column: %v", err),
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

	queryParts := []string{
		fmt.Sprintf("ALTER TABLE %s", tableName),
	}

	if body.DefaultValue != nil {
		switch columnType {
		case "varchar", "text":
			switch value := body.DefaultValue.(type) {
			case string:
				queryParts = append(queryParts, fmt.Sprintf(`ALTER COLUMN %s SET DEFAULT '%s'`, columnName, strings.ReplaceAll(value, "'", "''")))
			default:
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("default value for column %s must be string", *body.Name),
				})
				return
			}
		case "int4", "int8", "float4", "float8":
			switch value := body.DefaultValue.(type) {
			case float64:
				queryParts = append(queryParts, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT %f", columnName, value))
			default:
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("default value for column %s must be numeric", *body.Name),
				})
				return
			}
		case "bool":
			switch value := body.DefaultValue.(type) {
			case bool:
				if value {
					queryParts = append(queryParts, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT true", columnName))
				} else {
					queryParts = append(queryParts, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT false", columnName))
				}
			default:
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("default value for column %s must be boolean", *body.Name),
				})
				return
			}
		}
	}

	if _, err := tx.Exec(strings.Join(queryParts, " ")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to update column: %v", err),
		})
		return
	}

	if body.Name != nil && *body.Name != columnName {
		if _, err := tx.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, columnName, *body.Name)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to update column: %v", err),
			})
			return
		}
	}

	if _, err := tx.Exec(`
		UPDATE columns SET 
	     	name = $1,
	    	default_value = $2
	    WHERE name = $3`,
		*body.Name, body.DefaultValue, tableName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to update column: %v", err),
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
		"message": "Column updated successfully",
	})
}

// --------------------
// Delete Column
// --------------------
func DeleteColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	columnName := chi.URLParam(r, "column-name")
	tableName := chi.URLParam(r, "table-name")

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

	if _, err := tx.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, columnName)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to drop column: %v", err),
		})
		return
	}

	deleteQuery := `
		DELETE FROM user_defined_column
		WHERE user_defined_table_id = (SELECT id FROM user_defined_table WHERE name = $1)
		AND name = $2
	`
	if _, err := tx.Exec(deleteQuery, tableName, columnName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to delete column: %v", err),
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
		"message": "Column deleted successfully",
	})
}

// --------------------
// Delete Column End
// --------------------
