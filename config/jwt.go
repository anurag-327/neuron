package config

import "time"

var JwtSecret []byte

const (
	TokenExpirationTime = 7 * 24 * time.Hour
)
