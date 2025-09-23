package models

import (
	"gpt-load/internal/types"
	"time"

	"gorm.io/datatypes"
)

// Key状态
const (
	KeyStatusActive  = "active"
	KeyStatusInvalid = "invalid"
)

// 用户角色常量
const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleViewer = "viewer"
)

// 用户状态常量
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// User 用户模型
type User struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username       string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"username"`
	Email          string    `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	PasswordHash   string    `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName    string    `gorm:"type:varchar(100)" json:"display_name"`
	Role           string    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	Status         string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Avatar         string    `gorm:"type:varchar(500)" json:"avatar"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	LastLoginIP    string    `gorm:"type:varchar(45)" json:"last_login_ip"`
	LoginCount     int64     `gorm:"default:0" json:"login_count"`
	FailedAttempts int       `gorm:"default:0" json:"failed_attempts"`
	LockedUntil    *time.Time `json:"locked_until"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联关系
	UserGroups []UserGroup `gorm:"foreignKey:UserID" json:"user_groups,omitempty"`
	Sessions   []UserSession `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
}

// UserGroup 用户分组关联 - 多对多关系
type UserGroup struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	GroupID   uint      `gorm:"not null;index" json:"group_id"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Group Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

// UserSession 用户会话管理
type UserSession struct {
	ID        string    `gorm:"type:varchar(128);primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserActivity 用户活动日志
type UserActivity struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint            `gorm:"not null;index" json:"user_id"`
	Action      string          `gorm:"type:varchar(100);not null" json:"action"`
	Resource    string          `gorm:"type:varchar(100)" json:"resource"`
	ResourceID  string          `gorm:"type:varchar(100)" json:"resource_id"`
	IPAddress   string          `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent   string          `gorm:"type:text" json:"user_agent"`
	Details     datatypes.JSON  `gorm:"type:json" json:"details"`
	Timestamp   time.Time       `gorm:"not null;index" json:"timestamp"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// SystemSetting 对应 system_settings 表
