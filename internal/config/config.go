package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logging  LoggingConfig  `yaml:"logging"`
	Remote   RemoteConfig   `yaml:"remote"`
	Tools    ToolsConfig    `yaml:"tools"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	File   string `yaml:"file"`
}

// RemoteConfig 远程服务配置
type RemoteConfig struct {
	DefaultTimeout         time.Duration `yaml:"default_timeout"`
	DefaultConnectTimeout  time.Duration `yaml:"default_connect_timeout"`
	DefaultRetryAttempts   int           `yaml:"default_retry_attempts"`
	DefaultRetryDelay      time.Duration `yaml:"default_retry_delay"`
	SessionCleanupInterval time.Duration `yaml:"session_cleanup_interval"`
	DefaultIdleTTL         time.Duration `yaml:"default_idle_ttl"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	BuiltinEcho   bool `yaml:"builtin_echo"`
	BuiltinGreet  bool `yaml:"builtin_greet"`
	BuiltinStatus bool `yaml:"builtin_status"`
}

// GetDSN 获取数据库连接字符串
func (db *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.Username, db.Password, db.Database, db.SSLMode)
}

// GetServerAddr 获取服务器监听地址
func (s *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetMaxOpenConns 获取最大打开连接数
func (db *DatabaseConfig) GetMaxOpenConns() int {
	return db.MaxOpenConns
}

// GetMaxIdleConns 获取最大空闲连接数
func (db *DatabaseConfig) GetMaxIdleConns() int {
	return db.MaxIdleConns
}

// GetConnMaxLifetime 获取连接最大生存时间
func (db *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return db.ConnMaxLifetime
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果没有指定配置文件路径，使用默认路径
	if configPath == "" {
		configPath = "config/config.dev.yaml"
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	// 服务器默认值
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 9001
	}

	// 数据库默认值
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 5 * time.Minute
	}

	// 日志默认值
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}

	// 远程服务默认值
	if config.Remote.DefaultTimeout == 0 {
		config.Remote.DefaultTimeout = 30 * time.Second
	}
	if config.Remote.DefaultConnectTimeout == 0 {
		config.Remote.DefaultConnectTimeout = 10 * time.Second
	}
	if config.Remote.DefaultRetryAttempts == 0 {
		config.Remote.DefaultRetryAttempts = 3
	}
	if config.Remote.DefaultRetryDelay == 0 {
		config.Remote.DefaultRetryDelay = 3 * time.Second
	}
	if config.Remote.SessionCleanupInterval == 0 {
		config.Remote.SessionCleanupInterval = 30 * time.Second
	}
	if config.Remote.DefaultIdleTTL == 0 {
		config.Remote.DefaultIdleTTL = 5 * time.Minute
	}
}

// LoadConfigFromEnv 从环境变量加载配置（优先级高于配置文件）
func LoadConfigFromEnv(config *Config) {
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Database.Port)
	}
	if username := os.Getenv("DB_USERNAME"); username != "" {
		config.Database.Username = username
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if database := os.Getenv("DB_DATABASE"); database != "" {
		config.Database.Database = database
	}

	// 服务器配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Server.Port)
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
}
