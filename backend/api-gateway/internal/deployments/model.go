package deployments

import (
	"strings"
	"time"
)

type Deployment struct {
	ID          string    `json:"id"`
	AppName     string    `json:"app_name"`
	Image       string    `json:"image"`
	Namespace   string    `json:"namespace"`
	Replicas    int       `json:"replicas"`
	CPULimit    string    `json:"cpu_limit,omitempty"`
	MemoryLimit string    `json:"memory_limit,omitempty"`
	Privileged  bool      `json:"privileged"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PolicyViolation struct {
	ControlNo string `json:"control_no"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

type CreateRequest struct {
	AppName     string `json:"app_name"`
	Image       string `json:"image"`
	Namespace   string `json:"namespace"`
	Replicas    int    `json:"replicas"`
	CPULimit    string `json:"cpu_limit,omitempty"`
	MemoryLimit string `json:"memory_limit,omitempty"`
	Privileged  bool   `json:"privileged"`
}

// Normalize trims incoming request fields so validation and persistence use clean values.
func (r CreateRequest) Normalize() CreateRequest {
	r.AppName = strings.TrimSpace(r.AppName)
	r.Image = strings.TrimSpace(r.Image)
	r.Namespace = strings.TrimSpace(r.Namespace)
	r.CPULimit = strings.TrimSpace(r.CPULimit)
	r.MemoryLimit = strings.TrimSpace(r.MemoryLimit)

	return r
}

// Validate checks whether the minimum required deployment fields are present and valid.
func (r CreateRequest) Validate() map[string]string {
	issues := map[string]string{}

	if r.AppName == "" {
		issues["app_name"] = "app_name is required"
	}
	if r.Image == "" {
		issues["image"] = "image is required"
	}
	if r.Namespace == "" {
		issues["namespace"] = "namespace is required"
	}
	if r.Replicas < 1 {
		issues["replicas"] = "replicas must be greater than 0"
	}

	if len(issues) == 0 {
		return nil
	}

	return issues
}
