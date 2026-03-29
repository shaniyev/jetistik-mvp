package user

import "time"

// --- Requests ---

type UpdateProfileRequest struct {
	Email    *string `json:"email"`
	IIN      *string `json:"iin"`
	Language *string `json:"language"`
}

type AddStudentRequest struct {
	StudentIIN string `json:"student_iin"`
}

// --- Responses ---

type ProfileResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	IIN       string    `json:"iin,omitempty"`
	Role      string    `json:"role"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

type TeacherStudentResponse struct {
	ID         int64     `json:"id"`
	TeacherID  int64     `json:"teacher_id"`
	StudentIIN string    `json:"student_iin"`
	CreatedAt  time.Time `json:"created_at"`
}

// --- Validation ---

func (r UpdateProfileRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.IIN != nil && *r.IIN != "" && len(*r.IIN) != 12 {
		errs["iin"] = "IIN must be exactly 12 digits"
	}
	if r.Language != nil && *r.Language != "" {
		if *r.Language != "kz" && *r.Language != "ru" && *r.Language != "en" {
			errs["language"] = "language must be kz, ru, or en"
		}
	}
	return errs
}

func (r AddStudentRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.StudentIIN == "" {
		errs["student_iin"] = "student IIN is required"
	}
	if r.StudentIIN != "" && len(r.StudentIIN) != 12 {
		errs["student_iin"] = "IIN must be exactly 12 digits"
	}
	return errs
}
