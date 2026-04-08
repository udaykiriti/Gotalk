package storage

import (
	"database/sql"
	"fmt"
	"log"

	"gotalk/internal/models"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.initialize(); err != nil {
		return nil, fmt.Errorf("could not initialize database: %w", err)
	}

	return s, nil
}

func (s *Store) initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room TEXT NOT NULL,
		user_name TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *Store) SaveMessage(msg models.Message) error {
	query := `INSERT INTO messages (room, user_name, content) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, msg.Room, msg.User, msg.Content)
	if err != nil {
		log.Printf("Error saving message: %v", err)
	}
	return err
}

func (s *Store) GetRoomHistory(room string, limit int) ([]models.Message, error) {
	// We select the last N messages and then sort them ASC (chronologically)
	subQuery := `
		SELECT user_name, content, created_at
		FROM messages
		WHERE room = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	fullQuery := fmt.Sprintf("SELECT user_name, content FROM (%s) AS history ORDER BY created_at ASC", subQuery)

	rows, err := s.db.Query(fullQuery, room, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.Message
	for rows.Next() {
		var msg models.Message
		msg.Type = models.TypeMessage
		msg.Room = room
		if err := rows.Scan(&msg.User, &msg.Content); err != nil {
			return nil, err
		}
		history = append(history, msg)
	}

	return history, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
