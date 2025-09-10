package hashing

import (
	"github.com/alexedwards/argon2id"
)

var params = &argon2id.Params{
	Memory: 19 * 1024,
	Iterations: 2,
	Parallelism: 1,
	SaltLength: 16,
	KeyLength: 32,
}

func HashString(value string) (string, error) {
	return argon2id.CreateHash(value, params)
}

func CompareStringAndHash(hash string, value string) (bool, error) {
	return argon2id.ComparePasswordAndHash(value, hash)
}