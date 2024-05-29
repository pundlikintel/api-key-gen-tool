package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Tenant struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Company    string    `json:"company"`
	Address    string    `json:"address"`
	ParentId   uuid.UUID `json:"-"`
	Email      string    `json:"email"`
	ExternalId uuid.UUID `json:"-"`
	SourceId   uuid.UUID `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	UpdatedBy  uuid.UUID `json:"-"`
	CreatedAt  time.Time `json:"-"`
}

type Service struct {
	ID                       uuid.UUID `gorm:"primary_key;type:uuid"`
	TenantId                 uuid.UUID `gorm:"uniqueIndex:idx_tenant_service-offer_plan_name;type:uuid"`
	ServiceOfferId           uuid.UUID `gorm:"type:uuid;not null"`
	Name                     string    `gorm:"uniqueIndex:idx_tenant_service-offer_plan_name;type:string"`
	CreatedAt                time.Time `gorm:"not null"`
	UpdatedAt                time.Time `gorm:"not null"`
	CreatedBy                uuid.UUID `gorm:"type:uuid"`
	UpdatedBy                uuid.UUID `gorm:"type:uuid"`
	CreatorType              string    `gorm:"type:string;not null"`
	UpdaterType              string    `gorm:"type:string"`
	ExternalId               uuid.UUID `gorm:"type:uuid;not null"`
	PlanId                   uuid.UUID `gorm:"type:uuid;not null"`
	Active                   bool      `gorm:"type:bool"`
	Status                   string    `gorm:"type:string;not null"`
	ServiceOfferPlanSourceId uuid.UUID `gorm:"uniqueIndex:idx_tenant_service-offer_plan_name;type:uuid;not null"`
}

type Subscription struct {
	ID          uuid.UUID      `gorm:"primary_key;type:uuid"`
	ServiceId   uuid.UUID      `gorm:"uniqueIndex:idx_product_service_name;type:uuid"`
	ProductId   uuid.UUID      `gorm:"uniqueIndex:idx_product_service_name;type:uuid"`
	TenantId    uuid.UUID      `gorm:"type:uuid;not null"`
	Status      string         `gorm:"type:string"`
	Name        string         `gorm:"uniqueIndex:idx_product_service_name;type:string"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid"`
	UpdatedBy   uuid.UUID      `gorm:"type:uuid"`
	CreatorType string         `gorm:"type:string;not null"`
	UpdaterType string         `gorm:"type:string"`
	ExternalId  string         `gorm:"type:string" json:"-"`
	Version     string         `gorm:"type:string" json:"-"`
	VariableKey string         `gorm:"type:string"`
	DeletedAt   gorm.DeletedAt `gorm:"type:time"`
}

type SubscriptionPolicy struct {
	TenantId       uuid.UUID `gorm:"type:uuid"`
	SubscriptionId uuid.UUID `gorm:"type:uuid"`
	PolicyId       uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
	CreatedBy      uuid.UUID `gorm:"type:uuid"`
	UpdatedBy      uuid.UUID `gorm:"type:uuid"`
	Deleted        bool      `gorm:"uniqueIndex:idx_unique_sub_policy, where:deleted = 'f';not null"`
}

type ApiKeyModel struct {
	TenantId    uuid.UUID `json:"tenant_id"`
	ID          uuid.UUID `json:"id"`
	VariableKey string    `json:"variable_key"`
	ApiKey      string    `json:"api_key"`
	Version     string    `json:"version"`
	FullKey     string    `json:"full_key"`
	KeyType     string    `json:"key_type"`
	PolicyId    string    `json:"policy_id"`
}

type PolicyModel struct {
	Policy          string `json:"policy"`
	PolicyName      string `json:"policy_name"`
	PolicyType      string `json:"policy_type"`
	AttestationType string `json:"attestation_type"`
	ServiceOfferId  string `json:"service_offer_id"`
}

type ServiceModel struct {
	ServiceOfferId string `json:"service_offer_id"`
	PlanId         string `json:"plan_id"`
	Source         string `json:"source"`
	Name           string `json:"name"`
}
