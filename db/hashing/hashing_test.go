package hashing

import (
	"testing"
)


func TestHashing(t *testing.T) {
	passwords := []string{
		"123456",
		"ew3of#kjSJKL1JR$78",
		"ehflwfHLWKHDLÖWDHLwklhd3546",
		"E9O:o6q|DS2#Q7?!@LA8CKG`'3FowbvbD0l&>#7",
		"MamXFS8eY6{AE9dBt4v-qIV0[|HEr=5kPju0<gD",
	}

	for _, password := range passwords {
		hash, err := HashString(password)
		if err != nil {
			t.Errorf("=> HashString method failed for password %v: %v", password, err.Error())
		}

		passwordMatch, err := CompareStringAndHash(hash, password)
		if err != nil {
			t.Errorf("=> Hash and password string %v could not be compared: %v", password, err.Error())
		}

		if passwordMatch == false {
			t.Errorf("=> password %v did not match with its hashed value", password)
		}
	}
}