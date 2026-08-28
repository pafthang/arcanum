package apis

import (
	"context"
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/passwd"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/internal/token"
	"github.com/pafthang/arcanum/services/space/models"
)

func registerAuth(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("auth_register", mini.HandlerFunc(func(req mini.Request) {
		var in models.RegisterRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		out, code, err := registerUser(req.Context(), d, in)
		if err != nil {
			httpx.Error(req, code, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, out)
	}),
		mini.WithPublicHTTP("POST", "/api/auth/register"),
		mini.WithPublicSubject("space", "auth.register"),
		mini.WithPublicAuth(mini.AuthNone),
	))

	must(svc.AddEndpoint("auth_api_key", mini.HandlerFunc(func(req mini.Request) {
		var in models.APIKeyAuthRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		out, code, err := loginAPIKey(req.Context(), d, in.Secret)
		if err != nil {
			httpx.Error(req, code, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, out)
	}),
		mini.WithPublicHTTP("POST", "/api/auth/api-key"),
		mini.WithPublicSubject("space", "auth.api_key"),
		mini.WithPublicAuth(mini.AuthNone),
	))

	must(svc.AddEndpoint("auth_switch_space", mini.HandlerFunc(func(req mini.Request) {
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
			return
		}
		var in models.SwitchSpaceRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		out, code, err := switchSpace(req.Context(), d, tc.UserID, in.SpaceID)
		if err != nil {
			httpx.Error(req, code, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, out)
	}), mini.Public("POST", "/api/auth/switch-space", "space", "auth.switch_space")))
}

func registerUser(ctx context.Context, d *Deps, in models.RegisterRequest) (*models.LoginResponse, int, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return nil, 400, errMsg("email and password required")
	}
	if len(in.Password) < 4 {
		return nil, 400, errMsg("password too short")
	}
	existing, err := d.Store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, 500, err
	}
	if existing != nil {
		return nil, 409, errMsg("email already registered")
	}
	hash, err := passwd.Hash(in.Password)
	if err != nil {
		return nil, 500, err
	}
	u, err := d.Store.CreateUser(ctx, email, hash, store.ActorUser, false)
	if err != nil {
		return nil, 400, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = email
	}
	sp, err := d.Store.CreateSpace(ctx, "", name)
	if err != nil {
		return nil, 500, err
	}
	if _, err := d.Store.AddMember(ctx, sp.ID, u.ID, store.RoleOwner); err != nil {
		return nil, 500, err
	}
	tok, err := token.Issue([]byte(d.Config.JWTSecret), d.Config.JWTTTL, u.ID, u.Email, u.Actor, u.PlatformAdmin, sp.ID, store.RoleOwner)
	if err != nil {
		return nil, 500, err
	}
	return &models.LoginResponse{
		Token:     tok,
		User:      u.User,
		Space:     sp,
		SpaceRole: store.RoleOwner,
	}, 201, nil
}

func loginAPIKey(ctx context.Context, d *Deps, secret string) (*models.LoginResponse, int, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, 400, errMsg("secret required")
	}
	key, err := d.Store.GetAPIKeyByHash(ctx, passwd.KeyHash(secret))
	if err != nil {
		return nil, 500, err
	}
	if key == nil {
		return nil, 401, errMsg("invalid api key")
	}
	u, err := d.Store.GetUser(ctx, key.UserID)
	if err != nil {
		return nil, 500, err
	}
	if u == nil {
		return nil, 401, errMsg("invalid api key")
	}
	return tokenForUser(ctx, d, u)
}

func tokenForUser(ctx context.Context, d *Deps, u *store.UserRecord) (*models.LoginResponse, int, error) {
	spaces, err := d.Store.ListSpacesForUser(ctx, u.ID)
	if err != nil {
		return nil, 500, err
	}
	var (
		spaceID, spaceRole string
		sp                 *models.Space
	)
	if len(spaces) > 0 {
		chosen := spaces[0]
		for _, item := range spaces {
			if item.ID == store.DefaultSpaceID {
				chosen = item
				break
			}
		}
		spaceID = chosen.ID
		spaceRole = chosen.Role
		copy := chosen.Space
		sp = &copy
	}
	tok, err := token.Issue([]byte(d.Config.JWTSecret), d.Config.JWTTTL, u.ID, u.Email, u.Actor, u.PlatformAdmin, spaceID, spaceRole)
	if err != nil {
		return nil, 500, err
	}
	return &models.LoginResponse{
		Token:     tok,
		User:      u.User,
		Space:     sp,
		SpaceRole: spaceRole,
	}, 200, nil
}

func switchSpace(ctx context.Context, d *Deps, userID, spaceID string) (*models.LoginResponse, int, error) {
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return nil, 400, errMsg("spaceId required")
	}
	u, err := d.Store.GetUser(ctx, userID)
	if err != nil {
		return nil, 500, err
	}
	if u == nil {
		return nil, 401, errMsg("user not found")
	}
	m, err := d.Store.GetMember(ctx, spaceID, userID)
	if err != nil {
		return nil, 500, err
	}
	if m == nil && !u.PlatformAdmin {
		return nil, 403, errMsg("not a member of this space")
	}
	role := store.RoleOwner
	if m != nil {
		role = m.Role
	}
	sp, err := d.Store.GetSpace(ctx, spaceID)
	if err != nil {
		return nil, 500, err
	}
	if sp == nil {
		return nil, 404, errMsg("space not found")
	}
	tok, err := token.Issue([]byte(d.Config.JWTSecret), d.Config.JWTTTL, u.ID, u.Email, u.Actor, u.PlatformAdmin, sp.ID, role)
	if err != nil {
		return nil, 500, err
	}
	return &models.LoginResponse{
		Token:     tok,
		User:      u.User,
		Space:     sp,
		SpaceRole: role,
	}, 200, nil
}
