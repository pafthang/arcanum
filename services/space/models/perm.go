package models

// Identity-scoped permissions. Issue/machine perms stay in their services.
const (
	PermWorkspaceManage   = "workspace:manage"
	PermWorkspaceTransfer = "workspace:transfer"
	PermMemberInvite      = "member:invite"
	PermMemberRole        = "member:role"
	PermTeamManage        = "team:manage"
)

var rolePerms = map[string][]string{
	"owner": {
		PermWorkspaceManage, PermWorkspaceTransfer,
		PermMemberInvite, PermMemberRole, PermTeamManage,
	},
	"admin": {
		PermWorkspaceManage, PermWorkspaceTransfer,
		PermMemberInvite, PermMemberRole, PermTeamManage,
	},
	"member": {},
	"guest":  {},
}

// RoleHas reports whether a space role includes perm.
func RoleHas(role, perm string) bool {
	for _, p := range rolePerms[role] {
		if p == perm {
			return true
		}
	}
	return false
}
