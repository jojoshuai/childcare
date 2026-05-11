package store

import "childcare-backend/model"

type UserStore interface {
	Create(u *model.User) error
	GetByUsername(username string) (*model.User, error)
	GetByOpenID(openid string) (*model.User, error)
	GetByID(id string) (*model.User, error)
}

type ChildStore interface {
	Create(c *model.Child) error
	GetAll() ([]*model.Child, error)
	GetByID(id string) (*model.Child, error)
	Update(c *model.Child) error
	Delete(id string) error
}

type MeasurementStore interface {
	Create(m *model.Measurement) error
	GetByChildID(childID string, measureType *string) ([]*model.Measurement, error)
	GetByID(id string) (*model.Measurement, error)
	Update(m *model.Measurement) error
	Delete(id string) error
}

type SleepStore interface {
	Create(r *model.SleepRecord) error
	GetByChildID(childID string) ([]*model.SleepRecord, error)
	GetByID(id string) (*model.SleepRecord, error)
	Update(r *model.SleepRecord) error
	Delete(id string) error
}

type DietStore interface {
	Create(r *model.DietRecord) error
	GetByChildID(childID string) ([]*model.DietRecord, error)
	GetByID(id string) (*model.DietRecord, error)
	Update(r *model.DietRecord) error
	Delete(id string) error
}

type SupplementStore interface {
	Create(r *model.SupplementRecord) error
	GetByChildID(childID string) ([]*model.SupplementRecord, error)
	GetByID(id string) (*model.SupplementRecord, error)
	Update(r *model.SupplementRecord) error
	Delete(id string) error
}
