package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"wolfscream/database"
	"wolfscream/jobs"
	"wolfscream/models"
	"wolfscream/scheduler"

	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
)

// --------------------
// List Scheduled Messages
// --------------------
func ListScheduledMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	 query := `
    SELECT 
			sm.id,
      sm.name,
			sm.description,
			sm.schedule_type,
			t.name,
			p.name,
			p.image_url,
      rsm.id
    FROM scheduled_messages sm
		JOIN platforms p ON sm.platform_id = p.id
		JOIN tables t on sm.table_id = t.id
    LEFT JOIN running_scheduled_messages rsm ON sm.id = rsm.scheduled_message_id
    ORDER BY sm.created_at DESC;`

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

	type ScheduledMessage struct {
  	Id           int    `json:"id"`
    Name         string `json:"name"`
		Description  string `json:"description"`
		ScheduleType string `json:"schedule_type"`
  }

	type Table struct {
		Name string
	}

	type Platform struct {
		Name     string `json:"name"`
		ImageUrl string `json:"image_url"`
	}

	type RunningScheduledMessage struct {
		Id *string `json:"id"`
	}

	type Data struct{
		ScheduledMessage
		Table Table `json:"table"`
		Platform Platform       `json:"platform"`
		RunningScheduledMessage RunningScheduledMessage `json:"running_scheduled_message"`
	}

	var data []Data

	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.Id, &d.Name, &d.Description, &d.ScheduleType, &d.Table.Name, &d.Platform.Name, &d.Platform.ImageUrl, &d.RunningScheduledMessage.Id); err != nil {
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
// List Scheduled Messages End
// --------------------




// --------------------
// Add Scheduled Message
// --------------------
type DiscordConfig struct {
	ChannelId string `json:"channel_id"`
}

type Interval struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type Cron struct {
	Minute     int `json:"minute"`
	Hour       int `json:"hour"`
	DayOfMonth int `json:"day_of_month"`
	Month      int `json:"month"`
	DayOfWeek  int `json:"day_of_week"`
}

type AddScheduleMessageBody struct {
  Name          string        `json:"name"`
  Message	      string        `json:"message"`
	Rule			    string        `json:"rule"`
	TableId       int           `json:"table_id"`
	Description   *string       `json:"description"`
	ScheduleType  string        `json:"schedule_type"`
	PlatformId    int           `json:"platform_id"`

	DiscordConfig DiscordConfig `json:"discord_config"`

	Interval      Interval			`json:"interval"`

	Cron 					Cron					`json:"cron"`
}


func AddScheduledMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body AddScheduleMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid JSON request body",
		})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to start database transaction",
		})
		return
	}
	defer tx.Rollback() 

	var scheduledMessageID int

	query := `
		INSERT INTO scheduled_messages (
			name, description, table_id, message, rule, platform_id, schedule_type
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING
			id;
	`
	err = tx.QueryRow(
		query,
		body.Name,
		body.Description,
		body.TableId,
		body.Message,
		body.Rule,
		body.PlatformId,
		body.ScheduleType,
	).Scan(&scheduledMessageID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create scheduled message: %v", err),
		})
		return
	}

	var platformName string
	err = tx.QueryRow("SELECT name FROM platforms WHERE id = $1", body.PlatformId).Scan(&platformName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to get platform name: %v", err),
		})
		return
	}

	switch(platformName) {
		case "discord":
			query := "INSERT INTO discord_configs(channel_id, scheduled_message_id) VALUES ($1, $2)"
			if _, err := tx.Exec(query, body.DiscordConfig.ChannelId, scheduledMessageID); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to insert Discord configuration: %v", err),
				})
				return
			}
		default: 
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Unsupported platform: %s", platformName),
			})
			return
	}

	switch(body.ScheduleType) {
		case "interval":
			query = "INSERT INTO intervals(value, unit, scheduled_message_id) VALUES ($1, $2, $3);"
			if _, err := tx.Exec(query, body.Interval.Value, body.Interval.Unit, scheduledMessageID); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to insert interval schedule: %v", err),
				})
				return
			}
		case "cron":
			query = "INSERT INTO cron_jobs(minute, hour, day_of_month, month, day_of_week, scheduled_message_id) VALUES ($1, $2, $3, $4, $5, $6);"
			if _, err := tx.Exec(query, body.Cron.Minute, body.Cron.Hour, body.Cron.DayOfMonth, body.Cron.Month, body.Cron.DayOfWeek, scheduledMessageID); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to insert cron schedule: %v", err),
				})
				return
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Unsupported schedule type: %s", body.ScheduleType),
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
	
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Scheduled message added successfully",
	})
}
// --------------------
// Add Scheduled Message End
// --------------------


