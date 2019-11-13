package main

// StatsDConfig defines our statsd destination
type StatsDConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Prefix   string `json:"prefix"`
}
