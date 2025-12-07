package mssql

import ()

type DatabaseCredentials struct {
	Server   string `json:"server"`
	Database string `json:"database"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}
