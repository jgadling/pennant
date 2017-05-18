
package main

type StatsDConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Prefix      string `json:"prefix"`
}