type EnableScheduledMessageBody struct {
	ScheduledMessageId int `json:"scheduled_message_id"`
}

func EnableScheduledMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	

	var body EnableScheduledMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}
	
	var scheduledMessageId int
	var tableName, message, platformName, ruleText, scheduleType string

	err := database.DB.QueryRow(`
	SELECT 
		sm.id,
		sm.schedule_type,
		sm.rule,
		sm.message,
		t.name,
		p.name
		FROM scheduled_messages sm
		JOIN tables t ON sm.table_id = t.id
		JOIN platforms p ON sm.platform_id = p.id
		WHERE sm.id = $1
	`, body.ScheduledMessageId).Scan(
		&scheduledMessageId,
		&scheduleType,
		&ruleText,
		&message,
		&tableName,
		&platformName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Scheduled message not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("DB error: %v", err),
		})
		return
	}

	var config any
	switch platformName {
		case "discord":
			var discordConfig models.DiscordConfig
			err = database.DB.QueryRow(`
				SELECT id, scheduled_message_id, channel_id, created_at FROM discord_configs WHERE scheduled_message_id = $1
			`, scheduledMessageId).Scan(&discordConfig.Id, &discordConfig.ScheduledMessageId, &discordConfig.ChannelId, &discordConfig.CreatedAt)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to load Discord config: %v", err),
				})
				return
			}
			config = discordConfig
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Unsupported platform: " + platformName,
			})
			return
		}

	
	var cronSpec string

	switch scheduleType {
		case "interval":
			var interval models.Interval
			err = database.DB.QueryRow(`
				SELECT id, scheduled_message_id, value, unit, created_at
				FROM intervals
				WHERE scheduled_message_id = $1
			`, scheduledMessageId).Scan(&interval.Id, &interval.ScheduledMessageId, &interval.Value, &interval.Unit, &interval.CreatedAt)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to load interval: %v", err),
				})
				return
			}
			cronSpec = fmt.Sprintf("@every %d%s", interval.Value, interval.Unit)
		case "cronjob":
			var cronJob models.CronJob
			err = database.DB.QueryRow(`
				SELECT id, scheduled_message_id, minute, hour, day_of_month, month, day_of_week, created_at
				FROM cronjobs
				WHERE scheduled_message_id = $1
			`, scheduledMessageId).Scan(&cronJob.Id, &cronJob.ScheduledMessageId, &cronJob.Minute, &cronJob.Hour, &cronJob.DayOfMonth, &cronJob.Month, &cronJob.DayOfWeek, &cronJob.CreatedAt)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": fmt.Sprintf("Failed to load cronjob: %v", err),
				})
				return
			}
			opt := func(p *int) string {
				if p == nil {
					return "*"
				}
				return fmt.Sprintf("%d", *p)
			}
			cronSpec = fmt.Sprintf("%s %s %s %s %s",
				opt(cronJob.Minute),
				opt(cronJob.Hour),
				opt(cronJob.DayOfMonth),
				opt(cronJob.Month),
				opt(cronJob.DayOfWeek),
		)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "No interval or cronjob found for schedule",
			})
			return
	}

	jobFunc, ok := jobs.JobHandlers[platformName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Handler not found for platform: " + platformName,
		})
		return
	}

	wrappedJob := func() {
		rows, err := database.DB.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
		if err != nil {
			logText := fmt.Sprintf("Failed to query table: %v", err)
			database.DB.Exec("INSERT INTO logs (scheduled_message_id, text) VALUES ($1, $2);", scheduledMessageId, logText)
			return
		}
		defer rows.Close()

		columns, _ := rows.Columns()
		allMessages := []string{}

		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}

			if err := rows.Scan(pointers...); err != nil {
				continue
			}

			rowMap := make(map[string]any)
			for i, column := range columns {
				rowMap[column] = values[i]
			}

			match := true
			for _, rule := range strings.Split(ruleText, ";") {
				rule = strings.TrimSpace(rule)
				if rule == "" {
					continue
				}

				parts := strings.SplitN(rule, " ", 3)
				if len(parts) != 3 {
					match = false
					break
				}

				key, operator, val := parts[0], parts[1], parts[2]

				v, ok := rowMap[key]
				if !ok {
					match = false
					break
				}
		
				switch operator {
					case "==":
						if fmt.Sprintf("%v", v) != val {
							match = false
						}
					case "!=":
						if fmt.Sprintf("%v", v) == val {
							match = false
						}
					default:
						match = false
				}

				if !match {
					break
				}
			}

			if !match {
				continue
			}

			msgText := message
			for col, val := range rowMap {
				msgText = strings.ReplaceAll(msgText, "{{"+col+"}}", fmt.Sprintf("%v", val))
			}

			allMessages = append(allMessages, msgText)
		}

		if len(allMessages) == 0 {
			return
		}

		payload := map[string]any{
			"messages": allMessages,
			"config": config,
		}
		
		jobFunc(payload)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	cronJobId, err := scheduler.Cron.AddFunc(cronSpec, wrappedJob)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to register schedule: %v", err),
		})
		return
	}

	if _, err := tx.Exec("INSERT INTO running_scheduled_messages(id, scheduled_message_id) VALUES ($1, $2);", int(cronJobId), scheduledMessageId); err != nil {
		scheduler.Cron.Remove(cronJobId)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add column: %v", err),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		scheduler.Cron.Remove(cronJobId)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error", 
			"message": "Failed to commit transaction",
		})
		return
	}


	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Scheduled message activated",
	})
}




