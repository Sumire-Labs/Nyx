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

type Ticket struct {
	ID        int64      `json:"id"`
	GuildID   string     `json:"guild_id"`
	UserID    string     `json:"user_id"`
	ChannelID string     `json:"channel_id"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	ClosedBy  *string    `json:"closed_by"`
}

type TicketPanel struct {
	ID          int64     `json:"id"`
	GuildID     string    `json:"guild_id"`
	ChannelID   string    `json:"channel_id"`
	MessageID   string    `json:"message_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
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

func (d *Database) CreateTicket(ticket *Ticket) error {
	query := `INSERT INTO tickets (guild_id, user_id, channel_id, title, status)
			  VALUES (?, ?, ?, ?, ?)`
	
	result, err := d.db.Exec(query,
		ticket.GuildID,
		ticket.UserID,
		ticket.ChannelID,
		ticket.Title,
		ticket.Status,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	ticket.ID = id
	return nil
}

func (d *Database) GetTicket(channelID string) (*Ticket, error) {
	query := `SELECT id, guild_id, user_id, channel_id, title, status, created_at, closed_at, closed_by
			  FROM tickets WHERE channel_id = ?`
	
	var ticket Ticket
	row := d.db.QueryRow(query, channelID)
	err := row.Scan(
		&ticket.ID,
		&ticket.GuildID,
		&ticket.UserID,
		&ticket.ChannelID,
		&ticket.Title,
		&ticket.Status,
		&ticket.CreatedAt,
		&ticket.ClosedAt,
		&ticket.ClosedBy,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &ticket, nil
}

func (d *Database) GetUserOpenTickets(guildID, userID string) ([]*Ticket, error) {
	query := `SELECT id, guild_id, user_id, channel_id, title, status, created_at, closed_at, closed_by
			  FROM tickets WHERE guild_id = ? AND user_id = ? AND status = 'open'`
	
	rows, err := d.db.Query(query, guildID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tickets []*Ticket
	for rows.Next() {
		var ticket Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.GuildID,
			&ticket.UserID,
			&ticket.ChannelID,
			&ticket.Title,
			&ticket.Status,
			&ticket.CreatedAt,
			&ticket.ClosedAt,
			&ticket.ClosedBy,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}
	
	return tickets, nil
}

func (d *Database) CloseTicket(channelID, closedBy string) error {
	query := `UPDATE tickets SET status = 'closed', closed_at = CURRENT_TIMESTAMP, closed_by = ?
			  WHERE channel_id = ?`
	
	_, err := d.db.Exec(query, closedBy, channelID)
	return err
}

func (d *Database) CreateTicketPanel(panel *TicketPanel) error {
	query := `INSERT INTO ticket_panels (guild_id, channel_id, message_id, title, description)
			  VALUES (?, ?, ?, ?, ?)`
	
	result, err := d.db.Exec(query,
		panel.GuildID,
		panel.ChannelID,
		panel.MessageID,
		panel.Title,
		panel.Description,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	panel.ID = id
	return nil
}

func (d *Database) GetTicketPanelByMessage(messageID string) (*TicketPanel, error) {
	query := `SELECT id, guild_id, channel_id, message_id, title, description, created_at
			  FROM ticket_panels WHERE message_id = ?`
	
	var panel TicketPanel
	row := d.db.QueryRow(query, messageID)
	err := row.Scan(
		&panel.ID,
		&panel.GuildID,
		&panel.ChannelID,
		&panel.MessageID,
		&panel.Title,
		&panel.Description,
		&panel.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &panel, nil
}