package audit

import "time"

// Actions
const (
	ActionEventCreate         = "event.create"
	ActionEventUpdate         = "event.update"
	ActionEventDelete         = "event.delete"
	ActionTemplateUpload      = "template.upload"
	ActionTemplateDelete      = "template.delete"
	ActionBatchUpload         = "batch.upload"
	ActionBatchMapping        = "batch.mapping"
	ActionBatchGenerate       = "batch.generate"
	ActionBatchDelete         = "batch.delete"
	ActionCertificateRevoke   = "certificate.revoke"
	ActionCertificateUnrevoke = "certificate.unrevoke"
	ActionCertificateDelete   = "certificate.delete"
	ActionOrgCreate           = "organization.create"
	ActionOrgUpdate           = "organization.update"
	ActionOrgDelete           = "organization.delete"
	ActionMemberAdd           = "member.add"
	ActionMemberRemove        = "member.remove"
)

// --- Responses ---

type AuditLogResponse struct {
	ID            int64                  `json:"id"`
	ActorID       *int64                 `json:"actor_id,omitempty"`
	ActorUsername string                 `json:"actor_username,omitempty"`
	Action        string                 `json:"action"`
	ObjectType    string                 `json:"object_type,omitempty"`
	ObjectID      string                 `json:"object_id,omitempty"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}
