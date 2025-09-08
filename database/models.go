package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Guild struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Prefix           string    `json:"prefix"`
	WelcomeChannelID *string   `json:"welcome_channel_id"`
	LogChannelID     *string   `json:"log_channel_id"`
	AutoRoleID       *string   `json:"auto_role_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type User struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	Avatar        *string   `json:"avatar"`
	Bot           bool      `json:"bot"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserGuild struct {
	UserID    string     `json:"user_id"`
	GuildID   string     `json:"guild_id"`
	Nickname  *string    `json:"nickname"`
	JoinedAt  *time.Time `json:"joined_at"`
	Roles     []string   `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CommandLog struct {
	ID           int64     `json:"id"`
	CommandName  string    `json:"command_name"`
	UserID       string    `json:"user_id"`
	GuildID      *string   `json:"guild_id"`
	ChannelID    string    `json:"channel_id"`
	Success      bool      `json:"success"`
	ErrorMessage *string   `json:"error_message"`
	ExecutedAt   time.Time `json:"executed_at"`
}

func (d *Database) GetGuild(guildID string) (*Guild, error) {
	query := `SELECT id, name, prefix, welcome_channel_id, log_channel_id, auto_role_id, created_at, updated_at 
			  FROM guilds WHERE id = ?`
	
	var guild Guild
	row := d.db.QueryRow(query, guildID)
	err := row.Scan(
		&guild.ID,
		&guild.Name,
		&guild.Prefix,
		&guild.WelcomeChannelID,
		&guild.LogChannelID,
		&guild.AutoRoleID,
		&guild.CreatedAt,
		&guild.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &guild, nil
}

func (d *Database) CreateOrUpdateGuild(guild *Guild) error {
	query := `INSERT OR REPLACE INTO guilds 
			  (id, name, prefix, welcome_channel_id, log_channel_id, auto_role_id, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`
	
	var createdAt interface{}
	if !guild.CreatedAt.IsZero() {
		createdAt = guild.CreatedAt
	}
	
	_, err := d.db.Exec(query,
		guild.ID,
		guild.Name,
		guild.Prefix,
		guild.WelcomeChannelID,
		guild.LogChannelID,
		guild.AutoRoleID,
		createdAt,
	)
	
	return err
}

func (d *Database) GetUser(userID string) (*User, error) {
	query := `SELECT id, username, discriminator, avatar, bot, created_at, updated_at 
			  FROM users WHERE id = ?`
	
	var user User
	row := d.db.QueryRow(query, userID)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Discriminator,
		&user.Avatar,
		&user.Bot,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &user, nil
}

func (d *Database) CreateOrUpdateUser(user *User) error {
	query := `INSERT OR REPLACE INTO users 
			  (id, username, discriminator, avatar, bot, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`
	
	var createdAt interface{}
	if !user.CreatedAt.IsZero() {
		createdAt = user.CreatedAt
	}
	
	_, err := d.db.Exec(query,
		user.ID,
		user.Username,
		user.Discriminator,
		user.Avatar,
		user.Bot,
		createdAt,
	)
	
	return err
}

func (d *Database) GetUserGuild(userID, guildID string) (*UserGuild, error) {
	query := `SELECT user_id, guild_id, nickname, joined_at, roles, created_at, updated_at 
			  FROM user_guilds WHERE user_id = ? AND guild_id = ?`
	
	var userGuild UserGuild
	var rolesJSON string
	row := d.db.QueryRow(query, userID, guildID)
	err := row.Scan(
		&userGuild.UserID,
		&userGuild.GuildID,
		&userGuild.Nickname,
		&userGuild.JoinedAt,
		&rolesJSON,
		&userGuild.CreatedAt,
		&userGuild.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	if rolesJSON != "" {
		if err := json.Unmarshal([]byte(rolesJSON), &userGuild.Roles); err != nil {
			userGuild.Roles = []string{}
		}
	}
	
	return &userGuild, nil
}

func (d *Database) CreateOrUpdateUserGuild(userGuild *UserGuild) error {
	rolesJSON, _ := json.Marshal(userGuild.Roles)
	
	query := `INSERT OR REPLACE INTO user_guilds 
			  (user_id, guild_id, nickname, joined_at, roles, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`
	
	var createdAt interface{}
	if !userGuild.CreatedAt.IsZero() {
		createdAt = userGuild.CreatedAt
	}
	
	_, err := d.db.Exec(query,
		userGuild.UserID,
		userGuild.GuildID,
		userGuild.Nickname,
		userGuild.JoinedAt,
		string(rolesJSON),
		createdAt,
	)
	
	return err
}

func (d *Database) LogCommand(log *CommandLog) error {
	query := `INSERT INTO command_logs 
			  (command_name, user_id, guild_id, channel_id, success, error_message)
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := d.db.Exec(query,
		log.CommandName,
		log.UserID,
		log.GuildID,
		log.ChannelID,
		log.Success,
		log.ErrorMessage,
	)
	
	return err
}

func (d *Database) GetCommandStats(guildID string, limit int) ([]CommandLog, error) {
	var query string
	var args []interface{}
	
	if guildID != "" {
		query = `SELECT id, command_name, user_id, guild_id, channel_id, success, error_message, executed_at 
				 FROM command_logs WHERE guild_id = ? 
				 ORDER BY executed_at DESC LIMIT ?`
		args = []interface{}{guildID, limit}
	} else {
		query = `SELECT id, command_name, user_id, guild_id, channel_id, success, error_message, executed_at 
				 FROM command_logs 
				 ORDER BY executed_at DESC LIMIT ?`
		args = []interface{}{limit}
	}
	
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []CommandLog
	for rows.Next() {
		var log CommandLog
		err := rows.Scan(
			&log.ID,
			&log.CommandName,
			&log.UserID,
			&log.GuildID,
			&log.ChannelID,
			&log.Success,
			&log.ErrorMessage,
			&log.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}