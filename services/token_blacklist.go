package services

var TokenBlacklist = make(map[string]bool)

// add token ke blacklist
func BlacklistToken(token string) {
	TokenBlacklist[token] = true
}

// cek token blacklist
func IsTokenBlacklisted(token string) bool {
	return TokenBlacklist[token]
}