type SystemSetting struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SettingKey   string    `gorm:"type:varchar(255);not null;unique" json:"setting_key"`
	SettingValue string    `gorm:"type:text;not null" json:"setting_value"`
	Description  string    `gorm:"type:varchar(512)" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GroupConfig 存储特定于分组的配置
type GroupConfig struct {
	RequestTimeout               *int    `json:"request_timeout,omitempty"`
	IdleConnTimeout              *int    `json:"idle_conn_timeout,omitempty"`
	ConnectTimeout               *int    `json:"connect_timeout,omitempty"`
	MaxIdleConns                 *int    `json:"max_idle_conns,omitempty"`
	MaxIdleConnsPerHost          *int    `json:"max_idle_conns_per_host,omitempty"`
	ResponseHeaderTimeout        *int    `json:"response_header_timeout,omitempty"`
	ProxyURL                     *string `json:"proxy_url,omitempty"`
	MaxRetries                   *int    `json:"max_retries,omitempty"`
	BlacklistThreshold           *int    `json:"blacklist_threshold,omitempty"`
	KeyValidationIntervalMinutes *int    `json:"key_validation_interval_minutes,omitempty"`
	KeyValidationConcurrency     *int    `json:"key_validation_concurrency,omitempty"`
	KeyValidationTimeoutSeconds  *int    `json:"key_validation_timeout_seconds,omitempty"`
	EnableRequestBodyLogging     *bool   `json:"enable_request_body_logging,omitempty"`
	EnableResponseBodyLogging    *bool   `json:"enable_response_body_logging,omitempty"`
}

// HeaderRule defines a single rule for header manipulation.
type HeaderRule struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Action string `json:"action"` // "set" or "remove"
}

// Group 对应 groups 表
type Group struct {
	ID                 uint                 `gorm:"primaryKey;autoIncrement" json:"id"`
	EffectiveConfig    types.SystemSettings `gorm:"-" json:"effective_config,omitempty"`
	Name               string               `gorm:"type:varchar(255);not null;unique" json:"name"`
	Endpoint           string               `gorm:"-" json:"endpoint"`
	DisplayName        string               `gorm:"type:varchar(255)" json:"display_name"`
	ProxyKeys          string               `gorm:"type:text" json:"proxy_keys"`
	Description        string               `gorm:"type:varchar(512)" json:"description"`
	Upstreams          datatypes.JSON       `gorm:"type:json;not null" json:"upstreams"`
	ValidationEndpoint string               `gorm:"type:varchar(255)" json:"validation_endpoint"`
	ChannelType        string               `gorm:"type:varchar(50);not null" json:"channel_type"`
	Sort               int                  `gorm:"default:0" json:"sort"`
	TestModel          string               `gorm:"type:varchar(255);not null" json:"test_model"`
	ParamOverrides     datatypes.JSONMap    `gorm:"type:json" json:"param_overrides"`
	Config             datatypes.JSONMap    `gorm:"type:json" json:"config"`
	HeaderRules        datatypes.JSON       `gorm:"type:json" json:"header_rules"`
	APIKeys            []APIKey             `gorm:"foreignKey:GroupID" json:"api_keys"`
	LastValidatedAt    *time.Time           `json:"last_validated_at"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`

	// For cache
	ProxyKeysMap   map[string]struct{} `gorm:"-" json:"-"`
	HeaderRuleList []HeaderRule        `gorm:"-" json:"-"`
}

// APIKey 对应 api_keys 表
type APIKey struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyValue     string     `gorm:"type:text;not null" json:"key_value"`
	KeyHash      string     `gorm:"type:varchar(128);index" json:"key_hash"`
	GroupID      uint       `gorm:"not null;index" json:"group_id"`
	Status       string     `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	RequestCount int64      `gorm:"not null;default:0" json:"request_count"`
	FailureCount int64      `gorm:"not null;default:0" json:"failure_count"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RequestType 请求类型常量
const (
	RequestTypeRetry = "retry"
	RequestTypeFinal = "final"
)

// RequestLog 对应 request_logs 表
type RequestLog struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Timestamp    time.Time `gorm:"not null;index" json:"timestamp"`
	GroupID      uint      `gorm:"not null;index" json:"group_id"`
	GroupName    string    `gorm:"type:varchar(255);index" json:"group_name"`
	KeyValue     string    `gorm:"type:text" json:"key_value"`
	KeyHash      string    `gorm:"type:varchar(128);index" json:"key_hash"`
	Model        string    `gorm:"type:varchar(255);index" json:"model"`
	IsSuccess    bool      `gorm:"not null" json:"is_success"`
	SourceIP     string    `gorm:"type:varchar(64)" json:"source_ip"`
	StatusCode   int       `gorm:"not null" json:"status_code"`
	RequestPath  string    `gorm:"type:varchar(500)" json:"request_path"`
	Duration     int64     `gorm:"not null" json:"duration_ms"`
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	UserAgent    string    `gorm:"type:varchar(512)" json:"user_agent"`
	RequestType  string    `gorm:"type:varchar(20);not null;default:'final';index" json:"request_type"`
	UpstreamAddr string    `gorm:"type:varchar(500)" json:"upstream_addr"`
	IsStream     bool      `gorm:"not null" json:"is_stream"`
	RequestBody  string    `gorm:"type:text" json:"request_body"`
	ResponseBody string    `gorm:"type:text" json:"response_body"`

	// Token usage fields - 支持各种 token 类型
	PromptTokens           *int `gorm:"default:null" json:"prompt_tokens,omitempty"`
	CompletionTokens       *int `gorm:"default:null" json:"completion_tokens,omitempty"`
	TotalTokens            *int `gorm:"default:null" json:"total_tokens,omitempty"`

	// 缓存相关 token (适用于支持缓存的模型)
	CachedPromptTokens     *int `gorm:"default:null" json:"cached_prompt_tokens,omitempty"`
	CachedCompletionTokens *int `gorm:"default:null" json:"cached_completion_tokens,omitempty"`

	// 推理相关 token (适用于 Claude 等推理模型)
	ReasoningTokens        *int `gorm:"default:null" json:"reasoning_tokens,omitempty"`

	// 音频相关 token (适用于语音模型)
	AudioTokens            *int `gorm:"default:null" json:"audio_tokens,omitempty"`

	// 图像相关 token (适用于视觉模型)
	ImageTokens            *int `gorm:"default:null" json:"image_tokens,omitempty"`
}

// StatCard 用于仪表盘的单个统计卡片数据
type StatCard struct {
	Value         float64 `json:"value"`
	SubValue      int64   `json:"sub_value,omitempty"`
	SubValueTip   string  `json:"sub_value_tip,omitempty"`
	Trend         float64 `json:"trend"`
	TrendIsGrowth bool    `json:"trend_is_growth"`
}

// SecurityWarning 用于安全警告信息
type SecurityWarning struct {
	Type       string `json:"type"`       // 警告类型：auth_key, encryption_key 等
	Message    string `json:"message"`    // 警告信息
	Severity   string `json:"severity"`   // 严重程度：low, medium, high
	Suggestion string `json:"suggestion"` // 建议解决方案
}

// DashboardStatsResponse 用于仪表盘基础统计的API响应
type DashboardStatsResponse struct {
	KeyCount         StatCard          `json:"key_count"`
	RPM              StatCard          `json:"rpm"`
	RequestCount     StatCard          `json:"request_count"`
	ErrorRate        StatCard          `json:"error_rate"`
	SecurityWarnings []SecurityWarning `json:"security_warnings"`
}

// ChartDataset 用于图表的数据集
type ChartDataset struct {
	Label string  `json:"label"`
	Data  []int64 `json:"data"`
	Color string  `json:"color"`
}

// ChartData 用于图表的API响应
type ChartData struct {
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

// GroupHourlyStat 对应 group_hourly_stats 表，用于存储每个分组每小时的请求统计
type GroupHourlyStat struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Time         time.Time `gorm:"not null;uniqueIndex:idx_group_time" json:"time"` // 整点时间
	GroupID      uint      `gorm:"not null;uniqueIndex:idx_group_time" json:"group_id"`
	SuccessCount int64     `gorm:"not null;default:0" json:"success_count"`
	FailureCount int64     `gorm:"not null;default:0" json:"failure_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
