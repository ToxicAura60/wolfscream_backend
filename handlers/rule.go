package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wolfscream/database"
	"wolfscream/validator"
)

// --------------------
// List Rules
// --------------------
func ListRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query("SELECT id, name, description, rule FROM scheduled_message_rule ORDER BY created_at ASC;")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query rules: %v", err),
		})
		return
	}
	defer rows.Close()

	type ScheduledMessageRule struct {
		Id          int     `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Rule        string  `json:"rule"`
	}

	scheduledMessageRules := []ScheduledMessageRule{}

	for rows.Next() {
		var scheduledMessageRule ScheduledMessageRule
		if err := rows.Scan(&scheduledMessageRule.Id, &scheduledMessageRule.Name, &scheduledMessageRule.Description, &scheduledMessageRule.Rule); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan rule: %v", err),
			})
			return
		}
		scheduledMessageRules = append(scheduledMessageRules, scheduledMessageRule)
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
		"data":   scheduledMessageRules,
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

	type AddRuleBody struct {
		Name        string  `json:"name" validate:"required,snakecase,min=1"`
		Description *string `json:"description"`
		Rule        string  `json:"rule" validate:"required"`
	}

	var body AddRuleBody

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

	query := "INSERT INTO scheduled_message_rule(name, rule, description) VALUES ($1, $2, $3);"
	if _, err := database.DB.Exec(query, body.Name, body.Rule, body.Description); err != nil {
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
