package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Discord ID validation patterns
var (
	// Discord IDs are 17-19 digit numbers (Snowflake format)
	discordIDPattern = regexp.MustCompile(`^\d{17,19}$`)
	
	// Discord mention patterns
	userMentionPattern    = regexp.MustCompile(`^<@!?(\d{17,19})>$`)
	roleMentionPattern    = regexp.MustCompile(`^<@&(\d{17,19})>$`)
	channelMentionPattern = regexp.MustCompile(`^<#(\d{17,19})>$`)
	
	// Malicious content patterns
	maliciousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>`),                    // Script tags
		regexp.MustCompile(`(?i)javascript:`),                     // JavaScript protocol
		regexp.MustCompile(`(?i)data:(?:text\/html|application\/)`), // Data URLs
		regexp.MustCompile(`(?i)vbscript:`),                       // VBScript
		regexp.MustCompile(`(?i)on\w+\s*=`),                      // Event handlers
	}
)

// ValidateDiscordID validates a Discord Snowflake ID
func ValidateDiscordID(id string) bool {
	if id == "" {
		return false
	}
	return discordIDPattern.MatchString(id)
}

// ExtractUserIDFromMention extracts user ID from Discord mention
// Returns the ID if valid, empty string if invalid
func ExtractUserIDFromMention(mention string) string {
	matches := userMentionPattern.FindStringSubmatch(mention)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

// ExtractRoleIDFromMention extracts role ID from Discord mention  
func ExtractRoleIDFromMention(mention string) string {
	matches := roleMentionPattern.FindStringSubmatch(mention)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

// ExtractChannelIDFromMention extracts channel ID from Discord mention
func ExtractChannelIDFromMention(mention string) string {
	matches := channelMentionPattern.FindStringSubmatch(mention)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

// ValidateMessageContent validates Discord message content
func ValidateMessageContent(content string) error {
	if content == "" {
		return nil // Empty content is allowed
	}
	
	// Check length (Discord limit: 2000 characters)
	if utf8.RuneCountInString(content) > 2000 {
		return NewValidationError("message content exceeds 2000 characters")
	}
	
	// Check for malicious content
	if ContainsMaliciousContent(content) {
		return NewValidationError("message contains potentially malicious content")
	}
	
	return nil
}

// ContainsMaliciousContent checks for potentially malicious patterns
func ContainsMaliciousContent(content string) bool {
	for _, pattern := range maliciousPatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// SanitizeInput removes potentially dangerous characters and patterns
func SanitizeInput(input string) string {
	// Remove control characters except newlines and tabs
	result := strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			return -1 // Remove character
		}
		return r
	}, input)
	
	// Trim whitespace
	result = strings.TrimSpace(result)
	
	return result
}

// ValidateUsername validates Discord username format
func ValidateUsername(username string) error {
	if username == "" {
		return NewValidationError("username cannot be empty")
	}
	
	// Discord username length: 2-32 characters
	length := utf8.RuneCountInString(username)
	if length < 2 || length > 32 {
		return NewValidationError("username must be between 2-32 characters")
	}
	
	// Check for forbidden patterns
	forbidden := []string{
		"discord",
		"```", // Code blocks
		"@everyone",
		"@here",
	}
	
	lowerUsername := strings.ToLower(username)
	for _, pattern := range forbidden {
		if strings.Contains(lowerUsername, pattern) {
			return NewValidationError("username contains forbidden pattern: " + pattern)
		}
	}
	
	return nil
}

// ValidateGuildID validates Discord guild (server) ID
func ValidateGuildID(guildID string) bool {
	return ValidateDiscordID(guildID)
}

// ValidateChannelID validates Discord channel ID
func ValidateChannelID(channelID string) bool {
	return ValidateDiscordID(channelID)
}

// ValidateRoleID validates Discord role ID
func ValidateRoleID(roleID string) bool {
	return ValidateDiscordID(roleID)
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return "validation error in field '" + e.Field + "': " + e.Message
	}
	return "validation error: " + e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// NewFieldValidationError creates a new validation error for a specific field
func NewFieldValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// ValidateCommandArgs validates command arguments
func ValidateCommandArgs(args []string, maxArgs int) error {
	if len(args) > maxArgs {
		return NewValidationError("too many arguments provided")
	}
	
	// Validate each argument
	for i, arg := range args {
		if err := ValidateMessageContent(arg); err != nil {
			return NewFieldValidationError("arg_"+string(rune(i+'0')), err.Error())
		}
	}
	
	return nil
}

// SafeString returns a safe version of the input string for logging/display
func SafeString(input string, maxLength int) string {
	// Sanitize first
	safe := SanitizeInput(input)
	
	// Truncate if too long
	if utf8.RuneCountInString(safe) > maxLength {
		runes := []rune(safe)
		safe = string(runes[:maxLength]) + "..."
	}
	
	return safe
}