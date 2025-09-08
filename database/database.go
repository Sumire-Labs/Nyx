package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func New(dbPath string) (*Database, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	db, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	database := &Database{db: db}
	
	if err := database.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	
	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) runMigrations() error {
	migrations := []string{
		createGuildsTable,
		createUsersTable,
		createUserGuildsTable,
		createCommandLogsTable,
		createTicketsTable,
		createTicketPanelsTable,
		createLogSettingsTable,
		createServerLogsTable,
	}
	
	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	
	return nil
}

const createGuildsTable = `
CREATE TABLE IF NOT EXISTS guilds (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    prefix TEXT DEFAULT '!',
    welcome_channel_id TEXT,
    log_channel_id TEXT,
    auto_role_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    discriminator TEXT NOT NULL,
    avatar TEXT,
    bot BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const createUserGuildsTable = `
CREATE TABLE IF NOT EXISTS user_guilds (
    user_id TEXT,
    guild_id TEXT,
    nickname TEXT,
    joined_at DATETIME,
    roles TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, guild_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);`

const createCommandLogsTable = `
CREATE TABLE IF NOT EXISTS command_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    guild_id TEXT,
    channel_id TEXT NOT NULL,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const createTicketsTable = `
CREATE TABLE IF NOT EXISTS tickets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    channel_id TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    status TEXT DEFAULT 'open',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME,
    closed_by TEXT
);`

const createTicketPanelsTable = `
CREATE TABLE IF NOT EXISTS ticket_panels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    message_id TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const createLogSettingsTable = `
CREATE TABLE IF NOT EXISTS log_settings (
    guild_id TEXT PRIMARY KEY,
    log_channel_id TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    log_member_join BOOLEAN DEFAULT TRUE,
    log_member_leave BOOLEAN DEFAULT TRUE,
    log_message_edit BOOLEAN DEFAULT TRUE,
    log_message_delete BOOLEAN DEFAULT TRUE,
    log_role_create BOOLEAN DEFAULT TRUE,
    log_role_update BOOLEAN DEFAULT TRUE,
    log_role_delete BOOLEAN DEFAULT TRUE,
    log_nickname_change BOOLEAN DEFAULT TRUE,
    log_ban BOOLEAN DEFAULT TRUE,
    log_unban BOOLEAN DEFAULT TRUE,
    log_kick BOOLEAN DEFAULT TRUE,
    log_timeout BOOLEAN DEFAULT TRUE,
    log_channel_create BOOLEAN DEFAULT TRUE,
    log_channel_update BOOLEAN DEFAULT TRUE,
    log_channel_delete BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

const createServerLogsTable = `
CREATE TABLE IF NOT EXISTS server_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    user_id TEXT,
    target_id TEXT,
    channel_id TEXT,
    role_id TEXT,
    reason TEXT,
    old_content TEXT,
    new_content TEXT,
    metadata TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX(guild_id),
    INDEX(event_type),
    INDEX(user_id),
    INDEX(timestamp)
);`