package store

import "childcare-backend/model"

// UserStore defines the interface for user persistence operations.
type UserStore interface {
	Create(u *model.User) error
	GetByUsername(username string) (*model.User, error)
	GetByOpenID(openid string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	UpdateFamily(userID, familyID, role string) error
}

// FamilyStore defines the interface for family persistence operations.
type FamilyStore interface {
	Create(f *model.Family) error
	GetByID(id string) (*model.Family, error)
	GetMembers(familyID string) ([]*model.User, error)
}

// ChildStore defines the interface for child persistence operations.
type ChildStore interface {
	Create(c *model.Child) error
	GetByFamilyID(familyID string) ([]*model.Child, error)
	GetByID(id string) (*model.Child, error)
	Update(c *model.Child) error
	Delete(id string) error
}

// MeasurementStore defines the interface for measurement persistence operations.
type MeasurementStore interface {
	Create(m *model.Measurement) error
	GetByChildID(childID string, measureType *string) ([]*model.Measurement, error)
	GetByID(id string) (*model.Measurement, error)
	Update(m *model.Measurement) error
	Delete(id string) error
}

// InviteStore defines the interface for invite code persistence operations.
type InviteStore interface {
	Create(ic *model.InviteCode) error
	GetByCode(code string) (*model.InviteCode, error)
	MarkUsed(id string) error
}
