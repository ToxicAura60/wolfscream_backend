package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wolfscream/database"
	"wolfscream/models"

	"github.com/go-chi/chi/v5"
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

// --------------------
// List Message Templates
// --------------------
func ListMessageTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT id, name, text, description, created_at FROM message_templates;"

	rows, err := database.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query message templates: %v", err),
		})
		return
	}
	defer rows.Close()

	messageTemplates := []models.MessageTemplate{}

	for rows.Next() {
		var messageTemplate models.MessageTemplate
		if err := rows.Scan(&messageTemplate.Id, &messageTemplate.Name, &messageTemplate.Text, &messageTemplate.Description, &messageTemplate.CreatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan column: %v", err),
			})
			return
		}
		messageTemplates = append(messageTemplates, messageTemplate)
	}

	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to read rows: %v", err),
		})
    return
  }

	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"data": messageTemplates,
	})
}
// --------------------
// List Message Templates End
// --------------------

// --------------------
// Delete Message Template
// --------------------
func DeleteMessageTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	messageTemplateId := chi.URLParam(r, "messageTemplateId")

	if _, err := database.DB.Exec("DELETE FROM message_templates WHERE id = $1", messageTemplateId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to delete metadata: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Message template deleted successfully",
	})
}
// --------------------
// Delete Message Template End End
// --------------------