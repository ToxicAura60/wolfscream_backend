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

	query := "SELECT id, name, image_url FROM communication_platform;"

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

	type CommunicationPlatform struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		ImageUrl string `json:"image_url"`
	}

	communicationPlatforms := []CommunicationPlatform{}

	for rows.Next() {
		var communicationPlatform CommunicationPlatform
		if err := rows.Scan(&communicationPlatform.Id, &communicationPlatform.Name, &communicationPlatform.ImageUrl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan column: %v", err),
			})
			return
		}
		communicationPlatforms = append(communicationPlatforms, communicationPlatform)
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
		"data":   communicationPlatforms,
	})
}

// --------------------
// List Platforms End
// --------------------
