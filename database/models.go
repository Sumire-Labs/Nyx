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

type LogSettings struct {
	GuildID           string    `json:"guild_id"`
	LogChannelID      *string   `json:"log_channel_id"`
	Enabled           bool      `json:"enabled"`
	LogMemberJoin     bool      `json:"log_member_join"`
	LogMemberLeave    bool      `json:"log_member_leave"`
	LogMessageEdit    bool      `json:"log_message_edit"`
	LogMessageDelete  bool      `json:"log_message_delete"`
	LogRoleCreate     bool      `json:"log_role_create"`
	LogRoleUpdate     bool      `json:"log_role_update"`
	LogRoleDelete     bool      `json:"log_role_delete"`
	LogNicknameChange bool      `json:"log_nickname_change"`
	LogBan            bool      `json:"log_ban"`
	LogUnban          bool      `json:"log_unban"`
	LogKick           bool      `json:"log_kick"`
	LogTimeout        bool      `json:"log_timeout"`
	LogChannelCreate  bool      `json:"log_channel_create"`
	LogChannelUpdate  bool      `json:"log_channel_update"`
	LogChannelDelete  bool      `json:"log_channel_delete"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ServerLog struct {
	ID         int64      `json:"id"`
	GuildID    string     `json:"guild_id"`
	EventType  string     `json:"event_type"`
	UserID     *string    `json:"user_id"`
	TargetID   *string    `json:"target_id"`
	ChannelID  *string    `json:"channel_id"`
	RoleID     *string    `json:"role_id"`
	Reason     *string    `json:"reason"`
	OldContent *string    `json:"old_content"`
	NewContent *string    `json:"new_content"`
	Metadata   *string    `json:"metadata"`
	Timestamp  time.Time  `json:"timestamp"`
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

func (d *Database) GetLogSettings(guildID string) (*LogSettings, error) {
	query := `SELECT guild_id, log_channel_id, enabled, log_member_join, log_member_leave,
			  log_message_edit, log_message_delete, log_role_create, log_role_update, log_role_delete,
			  log_nickname_change, log_ban, log_unban, log_kick, log_timeout,
			  log_channel_create, log_channel_update, log_channel_delete, created_at, updated_at
			  FROM log_settings WHERE guild_id = ?`
	
	var settings LogSettings
	row := d.db.QueryRow(query, guildID)
	err := row.Scan(
		&settings.GuildID,
		&settings.LogChannelID,
		&settings.Enabled,
		&settings.LogMemberJoin,
		&settings.LogMemberLeave,
		&settings.LogMessageEdit,
		&settings.LogMessageDelete,
		&settings.LogRoleCreate,
		&settings.LogRoleUpdate,
		&settings.LogRoleDelete,
		&settings.LogNicknameChange,
		&settings.LogBan,
		&settings.LogUnban,
		&settings.LogKick,
		&settings.LogTimeout,
		&settings.LogChannelCreate,
		&settings.LogChannelUpdate,
		&settings.LogChannelDelete,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &settings, nil
}

func (d *Database) CreateOrUpdateLogSettings(settings *LogSettings) error {
	query := `INSERT OR REPLACE INTO log_settings 
			  (guild_id, log_channel_id, enabled, log_member_join, log_member_leave,
			   log_message_edit, log_message_delete, log_role_create, log_role_update, log_role_delete,
			   log_nickname_change, log_ban, log_unban, log_kick, log_timeout,
			   log_channel_create, log_channel_update, log_channel_delete, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	
	_, err := d.db.Exec(query,
		settings.GuildID,
		settings.LogChannelID,
		settings.Enabled,
		settings.LogMemberJoin,
		settings.LogMemberLeave,
		settings.LogMessageEdit,
		settings.LogMessageDelete,
		settings.LogRoleCreate,
		settings.LogRoleUpdate,
		settings.LogRoleDelete,
		settings.LogNicknameChange,
		settings.LogBan,
		settings.LogUnban,
		settings.LogKick,
		settings.LogTimeout,
		settings.LogChannelCreate,
		settings.LogChannelUpdate,
		settings.LogChannelDelete,
	)
	
	return err
}

func (d *Database) LogServerEvent(log *ServerLog) error {
	query := `INSERT INTO server_logs 
			  (guild_id, event_type, user_id, target_id, channel_id, role_id, reason, old_content, new_content, metadata)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := d.db.Exec(query,
		log.GuildID,
		log.EventType,
		log.UserID,
		log.TargetID,
		log.ChannelID,
		log.RoleID,
		log.Reason,
		log.OldContent,
		log.NewContent,
		log.Metadata,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	log.ID = id
	return nil
}

func (d *Database) GetRecentServerLogs(guildID, eventType string, limit int) ([]*ServerLog, error) {
	var query string
	var args []interface{}
	
	if eventType != "" {
		query = `SELECT id, guild_id, event_type, user_id, target_id, channel_id, role_id, reason, 
				 old_content, new_content, metadata, timestamp
				 FROM server_logs WHERE guild_id = ? AND event_type = ?
				 ORDER BY timestamp DESC LIMIT ?`
		args = []interface{}{guildID, eventType, limit}
	} else {
		query = `SELECT id, guild_id, event_type, user_id, target_id, channel_id, role_id, reason,
				 old_content, new_content, metadata, timestamp
				 FROM server_logs WHERE guild_id = ?
				 ORDER BY timestamp DESC LIMIT ?`
		args = []interface{}{guildID, limit}
	}
	
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []*ServerLog
	for rows.Next() {
		var log ServerLog
		err := rows.Scan(
			&log.ID,
			&log.GuildID,
			&log.EventType,
			&log.UserID,
			&log.TargetID,
			&log.ChannelID,
			&log.RoleID,
			&log.Reason,
			&log.OldContent,
			&log.NewContent,
			&log.Metadata,
			&log.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}
	
	return logs, nil
}

// DatabaseService インターフェースの追加実装メソッド

func (d *Database) CreateGuild(guild *Guild) error {
	return d.CreateOrUpdateGuild(guild)
}

func (d *Database) UpdateGuild(guild *Guild) error {
	return d.CreateOrUpdateGuild(guild)
}

func (d *Database) DeleteGuild(guildID string) error {
	query := `DELETE FROM guilds WHERE id = ?`
	_, err := d.db.Exec(query, guildID)
	return err
}

func (d *Database) CreateLogSettings(settings *LogSettings) error {
	return d.CreateOrUpdateLogSettings(settings)
}

func (d *Database) UpdateLogSettings(settings *LogSettings) error {
	return d.CreateOrUpdateLogSettings(settings)
}

func (d *Database) DeleteLogSettings(guildID string) error {
	query := `DELETE FROM log_settings WHERE guild_id = ?`
	_, err := d.db.Exec(query, guildID)
	return err
}

func (d *Database) GetServerLogs(guildID string, limit int) ([]*ServerLog, error) {
	return d.GetRecentServerLogs(guildID, "", limit)
}

func (d *Database) GetTicketPanel(messageID string) (*TicketPanel, error) {
	return d.GetTicketPanelByMessage(messageID)
}

func (d *Database) DeleteTicketPanel(messageID string) error {
	query := `DELETE FROM ticket_panels WHERE message_id = ?`
	_, err := d.db.Exec(query, messageID)
	return err
}

func (d *Database) GetAllTicketPanels() ([]*TicketPanel, error) {
	query := `SELECT id, guild_id, channel_id, message_id, title, description, created_at
			  FROM ticket_panels ORDER BY created_at DESC`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var panels []*TicketPanel
	for rows.Next() {
		var panel TicketPanel
		err := rows.Scan(
			&panel.ID,
			&panel.GuildID,
			&panel.ChannelID,
			&panel.MessageID,
			&panel.Title,
			&panel.Description,
			&panel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		panels = append(panels, &panel)
	}
	
	return panels, nil
}

func (d *Database) UpdateTicket(ticket *Ticket) error {
	query := `UPDATE tickets SET title = ?, status = ? WHERE channel_id = ?`
	_, err := d.db.Exec(query, ticket.Title, ticket.Status, ticket.ChannelID)
	return err
}

func (d *Database) DeleteTicket(channelID string) error {
	query := `DELETE FROM tickets WHERE channel_id = ?`
	_, err := d.db.Exec(query, channelID)
	return err
}

func (d *Database) GetUserTickets(userID string) ([]*Ticket, error) {
	query := `SELECT id, guild_id, user_id, channel_id, title, status, created_at, closed_at, closed_by
			  FROM tickets WHERE user_id = ? ORDER BY created_at DESC`
	
	rows, err := d.db.Query(query, userID)
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