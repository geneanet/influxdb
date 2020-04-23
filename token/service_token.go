package token

import (
	"context"

	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/kv"
)

var _ influxdb.AuthorizationService = (*Service)(nil)

func (s *Service) CreateAuthorization(ctx context.Context, a *influxdb.Authorization) error {
	// // todo (al): we also need to check if the user has write permissions
	// can the auth middleware layer do this?
	// http layer can handle adding user id...?
	// if u, err := s.findUserByID(ctx, tx, a.UserID); err != nil {
	// 	return influxdb.ErrUnableToCreateToken
	// }

	if err := a.Valid(); err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	// todo (al): thorough notes that the checking is done in http, but eventually
	// should be moved here if and when those lookups are removed from the HTTP layer\
	// up for discussion though.
	// make sure that org and user are real

	if a.Token == "" {
		token, err := s.tokenGenerator.Token()
		if err != nil {
			return &influxdb.Error{
				Err: err,
			}
		}
		a.Token = token
	}

	return s.store.Update(ctx, func(tx kv.Tx) error {
		return s.store.CreateAuthorization(ctx, tx, a)
	})
}

func (s *Service) FindAuthorizationByID(ctx context.Context, id influxdb.ID) (*influxdb.Authorization, error) {
	var a *influxdb.Authorization
	err := s.store.View(ctx, func(tx kv.Tx) error {
		auth, err := s.store.GetAuthorizationByID(ctx, tx, id)
		if err != nil {
			return nil
		}

		a = auth
		return nil
	})

	if err != nil {
		return nil, err
	}

	return a, nil
}

// FindAuthorizationByToken returns a authorization by token for a particular authorization.
func (s *Service) FindAuthorizationByToken(ctx context.Context, n string) (*influxdb.Authorization, error) {
	var a *influxdb.Authorization
	err := s.store.View(ctx, func(tx kv.Tx) error {
		auth, err := s.store.GetAuthorizationByToken(ctx, tx, n)
		if err != nil {
			return err
		}

		a = auth

		return nil
	})

	if err != nil {
		return nil, err
	}

	return a, nil
}

// FindAuthorizations retrives all authorizations that match an arbitrary authorization filter.
// Filters using ID, or Token should be efficient.
// Other filters will do a linear scan across all authorizations searching for a match.
func (s *Service) FindAuthorizations(ctx context.Context, filter influxdb.AuthorizationFilter, opt ...influxdb.FindOptions) ([]*influxdb.Authorization, int, error) {
	if filter.ID != nil {
		var auth *influxdb.Authorization
		err := s.store.View(ctx, func(tx kv.Tx) error {
			a, e := s.store.GetAuthorizationByID(ctx, tx, *filter.ID)
			if e != nil {
				return e
			}
			auth = a
			return nil
		})
		if err != nil {
			return nil, 0, &influxdb.Error{
				Err: err,
			}
		}

		return []*influxdb.Authorization{auth}, 1, nil
	}

	if filter.Token != nil {
		var auth *influxdb.Authorization
		err := s.store.View(ctx, func(tx kv.Tx) error {
			a, e := s.store.GetAuthorizationByToken(ctx, tx, *filter.Token)
			if e != nil {
				return e
			}
			auth = a
			return nil
		})
		if err != nil {
			return nil, 0, &influxdb.Error{
				Err: err,
			}
		}

		return []*influxdb.Authorization{auth}, 1, nil
	}

	as := []*influxdb.Authorization{}
	err := s.store.View(ctx, func(tx kv.Tx) error {
		auths, err := s.store.ListAuthorizations(ctx, tx, filter)
		if err != nil {
			return err
		}
		as = auths
		return nil
	})

	if err != nil {
		return nil, 0, &influxdb.Error{
			Err: err,
		}
	}

	return as, len(as), nil
}

// UpdateAuthorization updates the status and description if available.
func (s *Service) UpdateAuthorization(ctx context.Context, id influxdb.ID, upd *influxdb.AuthorizationUpdate) (*influxdb.Authorization, error) {
	var auth *influxdb.Authorization
	err := s.store.Update(ctx, func(tx kv.Tx) error {
		a, e := s.store.UpdateAuthorization(ctx, tx, id, upd)
		if e != nil {
			return e
		}
		auth = a
		return nil
	})
	return auth, err
}

func (s *Service) DeleteAuthorization(ctx context.Context, id influxdb.ID) error {
	return s.store.Update(ctx, func(tx kv.Tx) (err error) {
		return s.store.DeleteAuthorization(ctx, tx, id)
	})
}
