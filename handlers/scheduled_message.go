package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wolfscream/database"
	"wolfscream/discord"
	"wolfscream/models"
	"wolfscream/scheduler"

	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
)

func UpdateScheduledMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

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
		Id *int `json:"id"`
	}

	type Data struct {
		ScheduledMessage
		Table                   Table                   `json:"table"`
		Platform                Platform                `json:"platform"`
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
		"status": "success",
		"data":   data,
	})
}

// --------------------
// List Scheduled Messages End
// --------------------

// --------------------
// GetScheduledMessage
// --------------------
func GetScheduledMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	scheduledMessageName := chi.URLParam(r, "scheduled-message-name")

	type ScheduledMessage struct {
		Id           int    `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		ScheduleType string `json:"schedule_type"`
	}

	type ExecutionStatistic struct {
		SuccessCount int `json:"success_count"`
		FailedCount  int `json:"failed_count"`
	}

	type Table struct {
		Name string `json:"name"`
	}

	type Platform struct {
		Name     string `json:"name"`
		ImageUrl string `json:"image_url"`
	}

	type RunningScheduledMessage struct {
		Id      *int       `json:"id"`
		PrevRun *time.Time `json:"prev_run"`
		NextRun *time.Time `json:"next_run"`
	}

	type Data struct {
		ScheduledMessage
		Table                   Table                   `json:"table"`
		Platform                Platform                `json:"platform"`
		RunningScheduledMessage RunningScheduledMessage `json:"running_scheduled_message"`
		ExecutionHistory        ExecutionStatistic      `json:"execution_statistics"`
	}

	var data Data

	query := `
		SELECT
			sm.id, 
			sm.name,
			sm.description,
			sm.schedule_type,
			t.name AS table_name,
			p.name AS platform_name,
			p.image_url,
			rsm.id AS running_id,
			smeh.success_count,
			smeh.failed_count
		FROM scheduled_messages sm
		JOIN platforms p 
			ON sm.platform_id = p.id
		JOIN tables t 
			ON sm.table_id = t.id
		LEFT JOIN running_scheduled_messages rsm 
			ON sm.id = rsm.scheduled_message_id

		LEFT JOIN LATERAL (
			SELECT
				COUNT(*) FILTER (WHERE status = 'success') AS success_count,
				COUNT(*) FILTER (WHERE status = 'failed')  AS failed_count
			FROM scheduled_message_execution_history
			WHERE scheduled_message_id = sm.id
		) smeh ON true
		WHERE sm.name = $1;
	`

	err := database.DB.
		QueryRow(query, scheduledMessageName).
		Scan(
			&data.Id,
			&data.Name,
			&data.Description,
			&data.ScheduleType,
			&data.Table.Name,
			&data.Platform.Name,
			&data.Platform.ImageUrl,
			&data.RunningScheduledMessage.Id,
			&data.ExecutionHistory.SuccessCount,
			&data.ExecutionHistory.FailedCount,
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
			"message": fmt.Sprintf("Failed to query scheduled message: %v", err),
		})
		return
	}

	if data.RunningScheduledMessage.Id != nil {
		for _, entry := range scheduler.Cron.Entries() {
			if entry.ID == cron.EntryID(*data.RunningScheduledMessage.Id) {

				if !entry.Prev.IsZero() {
					data.RunningScheduledMessage.PrevRun = &entry.Prev
				}

				data.RunningScheduledMessage.NextRun = &entry.Next
				break
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   data,
	})
}

// --------------------
// GetScheduledMessage
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
	Name         string  `json:"name"`
	Message      string  `json:"message"`
	Rule         string  `json:"rule"`
	TableId      int     `json:"table_id"`
	Description  *string `json:"description"`
	ScheduleType string  `json:"schedule_type"`
	PlatformId   int     `json:"platform_id"`

	DiscordConfig DiscordConfig `json:"discord_config"`

	Interval Interval `json:"interval"`

	Cron Cron `json:"cron"`
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

	switch platformName {
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

	switch body.ScheduleType {
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
func EnableScheduledMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	scheduledMessageName := chi.URLParam(r, "scheduled-message-name")

	type ScheduledMessage struct {
		Id           int
		Rule         string
		Message      string
		ScheduleType string
	}

	type Table struct {
		name string
	}

	type Platform struct {
		name string
	}

	var (
		scheduledMessage ScheduledMessage
		table            Table
		platform         Platform
	)

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
		WHERE sm.name = $1
	`, scheduledMessageName).Scan(
		&scheduledMessage.Id,
		&scheduledMessage.ScheduleType,
		&scheduledMessage.Rule,
		&scheduledMessage.Message,
		&table.name,
		&platform.name,
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

	type DiscordConfig struct {
		ChannelId string
	}
	var platformConfig any

	switch platform.name {
	case "discord":
		var discordConfig DiscordConfig
		err = database.DB.QueryRow(`
				SELECT channel_id FROM discord_configs WHERE scheduled_message_id = $1
			`, scheduledMessage.Id).Scan(&discordConfig.ChannelId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to load Discord config: %v", err),
			})
			return
		}
		platformConfig = discordConfig
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Unsupported platform",
		})
		return
	}

	type Interval struct {
		Value int
		Unit  string
	}

	type CronJob struct {
		Minute     *string
		Hour       *string
		DayOfMonth *string
		Month      *string
		DayOfWeek  *string
	}

	var cronSpec string
	switch scheduledMessage.ScheduleType {
	case "interval":
		var interval Interval
		err = database.DB.QueryRow(`
				SELECT value, unit FROM intervals
				WHERE scheduled_message_id = $1
			`, scheduledMessage.Id).Scan(&interval.Value, &interval.Unit)
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
		var cronJob CronJob
		err = database.DB.QueryRow(`
				SELECT minute, hour, day_of_month, month, day_of_week FROM cronjobs WHERE scheduled_message_id = $1
			`, scheduledMessage.Id).Scan(&cronJob.Minute, &cronJob.Hour, &cronJob.DayOfMonth, &cronJob.Month, &cronJob.DayOfWeek)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to load cronjob: %v", err),
			})
			return
		}
		opt := func(p *string) string {
			if p == nil {
				return "*"
			}
			return fmt.Sprintf("%s", *p)
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

	sendMessage := func() {
		rows, err := database.DB.Query(fmt.Sprintf("SELECT * FROM %s", table.name))
		if err != nil {
			logText := fmt.Sprintf("Failed to query table: %v", err)
			database.DB.Exec("INSERT INTO scheduled_message_error_logs (scheduled_message_id, text, level) VALUES ($1, $2, $3);", scheduledMessage.Id, logText, "ERROR")
			return
		}
		defer rows.Close()

		columns, _ := rows.Columns()
		messages := []string{}

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
			for rule := range strings.SplitSeq(scheduledMessage.Rule, ";") {
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

			message := scheduledMessage.Message
			for col, val := range rowMap {
				message = strings.ReplaceAll(message, "{{"+col+"}}", fmt.Sprintf("%v", val))
			}

			messages = append(messages, message)
		}

		if len(messages) == 0 {
			return
		}

		switch platform.name {
		case "discord":
			config := platformConfig.(DiscordConfig)

			channelId := config.ChannelId

			if _, err := discord.DiscordBot.ChannelMessageSend(channelId, strings.Join(messages, "\n\n")); err != nil {
				logText := fmt.Sprintf("Failed to send message to channel %s: %v", channelId, err)
				database.DB.Exec("INSERT INTO scheduled_message_error_logs (scheduled_message_id, text, level) VALUES ($1, $2);", scheduledMessage.Id, logText, "ERROR")
				return
			}
		}
		if _, err := database.DB.Exec("INSERT INTO scheduled_message_execution_history(scheduled_message_id, status) VALUES ($1, $2);", scheduledMessage.Id, "success"); err != nil {
			fmt.Println(err)
		}

	}

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	cronJobId, err := scheduler.Cron.AddFunc(cronSpec, sendMessage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to register schedule: %v", err),
		})
		return
	}

	if _, err := tx.Exec("INSERT INTO scheduled_message_state_history(scheduled_message_id, state) VALUES ($1, $2);", scheduledMessage.Id, "started"); err != nil {
		scheduler.Cron.Remove(cronJobId)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add column: %v", err),
		})
		return
	}

	if _, err := tx.Exec("INSERT INTO running_scheduled_messages(id, scheduled_message_id) VALUES ($1, $2);", int(cronJobId), scheduledMessage.Id); err != nil {
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
			"status":  "error",
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
func DisableScheduledMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	scheduledMessageName := chi.URLParam(r, "scheduled-message-name")

	var runningScheduledMessageId int
	var scheduledMessageId int
	err := database.DB.QueryRow(`
		SELECT 
			sm.id,
			rsm.id
		FROM scheduled_messages sm
		LEFT JOIN running_scheduled_messages rsm ON sm.id = rsm.scheduled_message_id
		WHERE name = $1
	`, scheduledMessageName).Scan(&scheduledMessageId, &runningScheduledMessageId)

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

	tx, err := database.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM running_scheduled_messages WHERE id = $1", runningScheduledMessageId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to remove running scheduled message: %v", err),
		})
		return
	}

	if _, err := tx.Exec("INSERT INTO scheduled_message_state_history(scheduled_message_id, state) VALUES ($1, $2);", scheduledMessageId, "stopped"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to remove running scheduled message: %v", err),
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

	scheduledMessageName := chi.URLParam(r, "scheduled-message-name")

	query := `
    SELECT 
			smel.id,
      smel.text,
			smel.level,
			smel.created_at
    FROM scheduled_message_error_logs smel
		LEFT JOIN scheduled_messages sm ON smel.scheduled_message_id = sm.id
		WHERE sm.name = $1
		ORDER BY smel.created_at ASC;
		`

	rows, err := database.DB.Query(query, scheduledMessageName)
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
		"status": "success",
		"data":   logs,
	})

}

// --------------------
// Fetch Logs End
// --------------------

func FetchStateHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	scheduledMessageName := chi.URLParam(r, "scheduled-message-name")

	query := `
		SELECT
			id,
			state,
			created_at
		FROM (
			SELECT
				smsh.id,
				smsh.state,
				smsh.created_at
			FROM scheduled_message_state_history smsh
			LEFT JOIN scheduled_messages sm
				ON smsh.scheduled_message_id = sm.id
			WHERE sm.name = $1
			ORDER BY smsh.created_at DESC
			LIMIT 100
		) latest
		ORDER BY created_at ASC;
		`

	rows, err := database.DB.Query(query, scheduledMessageName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to query logs: %v", err),
		})
		return
	}
	defer rows.Close()

	type Data struct {
		Id        int       `json:"id"`
		State     string    `json:"state"`
		CreatedAt time.Time `json:"created_at"`
	}

	data := []Data{}

	for rows.Next() {
		var d Data
		if err := rows.Scan(&d.Id, &d.State, &d.CreatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan log: %v", err),
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

	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data":   data,
	})

}
