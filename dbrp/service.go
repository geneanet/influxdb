package dbrp

import (
	"context"
	"encoding/json"

	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/kv"
	"github.com/influxdata/influxdb/v2/rand"
)

var (
	bucket = []byte("dbrpv1")
)

type Service struct {
	store          kv.Store
	tokenGenerator influxdb.TokenGenerator
}

func NewService(st kv.Store) influxdb.DBRPMappingServiceV2 {
	return &Service{
		store:          st,
		tokenGenerator: rand.NewTokenGenerator(64),
	}
}

// FindBy returns the dbrp mapping the for cluster, db and rp.
func (s *Service) FindByID(ctx context.Context, id influxdb.ID) (*influxdb.DBRPMapping, error) {
	encodedID, err := id.Encode()
	if err != nil {
		return nil, ErrInvalidDBRPID
	}

	b := []byte{}

	err = s.store.View(ctx, func(tx kv.Tx) error {
		bucket, err := tx.Bucket(bucket)
		if err != nil {
			return ErrInternalServiceError(err)
		}
		b, err = bucket.Get(encodedID)
		if err != nil {
			return ErrDBRPNotFound
		}
		return nil
	})

	dbrp := &influxdb.DBRPMapping{}
	return dbrp, json.Unmarshal(b, dbrp)
}

// FindMany returns a list of dbrp mappings that match filter and the total count of matching dbrp mappings.
func (s *Service) FindMany(ctx context.Context, filter influxdb.DBRPMappingFilter, opts ...influxdb.FindOptions) ([]*influxdb.DBRPMapping, int, error) {
	dbrps := []*influxdb.DBRPMapping{}
	err := s.store.View(ctx, func(tx kv.Tx) error {
		bucket, err := tx.Bucket(bucket)
		if err != nil {
			return ErrInternalServiceError(err)
		}
		cur, err := bucket.Cursor()
		if err != nil {
			return ErrInternalServiceError(err)
		}

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			dbrp := &influxdb.DBRPMapping{}
			json.Unmarshal(v, dbrp)
			if filterFunc(dbrp, filter) {
				dbrps = append(dbrps, dbrp)
			}
		}
		return nil
	})
	if err != nil {
		return nil, len(dbrps), err
	}
	return dbrps, len(dbrps), nil
}

// Create creates a new dbrp mapping, if a different mapping exists an error is returned.
func (s *Service) Create(ctx context.Context, dbrp *influxdb.DBRPMapping) error {
	encodedID, err := dbrp.ID.Encode()
	if err != nil {
		return ErrInternalServiceError(err)
	}
	b, err := json.Marshal(dbrp)
	if err != nil {
		return ErrInternalServiceError(err)
	}

	// if a dbrp with this particular ID already exists an error is returned
	if _, err := s.FindByID(ctx, dbrp.ID); err == nil {
		return ErrDBRPAlreadyExist(err)
	}
	err = s.store.Update(ctx, func(tx kv.Tx) error {
		bucket, err := tx.Bucket(bucket)
		if err != nil {
			return ErrInternalServiceError(err)
		}
		bucket.Put(encodedID, b)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Update a dbrp mapping
func (s *Service) Update(ctx context.Context, dbrp *influxdb.DBRPMapping) error {
	encodedID, err := dbrp.ID.Encode()
	if err != nil {
		return ErrInternalServiceError(err)
	}
	b, err := json.Marshal(dbrp)
	if err != nil {
		return ErrInternalServiceError(err)
	}

	if _, err := s.FindByID(ctx, dbrp.ID); err != nil {
		return ErrDBRPNotFound
	}

	err = s.store.Update(ctx, func(tx kv.Tx) error {
		bucket, err := tx.Bucket(bucket)
		if err != nil {
			return ErrInternalServiceError(err)
		}
		bucket.Put(encodedID, b)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a dbrp mapping.
// Deleting a mapping that does not exists is not an error.
func (s *Service) Delete(ctx context.Context, id influxdb.ID) error {
	encodedID, err := id.Encode()
	if err != nil {
		return ErrInternalServiceError(err)
	}
	err = s.store.Update(ctx, func(tx kv.Tx) error {
		bucket, err := tx.Bucket(bucket)
		if err != nil {
			return ErrInternalServiceError(err)
		}
		return bucket.Delete(encodedID)
	})
	if err != nil {
		return err
	}
	return nil
}

// filterFunc is capable to validate if the dbrp is valid from a given filter.
// it runs true if the filtering data are contained in the dbrp
func filterFunc(dbrp *influxdb.DBRPMapping, filter influxdb.DBRPMappingFilter) bool {
	return (filter.Cluster == nil || (*filter.Cluster) == dbrp.Cluster) &&
		(filter.Database == nil || (*filter.Database) == dbrp.Database) &&
		(filter.RetentionPolicy == nil || (*filter.RetentionPolicy) == dbrp.RetentionPolicy) &&
		(filter.Default == nil || (*filter.Default) == dbrp.Default)
}
