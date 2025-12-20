package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"wolfscream/database"
	"wolfscream/models"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)


type AddColumnBody struct {
	models.Column
}

// *
func AddColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	table := chi.URLParam(r, "table")

	var body AddColumnBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	validName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	if !validName.MatchString(table) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "invalid table name",
		})
		return
	}

	if !validName.MatchString(body.Name) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "invalid column name",
		})
		return
	}

	var tableId int
	err := database.DB.QueryRow("SELECT id FROM tables WHERE name = $1", table).Scan(&tableId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Table %s does not exist", table),
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

	parts := []string{
		fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s`, table, body.Name),
	}

	var length int
	switch body.Type {
		case "int4":
			parts = append(parts, "INTEGER")
		case "int8":
			parts = append(parts, "BIGINT")
		case "varchar":
			if body.Length != nil && *body.Length != 0 {
				parts = append(parts, fmt.Sprintf("VARCHAR(%d)", *body.Length))
				length = *body.Length
			} else {
				parts = append(parts, "VARCHAR(255)")
				length = 255
			}
		case "text":
			parts = append(parts, "TEXT")
		case "bool":
			parts = append(parts, "BOOLEAN")
		case "float4":
			parts = append(parts, "REAL")
		case "float8":
			parts = append(parts, "DOUBLE PRECISION")
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Unknown column type: %s", body.Type),
			})
			return
	}

	if length != 0 {
		body.Length = &length
	} else {
		body.Length = nil
	}

	if body.DefaultValue != nil {
		switch body.Type {
			case "varchar", "text":
				switch value := body.DefaultValue.(type) {
					case string:
						parts = append(parts, fmt.Sprintf(`DEFAULT '%s'`, strings.ReplaceAll(value, "'", "''")))
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
						parts = append(parts, fmt.Sprintf("DEFAULT %.0f", value))
					default:
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(map[string]string{
							"status":  "error",
							"message": fmt.Sprintf("default value for column %s must be numeric",	body.Name),
						})
						return
				}
			case "bool":
				switch value := body.DefaultValue.(type) {
					case bool:
						if value {
							parts = append(parts, "DEFAULT true")
						} else {
							parts = append(parts, "DEFAULT false")
						}
					default:
						json.NewEncoder(w).Encode(map[string]string{
							"status":  "error",
							"message": fmt.Sprintf("default value for column %s must be boolean",	body.Name),
						})
						return
				}
		}
	}

	alterQuery := strings.Join(parts, " ")
	
	if _, err := tx.Exec(alterQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add column: %v", err),
		})
		return
	}

	insertQuery := `
		INSERT INTO columns (table_id, name, type, length, default_value)
		VALUES ($1, $2, $3, $4, $5)`

	if _, err := tx.Exec(insertQuery, tableId, body.Name, body.Type, body.Length, body.DefaultValue); err != nil {
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
			"status": "error", 
			"message": "Failed to commit transaction",
		})
		return
	}


	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Column added successfully",
	})
}

type UpdateColumnBody struct {
	models.Column
	FallbackValue any `json:"fallback_value"`
}

// *
func UpdateColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	table := chi.URLParam(r, "table")
	column := chi.URLParam(r, "column")

	var body UpdateColumnBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	validName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	if !validName.MatchString(table) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "invalid table name",
		})
		return
	}

	if !validName.MatchString(column) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "invalid column name",
		})
		return
	}

	oldColumnName := column
	newColumnName := column
	if body.Name != "" {
		newColumnName = body.Name
	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error", 
			"message": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	if newColumnName != oldColumnName {
		renameQuery := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, oldColumnName, newColumnName)

		if _, err := tx.Exec(renameQuery); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to update column: %v", err),
			})
			return
		}
	}

	parts := []string{
		fmt.Sprintf("ALTER COLUMN %s DROP DEFAULT", newColumnName),
	}


	if body.DefaultValue != nil {
		switch body.Type {
			case "varchar", "text":
				switch value := body.DefaultValue.(type) {
					case string:
						parts = append(parts, fmt.Sprintf(`ALTER COLUMN %s SET DEFAULT '%s'`, newColumnName, strings.ReplaceAll(value, "'", "''")))
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
						parts = append(parts, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT %.0f", newColumnName, value))
					default:
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(map[string]string{
							"status":  "error",
							"message": fmt.Sprintf("default value for column %s must be numeric",	body.Name),
						})
						return
				}
			case "bool":
				switch value := body.DefaultValue.(type) {
					case bool:
						if value {
							parts = append(parts, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT true", newColumnName))
						} else {
							parts = append(parts, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT false", newColumnName))
						}
					default:
						json.NewEncoder(w).Encode(map[string]string{
							"status":  "error",
							"message": fmt.Sprintf("default value for column %s must be boolean",	body.Name),
						})
						return
				}
		}
	}

	var columnType string
	var regex string
	fallbackValue := "NULL"

	var length int
	switch body.Type {
		case "int4":
			if body.FallbackValue == nil {
				fallbackValue = "0"
			}
			columnType = "INTEGER"
			regex = `^[0-9]+$`
		case "int8":
			if body.FallbackValue == nil  {
				fallbackValue = "0"
			}
			columnType = "BIGINT"
			regex = `^[0-9]+$`
		case "varchar":
			if body.FallbackValue == nil {
				fallbackValue = "''"
			}
			if body.Length != nil && *body.Length != 0 {
				columnType = fmt.Sprintf("VARCHAR(%d)", *body.Length)
				length = *body.Length
			} else {
				columnType = "VARCHAR(255)"
				length = 255
			}
		case "text":
			if body.FallbackValue == nil {
				fallbackValue = "''"
			}
			columnType = "TEXT"
			regex = ".*"
		case "bool":
			if body.FallbackValue == nil {
				fallbackValue = "FALSE"
			}
			columnType = "BOOLEAN"
			regex = "^(true|false|t|f|yes|no|on|off)$"
		case "float4":
			if body.FallbackValue == nil {
				fallbackValue = "0"
			}
			columnType = "REAL"
			regex = `^[0-9]+(\.[0-9]+)?$`
		case "float8":
			if body.FallbackValue == nil  {
				fallbackValue = "0"
			}
			columnType = "DOUBLE PRECISION"
			regex = `^[0-9]+(\.[0-9]+)?$`
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Unknown column type: %s", body.Type),
			})
			return
	}

	if length != 0 {
		body.Length = &length
	}
	body.Length = nil


	parts = append(parts, fmt.Sprintf("ALTER COLUMN %s TYPE %s USING (CASE WHEN %s::TEXT ~ '%s' THEN (%s::TEXT)::%s ELSE %s END)", newColumnName, columnType, newColumnName, regex, newColumnName, columnType, fallbackValue))

	alterQuery := fmt.Sprintf("ALTER TABLE %s %s;", table, strings.Join(parts, ", "))

	if _, err := tx.Exec(alterQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to update column: %v", err),
		})
		return
	}

	updateQuery := `
	  UPDATE columns
	    SET 
	      name = $1,
	      type = $2,
	      length = $3,
	      default_value = $4
	    WHERE name = $5`
		
	if _, err := tx.Exec(updateQuery, newColumnName, body.Type, body.Length, body.DefaultValue, oldColumnName); err != nil {
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
			"status": "error", 
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
// List Columns
// --------------------
func ListColumns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	table := chi.URLParam(r, "table")

	query := `
		SELECT 
    	columns.name AS column_name,
    	columns.type,
			columns.length,
    	columns.default_value
		FROM columns
			JOIN tables ON columns.table_id = tables.id
		WHERE tables.name = $1;
	`
	
	rows, err := database.DB.Query(query, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query columns: %v", err),
		})
		return
	}
	defer rows.Close()

 	columns := []models.Column{}

	for rows.Next() {
		var column models.Column
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
		"status":  "success",
		"data": columns,
	})
}
// --------------------
// List Columns End
// --------------------

// --------------------
// Delete Column
// --------------------
func DeleteColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	column := chi.URLParam(r, "column")
	table := chi.URLParam(r, "table")

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

	dropQuery := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", pq.QuoteIdentifier(table), pq.QuoteIdentifier(column))
	if _, err := tx.Exec(dropQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to drop column: %v", err),
		})
		return
	}

	deleteQuery := `
		DELETE FROM columns
		WHERE table_id = (SELECT id FROM tables WHERE name = $1)
		AND name = $2
	`
	if _, err := tx.Exec(deleteQuery, table, column); err != nil {
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
			"status": "error", 
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