// --------------------
// Disable Scheduled Message
// --------------------
type DisableScheduledMessageBody struct {
	ScheduledMessageId int `json:"scheduled_message_id"`
}

func DisableScheduledMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body DisableScheduledMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	var runningScheduledMessageId int
	err := database.DB.QueryRow(`
		SELECT id
		FROM running_scheduled_messages
		WHERE scheduled_message_id = $1
	`, body.ScheduledMessageId).Scan(&runningScheduledMessageId)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Scheduled message is not running",
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("DB error: %v", err),
		})
		return
	}

	scheduler.Cron.Remove(cron.EntryID(runningScheduledMessageId))

	if _, err := database.DB.Exec("DELETE FROM running_scheduled_messages WHERE id = $1", runningScheduledMessageId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to remove running scheduled message: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Scheduled message disabled",
	})
}

// --------------------
// Disable Scheduled Message End
// --------------------

// --------------------
// Fetch Logs
// --------------------
func FetchLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	scheduledMessageId := chi.URLParam(r, "id")

	query := `
    SELECT 
			id,
      text,
			level,
			created_at
    FROM logs WHERE scheduled_message_id = $1
		ORDER BY created_at ASC;
		`

	rows, err := database.DB.Query(query, scheduledMessageId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query logs: %v", err),
		})
		return
	}
	defer rows.Close()

	logs := []models.Log{}

	for rows.Next() {
		var log models.Log
		if err := rows.Scan(&log.Id, &log.Text, &log.Level, &log.CreatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan log: %v", err),
			})
			return
		}
		logs = append(logs, log)
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
		"data": logs,
	})

}

// --------------------
// Fetch Logs End
// --------------------