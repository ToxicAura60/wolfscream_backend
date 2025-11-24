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

// --------------------
// Create Table
// --------------------
type CreateTableBody struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Categories []string `json:"categories"`
}

func CreateTable(w http.ResponseWriter, r *http.Request) {
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

	parts := []string{
		`"id" SERIAL PRIMARY KEY`,
	}

	for _, category := range body.Categories {
		parts = append(parts, fmt.Sprintf(`%s INTEGER REFERENCES %s("id")`, pq.QuoteIdentifier(category + "_id"), pq.QuoteIdentifier(category)))
	}

	query := fmt.Sprintf(`CREATE TABLE %s (%s);`, pq.QuoteIdentifier(body.Name), strings.Join(parts, ", "))

	if _, err := tx.Exec(query); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create table: %v", err),
		})
		return
	}

	var tableId int
	err = tx.QueryRow(`INSERT INTO "tables" ("name", "description") VALUES ($1, $2) RETURNING id`, body.Name, body.Description).Scan(&tableId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to insert table: %v", err),
		})
		return
	}

	if len(body.Categories) > 0 {
		rows, err := tx.Query("SELECT id FROM categories WHERE name = ANY($1)", pq.Array(body.Categories))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to fetch category IDs: %v", err),
			})
			return
		}
    defer rows.Close()

		var categoryIds []int
    for rows.Next() {
      var categoryId int
      if err := rows.Scan(&categoryId); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to read category ID from database: %v", err),
			})
			return
      }
			categoryIds = append(categoryIds, categoryId)
    }

		query = `INSERT INTO tables_categories (table_id, category_id)
            SELECT $1, UNNEST($2::int[]);`

    if _, err := tx.Exec(query, tableId, pq.Array(categoryIds)); err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      json.NewEncoder(w).Encode(map[string]string{
        "status":  "error",
        "message": fmt.Sprintf("Failed to insert tables_categories: %v", err),
      })
      return
    }
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
		"message": fmt.Sprintf("Table %s added successfully", body.Name),
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

	query := "SELECT name, description, created_at FROM tables;"

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

	tables := []models.Table{}

	for rows.Next() {
		var table models.Table
		if err := rows.Scan(&table.Name, &table.Description, &table.CreatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan column: %v", err),
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
		"status":  "success",
		"data": tables,
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

	validName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	
	if !validName.MatchString(table) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "invalid table name",
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

	dropQuery := fmt.Sprintf("DROP TABLE %s", table)
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
		"message": fmt.Sprintf("Table %s dropped successfully", table),
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

	query := `
    SELECT c.name
    FROM tables_categories tc
    JOIN categories c ON tc.category_id = c.id
    JOIN tables t ON t.id = tc.table_id
    WHERE t.name = $1;
  `

	rows, err := database.DB.Query(query, table)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to fetch categories: %v", err),
		})
    return
  }
  defer rows.Close()

	categories := []string{}
  for rows.Next() {
    var category string
    if err := rows.Scan(&category); err != nil {
       w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "failed to read category",
			})
      return
    }
		categories = append(categories, category)
  }

	selectParts := []string{fmt.Sprintf("%s.*", table)}
	for _, category := range categories {
			selectParts = append(selectParts, fmt.Sprintf("%s.name AS %s", category, category))
	}

	var joinParts []string
	for _, category := range categories {
		joinParts = append(joinParts, fmt.Sprintf("LEFT JOIN %s ON %s.%s_id = %s.id", category, table, category, category))
	}

	query = fmt.Sprintf(
		"SELECT %s FROM %s %s;", 
		strings.Join(selectParts, ", "),
		table,
		strings.Join(joinParts, " "),
	)
  rows, err = database.DB.Query(query)
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

		entry := map[string]any{}
		for i, col := range columns {
			entry[col] = values[i]
		}

		for _, category := range categories {
			delete(entry, category+"_id")
		}

    results = append(results, entry)
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

	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validName.MatchString(table) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid table name",
		})
		return
	}

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

	rows, err := tx.Query(`
		SELECT c.id, c.name
		FROM tables_categories tc
		JOIN categories c ON tc.category_id = c.id
		JOIN tables t ON tc.table_id = t.id
		WHERE t.name = $1
	`, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	defer rows.Close()

	categories := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		categories[name] = id
	}

	columns := []string{}
	placeholders := []string{}
	values := []any{}
	i := 1

	for key, val := range body {
		if _, ok := categories[key]; ok {
			columnName := fmt.Sprintf("%s_id", key)
			var valueID int

			err := tx.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE name=$1", key), val).Scan(&valueID)
			if err == sql.ErrNoRows {
				err = tx.QueryRow(fmt.Sprintf("INSERT INTO %s(name) VALUES($1) RETURNING id", key), val).Scan(&valueID)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"status":  "error",
						"message": fmt.Sprintf("Failed to insert value '%v' into category '%s'", val, key),
					})
					return
				}
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}

			columns = append(columns, columnName)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, valueID)
		} else {
			columns = append(columns, key)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, val)
		}
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














