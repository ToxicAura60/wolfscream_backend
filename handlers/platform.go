package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wolfscream/database"
)

// --------------------
// List Platforms
// --------------------
func ListPlatforms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name, image_url FROM platforms;"

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
		Id       string `json:"id"`
		Name     string `json:"name"`
		ImageUrl string `json:"image_url"`

	}

	data := []Data{}

	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.Id, &d.Name, &d.ImageUrl); err != nil {
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
// List Platforms End
// --------------------