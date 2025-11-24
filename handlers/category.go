package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wolfscream/database"
	"wolfscream/models"

	"github.com/go-chi/chi/v5"
)

// --------------------
// Add Category
// --------------------
type AddCategoryBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func AddCategory(w http.ResponseWriter, r *http.Request) {
	var body AddCategoryBody

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

	createQuery := fmt.Sprintf(`CREATE TABLE "%s" (id SERIAL PRIMARY KEY, name VARCHAR(64) NOT NULL UNIQUE, description TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`, body.Name)
	if _, err := tx.Exec(createQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create table: %v", err),
		})
		return
	}

	insertQuery := `INSERT INTO categories(name, description) VALUES ($1, $2)`
	if _, err := database.DB.Exec(insertQuery, body.Name, body.Description); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add category: %v", err),
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
		"message": fmt.Sprintf("Categories %s added successfully", body.Name),
	})
}
// --------------------
// Add Categories End
// --------------------

// --------------------
// List Categories
// --------------------
func ListCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT name, description, created_at FROM categories;"

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

	categories := []models.Category{}

	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.Name, &category.Description, &category.CreatedAt); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to scan column: %v", err),
			})
			return
		}
		categories = append(categories, category)
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
		"data": categories,
	})
}
// --------------------
// List Tables End
// --------------------


// --------------------
// Delete Category
// --------------------
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	categoryName := chi.URLParam(r, "categoryName")

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

	dropQuery := fmt.Sprintf("DROP TABLE %s", categoryName)
	if _, err := tx.Exec(dropQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to drop category: %v", err),
		})
		return
	}

	deleteQuery := `DELETE FROM categories WHERE name = $1`
	if _, err := tx.Exec(deleteQuery, categoryName); err != nil {
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
		"message": fmt.Sprintf("Category %s deleted successfully", categoryName),
	})
}


// --------------------
// Delete Category End
// --------------------



// --------------------
// Add Category Item
// --------------------
type AddCategoryItemBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func AddCategoryItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	categoryName := chi.URLParam(r, "categoryName")

	var body AddCategoryItemBody
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

	createQuery := fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
    	name VARCHAR(64) NOT NULL UNIQUE,
    	description TEXT, 
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, body.Name)
	if _, err := tx.Exec(createQuery); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create table: %v", err),
		})
		return
	}

	query := fmt.Sprintf("INSERT INTO %s(name, description) VALUES ($1, $2);", categoryName)
	if _, err := tx.Exec(query, body.Name, body.Description); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to add category: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Severity %s added successfully", body.Name),
	})
}
// --------------------
// Add Category End
// --------------------




// --------------------
// Update Category
// --------------------
type UpdateCategoryBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	categoryName := chi.URLParam(r, "categoryName")

	var body AddCategoryItemBody
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
			"status": "error", 
			"message": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	if body.Name != "" && categoryName != body.Name  {
		alterQuery := fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", categoryName, body.Name)
		if _, err := tx.Exec(alterQuery); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": fmt.Sprintf("Failed to add category: %v", err),
			})
			return
		}
	}
	
	updateQuery := "UPDATE categories SET description = $1"
	args := []any{body.Description}

	if body.Name != "" {
		updateQuery += ", name = $2 WHERE name = $3"
		args = append(args, body.Name, categoryName)
	} else {
		updateQuery += " WHERE name = $2"
		args = append(args, categoryName)
	}

	if _, err := tx.Exec(updateQuery, args...); err != nil {
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
			"status":  "error",
			"message": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Category %s added successfully", body.Name),
	})
}

// --------------------
// Update Category End
// --------------------