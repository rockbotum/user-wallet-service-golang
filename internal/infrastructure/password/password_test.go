package password

import "testing"

func TestHashAndCheck(t *testing.T) {
	hash, err := Hash("mypassword")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty string")
	}

	if err := Check("mypassword", hash); err != nil {
		t.Fatalf("Check() error = %v, want nil", err)
	}
}

func TestCheck_WrongPassword(t *testing.T) {
	hash, err := Hash("mypassword")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if err := Check("wrongpassword", hash); err == nil {
		t.Fatal("Check() error = nil, want error for wrong password")
	}
}

func TestHash_DifferentHashes(t *testing.T) {
	h1, _ := Hash("mypassword")
	h2, _ := Hash("mypassword")

	if h1 == h2 {
		t.Fatal("two hashes of the same password should not be equal (bcrypt uses random salt)")
	}
}

func TestHash_EmptyPassword(t *testing.T) {
	hash, err := Hash("")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if err := Check("", hash); err != nil {
		t.Fatalf("Check() error = %v for empty password, want nil", err)
	}
}

func TestHash_SpecialCharacters(t *testing.T) {
	pwd := "P@$$w0rd!#%"
	hash, err := Hash(pwd)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if err := Check(pwd, hash); err != nil {
		t.Fatalf("Check() error = %v, want nil", err)
	}
}
