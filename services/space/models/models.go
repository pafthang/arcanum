package models

// User is a platform principal. Password hash never leaves the store.
type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Actor         string `json:"actor"`
	PlatformAdmin bool   `json:"platformAdmin"`
	CreatedAt     string `json:"createdAt"`
}

// Space is a tenant / workspace. Owner lives on space_members, not here.
type Space struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Member is a user's role inside a space.
type Member struct {
	SpaceID   string `json:"spaceId"`
	UserID    string `json:"userId"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
}

// SpaceWithRole is a space plus the calling user's membership role.
type SpaceWithRole struct {
	Space
	Role string `json:"role"`
}

// Team is a nested group inside a space.
type Team struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	ParentID  string `json:"parentId,omitempty"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

// TeamMember is a user's role inside a team.
type TeamMember struct {
	TeamID string `json:"teamId"`
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// APIKey is the public view of an agent key. The raw secret is never stored.
type APIKey struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	CreatedAt string `json:"createdAt"`
}

// LoginRequest is POST /api/auth/login body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned after a successful password login.
type LoginResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	Space     *Space `json:"space,omitempty"`
	SpaceRole string `json:"spaceRole,omitempty"`
}

// CreateSpaceRequest is POST /api/spaces body.
type CreateSpaceRequest struct {
	Name string `json:"name"`
}

// RegisterRequest is POST /api/auth/register body.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// SwitchSpaceRequest is POST /api/auth/switch-space body.
type SwitchSpaceRequest struct {
	SpaceID string `json:"spaceId"`
}

// APIKeyAuthRequest is POST /api/auth/api-key body.
type APIKeyAuthRequest struct {
	Secret string `json:"secret"`
}

// InviteMemberRequest is POST /api/spaces/{spaceId}/members body.
type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UpdateMemberRequest is PATCH /api/spaces/{spaceId}/members/{userId} body.
type UpdateMemberRequest struct {
	Role string `json:"role"`
}

// CreateTeamRequest is POST /api/spaces/{spaceId}/teams body.
type CreateTeamRequest struct {
	Name     string `json:"name"`
	ParentID string `json:"parentId"`
}

// UpdateTeamRequest is PATCH .../teams/{teamId} body.
type UpdateTeamRequest struct {
	Name string `json:"name"`
}

// AddTeamMemberRequest is POST .../teams/{teamId}/members body.
type AddTeamMemberRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// CreateAPIKeyRequest is POST /api/spaces/{spaceId}/keys body.
type CreateAPIKeyRequest struct {
	Email string `json:"email"`
}

// CreateAPIKeyResponse returns the secret once.
type CreateAPIKeyResponse struct {
	APIKey
	User   User   `json:"user"`
	Secret string `json:"secret"`
}

// CanRequest is internal.space.can payload.
type CanRequest struct {
	UserID  string `json:"userId"`
	SpaceID string `json:"spaceId"`
	Perm    string `json:"perm"`
}

// CanResponse is internal.space.can result.
type CanResponse struct {
	OK   bool   `json:"ok"`
	Role string `json:"role,omitempty"`
}
