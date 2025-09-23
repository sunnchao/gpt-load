package db

import (
	"strings"

	"gorm.io/gorm"
)

// AddTokenFields adds token usage fields to request_logs table
func AddTokenFields(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Check if columns already exist by querying information_schema
		var columnExists int64

		// Check if prompt_tokens column exists
		err := tx.Raw(`
			SELECT COUNT(*)
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = 'request_logs'
			AND COLUMN_NAME = 'prompt_tokens'
		`).Scan(&columnExists).Error

		if err != nil {
			return err
		}

		// If columns don't exist, add them
		if columnExists == 0 {
			alterQueries := []string{
				"ALTER TABLE request_logs ADD COLUMN prompt_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN completion_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN total_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN cached_prompt_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN cached_completion_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN reasoning_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN audio_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN image_tokens INT DEFAULT NULL",
				"ALTER TABLE request_logs ADD COLUMN response_body TEXT DEFAULT NULL",
			}

			for _, query := range alterQueries {
				if err := tx.Exec(query).Error; err != nil {
					return err
				}
			}
		}

		// Add indexes if they don't exist (MySQL compatible syntax)
		indexQueries := []struct {
			name  string
			query string
		}{
			{
				name:  "idx_request_logs_total_tokens",
				query: "CREATE INDEX idx_request_logs_total_tokens ON request_logs(total_tokens)",
			},
			{
				name:  "idx_request_logs_group_tokens",
				query: "CREATE INDEX idx_request_logs_group_tokens ON request_logs(group_id, total_tokens)",
			},
		}

		for _, index := range indexQueries {
			// Check if index exists
			var indexExists int64
			err := tx.Raw(`
				SELECT COUNT(*)
				FROM information_schema.STATISTICS
				WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_NAME = 'request_logs'
				AND INDEX_NAME = ?
			`, index.name).Scan(&indexExists).Error

			if err != nil {
				return err
			}

			// Create index if it doesn't exist
			if indexExists == 0 {
				if err := tx.Exec(index.query).Error; err != nil {
					// Ignore index already exists errors
					if !isIndexExistsError(err) {
						return err
					}
				}
			}
		}

		return nil
	})
}

// isIndexExistsError checks if error is due to index already existing
func isIndexExistsError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "duplicate key name") ||
		contains(errStr, "index already exists") ||
		contains(errStr, "Duplicate key name")
}

// contains checks if string contains substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
		 len(str) > len(substr) &&
		 (strings.Contains(strings.ToLower(str), strings.ToLower(substr))))
}