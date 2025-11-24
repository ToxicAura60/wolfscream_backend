package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wolfscream/database"
)

type AddTemplateBody struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func AddTemplate(w http.ResponseWriter, r *http.Request) {
	var body AddTemplateBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	query := "INSERT INTO message_templates(name, text) VALUES ($1, $2);"
	if _, err := database.DB.Exec(query, body.Name, body.Text); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add template: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Table %s added successfully", body.Name),
	})
}