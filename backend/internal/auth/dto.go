package auth

import "time"

// --- Requests ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IIN      string `json:"iin"`
	Role     string `json:"role"`
	Language string `json:"language"`
}

type RegisterOrgRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgName  string `json:"org_name"`
	Language string `json:"language"`
}

// --- Responses ---

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IIN       string    `json:"iin,omitempty"`
	Role      string    `json:"role"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Validation ---

func (r LoginRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Username == "" {
		errs["username"] = "username is required"
	}
	if r.Password == "" {
		errs["password"] = "password is required"
	}
	return errs
}

func (r RegisterRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Username == "" {
		errs["username"] = "username is required"
	}
	if len(r.Username) < 3 {
		errs["username"] = "username must be at least 3 characters"
	}
	if r.Password == "" {
		errs["password"] = "password is required"
	}
	if len(r.Password) < 8 {
		errs["password"] = "password must be at least 8 characters"
	}
	if r.Role == "" {
		errs["role"] = "role is required"
	}
	if r.Role != "" && r.Role != "student" && r.Role != "teacher" {
		errs["role"] = "role must be student or teacher"
	}
	if r.IIN != "" && len(r.IIN) != 12 {
		errs["iin"] = "IIN must be exactly 12 digits"
	}
	if r.Language == "" {
		r.Language = "kz"
	}
	return errs
}

func (r RegisterOrgRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Username == "" {
		errs["username"] = "username is required"
	}
	if len(r.Username) < 3 {
		errs["username"] = "username must be at least 3 characters"
	}
	if r.Email == "" {
		errs["email"] = "email is required"
	}
	if r.Password == "" {
		errs["password"] = "password is required"
	}
	if len(r.Password) < 8 {
		errs["password"] = "password must be at least 8 characters"
	}
	if r.OrgName == "" {
		errs["org_name"] = "organization name is required"
	}
	return errs
}
