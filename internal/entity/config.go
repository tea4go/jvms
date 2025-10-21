package entity

// TConfig 配置结构体，用于存储JVMS的全局配置信息
type TConfig struct {
	// JavaHome Java环境变量路径
	JavaHome string `json:"java_home"`
	// CurrentJDKVersion 当前使用的JDK版本
	CurrentJDKVersion string `json:"current_jdk_version"`
	// JDK下载来源
	WebType string `json:"web_type"`
	// 是否全部JDK，如果为否，则一个大版本只保留一个
	WebAll bool `json:"web_all"`
	// Proxy 代理服务器地址
	Proxy string `json:"proxy"`
	// Store JDK存储路径
	Store string
	// Download JDK下载路径
	Download string
	// 日志级别
	LogLevel int
}
