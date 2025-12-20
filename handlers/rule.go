package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wolfscream/database"
	"wolfscream/models"
)

// --------------------
// List Rules
// --------------------
func ListRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name, description, text FROM rules;"

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
		Text        string    `json:"text"`
	}

	data := []Data{}

	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.Id, &d.Name, &d.Description, &d.Text); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan rule: %v", err),
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
// List Rules End
// --------------------

// --------------------
// Add Rule
// --------------------
func AddRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body models.Rule

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	query := "INSERT INTO rules(name, text, description) VALUES ($1, $2, $3);"
	if _, err := database.DB.Exec(query, body.Name, body.Text, body.Description); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add rule: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Rule added successfully",
	})
}
