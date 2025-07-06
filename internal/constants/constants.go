package constants

import "time"

// Tag prefixes
const (
	TagPrefix         = "mdtask"
	StatusTagPrefix   = "mdtask/status/"
	ArchivedTag       = "mdtask/archived"
	DeadlineTagPrefix = "mdtask/deadline/"
	WaitForTagPrefix  = "mdtask/waitfor/"
	ReminderTagPrefix = "mdtask/reminder/"
	ParentTagPrefix   = "mdtask/parent/"
)

// Status values
const (
	StatusTODO = "TODO"
	StatusWIP  = "WIP"
	StatusWAIT = "WAIT"
	StatusSCHE = "SCHE"
	StatusDONE = "DONE"
)

// Date and time formats
const (
	DateTimeFormat = "2006-01-02 15:04"
	DateFormat     = "2006-01-02"
	IDTimeFormat   = "20060102150405"
	TaskIDPrefix   = "task/"
)

// File and directory permissions
const (
	DirPermission  = 0755
	FilePermission = 0644
)

// Repository constants
const (
	MaxFilenameSuffix   = 100
	MarkdownExtension   = ".md"
	DefaultSearchPath   = "."
	ConfigFilename      = ".mdtask.toml"
	AltConfigFilename   = "mdtask.toml"
)

// Web server constants
const (
	DefaultWebPort      = 7000
	MaxPortRetries      = 10
	DefaultOpenBrowser  = true
)

// Time constants
const (
	GenerateIDSleepDuration = time.Second
	ReminderCheckInterval   = 5 * time.Minute
	WeekDuration           = 7 * 24 * time.Hour
)