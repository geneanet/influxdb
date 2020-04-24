package token

import (
	"context"

	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/authorizer"
	icontext "github.com/influxdata/influxdb/v2/context"
)

type AuthedTokenService struct {
	s  influxdb.AuthorizationService
	ts influxdb.TenantService
}

func NewAuthedTokenService(s influxdb.AuthorizationService, ts influxdb.TenantService) *AuthedTokenService {
	return &AuthedTokenService{
		s:  s,
		ts: ts,
	}
}

func (s *AuthedTokenService) CreateAuthorization(ctx context.Context, a *influxdb.Authorization) error {
	if a.UserID == 0 {
		auth, err := icontext.GetAuthorizer(ctx)
		if err != nil {
			return err
		}

		user, err := s.ts.FindUserByID(ctx, auth.GetUserID())
		if err != nil {
			// if we could not get the user from the Authorization object or the Context,
			// then we cannot authorize the user
			return err
		}
		a.UserID = user.ID
	}

	if _, _, err := authorizer.AuthorizeCreate(ctx, influxdb.AuthorizationsResourceType, a.OrgID); err != nil {
		return err
	}
	if _, _, err := authorizer.AuthorizeWriteResource(ctx, influxdb.UsersResourceType, a.UserID); err != nil {
		return err
	}
	if err := authorizer.VerifyPermissions(ctx, a.Permissions); err != nil {
		return err
	}

	return s.s.CreateAuthorization(ctx, a)
}

func (s *AuthedTokenService) FindAuthorizationByToken(ctx context.Context, t string) (*influxdb.Authorization, error) {
	return s.s.FindAuthorizationByToken(ctx, t)
}

func (s *AuthedTokenService) FindAuthorizations(ctx context.Context, filter influxdb.AuthorizationFilter, opt ...influxdb.FindOptions) ([]*influxdb.Authorization, int, error) {
	return s.s.FindAuthorizations(ctx, filter, opt...)
}

func (s *AuthedTokenService) UpdateAuthorization(ctx context.Context, id influxdb.ID, upd *influxdb.AuthorizationUpdate) (*influxdb.Authorization, error) {
	return s.s.UpdateAuthorization(ctx, id, upd)
}

func (s *AuthedTokenService) DeleteAuthorization(ctx context.Context, id influxdb.ID) error {
	return s.s.DeleteAuthorization(ctx, id)
}
