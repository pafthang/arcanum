package apis

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/passwd"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/space/internal/config"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/internal/token"
	"github.com/pafthang/arcanum/services/space/models"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Config config.Config
}

// Register attaches public HTTP and internal NATS endpoints.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("space/apis.Register: nil Deps")
	}
	registerPublic(svc, d)
	registerInternal(d)
}

func registerPublic(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("auth_login", mini.HandlerFunc(func(req mini.Request) {
		var in models.LoginRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		out, code, err := login(req.Context(), d, in)
		if err != nil {
			httpx.Error(req, code, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, out)
	}),
		mini.WithPublicHTTP("POST", "/api/auth/login"),
		mini.WithPublicSubject("space", "auth.login"),
		mini.WithPublicAuth(mini.AuthNone),
	))

	must(svc.AddEndpoint("spaces_list", mini.HandlerFunc(func(req mini.Request) {
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" && !tc.IsPlatform {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
			return
		}
		items, err := d.Store.ListSpacesForUser(req.Context(), tc.UserID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}),
		mini.Public("GET", "/api/spaces", "space", "spaces.list"),
	))

	must(svc.AddEndpoint("spaces_create", mini.HandlerFunc(func(req mini.Request) {
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
			return
		}
		var in models.CreateSpaceRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		sp, err := d.Store.CreateSpace(req.Context(), "", in.Name)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if _, err := d.Store.AddMember(req.Context(), sp.ID, tc.UserID, store.RoleOwner); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, models.SpaceWithRole{Space: *sp, Role: store.RoleOwner})
	}),
		mini.Public("POST", "/api/spaces", "space", "spaces.create"),
	))

	must(svc.AddEndpoint("spaces_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" && !tc.IsPlatform {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
			return
		}
		if !tc.IsPlatform {
			m, err := d.Store.GetMember(req.Context(), spaceID, tc.UserID)
			if err != nil {
				httpx.Error(req, 500, err.Error(), nil)
				return
			}
			if m == nil {
				httpx.Error(req, 403, "Not a member of this space.", nil)
				return
			}
		}
		sp, err := d.Store.GetSpace(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if sp == nil {
			httpx.Error(req, 404, "Space not found.", nil)
			return
		}
		httpx.JSON(req, 200, sp)
	}),
		mini.Public("GET", "/api/spaces/{spaceId}", "space", "spaces.get"),
	))

	must(svc.AddEndpoint("members_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" && !tc.IsPlatform {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
			return
		}
		if !tc.IsPlatform {
			m, err := d.Store.GetMember(req.Context(), spaceID, tc.UserID)
			if err != nil {
				httpx.Error(req, 500, err.Error(), nil)
				return
			}
			if m == nil {
				httpx.Error(req, 403, "Not a member of this space.", nil)
				return
			}
		}
		items, err := d.Store.ListMembers(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}),
		mini.Public("GET", "/api/spaces/{spaceId}/members", "space", "members.list"),
	))
}

func registerInternal(d *Deps) {
	if d.NC == nil {
		return
	}
	_, _ = d.NC.Subscribe(subjects.InternalSpaceGet, func(msg *nats.Msg) {
		var in struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		sp, err := d.Store.GetSpace(context.Background(), in.ID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if sp == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, sp)
	})
	_, _ = d.NC.Subscribe(subjects.InternalSpaceListForUser, func(msg *nats.Msg) {
		var in struct {
			UserID string `json:"userId"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		items, err := d.Store.ListSpacesForUser(context.Background(), in.UserID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		respondJSON(msg, map[string]any{"items": items})
	})
	_, _ = d.NC.Subscribe(subjects.InternalSpaceUserGet, func(msg *nats.Msg) {
		var in struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		var (
			u   *store.UserRecord
			err error
		)
		if in.ID != "" {
			u, err = d.Store.GetUser(context.Background(), in.ID)
		} else {
			u, err = d.Store.GetUserByEmail(context.Background(), in.Email)
		}
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if u == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, u.User)
	})
}

func login(ctx context.Context, d *Deps, in models.LoginRequest) (*models.LoginResponse, int, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return nil, 400, errMsg("email and password required")
	}
	u, err := d.Store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, 500, err
	}
	if u == nil || !passwd.Verify(in.Password, u.PasswordHash) {
		return nil, 401, errMsg("invalid credentials")
	}
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

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errMsg(s string) error { return simpleError(s) }

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func respondJSON(msg *nats.Msg, v any) {
	b, _ := json.Marshal(v)
	_ = msg.Respond(b)
}

func respondErr(nc *nats.Conn, msg *nats.Msg, code, text string) {
	reply := nats.NewMsg(msg.Reply)
	reply.Header.Set("Nats-Service-Error", text)
	reply.Header.Set("Nats-Service-Error-Code", code)
	reply.Data = []byte(`{}`)
	_ = nc.PublishMsg(reply)
}
