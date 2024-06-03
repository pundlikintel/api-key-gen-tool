package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/apikey-gen/model"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

func GetConnection(ctx context.Context, cfg model.DBConf) (db *gorm.DB, err error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	dil := postgres.Open(connStr)
	db, err = gorm.Open(dil, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	return db, err
}

func MakeTenantEntry(ctx context.Context, tx *gorm.DB, tenant *model.Tenant) error {
	t := tx.Create(tenant)
	return t.Error
}

func MakeServiceEntry(ctx context.Context, tx *gorm.DB, service *model.Service) error {
	t := tx.Create(service)
	return t.Error
}

func MakeSubscriptionEntry(ctx context.Context, tx *gorm.DB, subscription *model.Subscription) error {
	t := tx.Create(subscription)
	return t.Error
}

func MakeSubscriptionPolicyEntry(ctx context.Context, tx *gorm.DB, subscriptionPolicy *model.SubscriptionPolicy) error {
	t := tx.Create(subscriptionPolicy)
	return t.Error
}

func GetProductExtId(ctx context.Context, tx *gorm.DB, productId uuid.UUID) (string, error) {
	var externalId string
	if d := tx.Table("product").Select("external_id").Where("id = ?", productId); d.Error != nil {
		logrus.Errorf("Error in gettting external id %s, %v", productId.String(), d.Error)
		return "", d.Error
	} else {
		d.Scan(&externalId)
	}
	return externalId, nil
}

func GetSubscriptionIds(ctx context.Context, tx *gorm.DB, tenantEmailDomain string) ([]string, error) {
	if strings.TrimSpace(tenantEmailDomain) == "" {
		return nil, errors.New("tenantEmailDomain can not be empty")
	}
	extIds := make([]string, 0)
	query := fmt.Sprintf("select external_id from subscription where tenant_id in (select id from tenant where email like '%s@%s')", "%", tenantEmailDomain)
	res := tx.Raw(query)
	if res.Error != nil {
		return nil, res.Error
	}
	if rows, err := res.Rows(); err != nil {
		return nil, err
	} else {
		for rows.Next() {
			var extId string
			err := rows.Scan(&extId)
			if err != nil {
				return nil, err
			}
			extIds = append(extIds, extId)
		}
	}
	return extIds, nil
}

func DeleteSubscriptions(ctx context.Context, tx *gorm.DB, tenantEmailDomain string, count int) error {
	if strings.TrimSpace(tenantEmailDomain) == "" {
		return errors.New("tenantEmailDomain can not be empty")
	}

	query := fmt.Sprintf("delete from subscription where tenant_id in (select id from tenant where email like '%s@%s')", "%", tenantEmailDomain)
	if count > 0 {
		query = fmt.Sprintf("delete from subscription where tenant_id in (select id from tenant where email like '%s@%s' order by id limit %d)", "%", tenantEmailDomain, count)
	}

	res := tx.Exec(query)
	if res.Error != nil {
		logrus.Errorf("Error in deleting subscriptions %v", res.Error)
		return res.Error
	} else {
		logrus.Infof("%d subscriptions deleted", res.RowsAffected)
	}
	return nil
}

func DeletePolicies(ctx context.Context, tx *gorm.DB, tenantEmailDomain string, count int) error {
	if strings.TrimSpace(tenantEmailDomain) == "" {
		return errors.New("tenantEmailDomain can not be empty")
	}
	query := fmt.Sprintf("delete from policy where cast( tenant_id as uuid) in (select id from tenant where email like '%s@%s')", "%", tenantEmailDomain)
	if count > 0 {
		query = fmt.Sprintf("delete from policy where cast( tenant_id as uuid) in (select id from tenant where email like '%s@%s' order by id limit %d)", "%", tenantEmailDomain, count)
	}
	res := tx.Exec(query)
	if res.Error != nil {
		logrus.Errorf("Error in deleting policies %v", res.Error)
		return res.Error
	} else {
		logrus.Infof("%d Policies deleted", res.RowsAffected)
	}
	return nil
}

func DeleteService(ctx context.Context, tx *gorm.DB, tenantEmailDomain string, count int) error {
	if strings.TrimSpace(tenantEmailDomain) == "" {
		return errors.New("tenantEmailDomain can not be empty")
	}
	query := fmt.Sprintf("delete from service where tenant_id in (select id from tenant where email like '%s@%s')", "%", tenantEmailDomain)
	if count > 0 {
		query = fmt.Sprintf("delete from service where tenant_id in (select id from tenant where email like '%s@%s' order by id limit %d)", "%", tenantEmailDomain, count)
	}
	res := tx.Exec(query)
	if res.Error != nil {
		logrus.Errorf("Error in deleting Service %v", res.Error)
		return res.Error
	} else {
		logrus.Infof("%d Services deleted", res.RowsAffected)
	}
	return nil
}

func DeleteTenants(ctx context.Context, tx *gorm.DB, tenantEmailDomain string, count int) error {
	if strings.TrimSpace(tenantEmailDomain) == "" {
		return errors.New("tenantEmailDomain can not be empty")
	}
	query := fmt.Sprintf("delete from tenant where email like '%s@%s'", "%", tenantEmailDomain)
	if count > 0 {
		query = fmt.Sprintf("delete from tenant where id in (select id from tenant where email like '%s@%s' order by id limit %d)", "%", tenantEmailDomain, count)
	}
	res := tx.Exec(query)
	if res.Error != nil {
		logrus.Errorf("Error in deleting tenants %v", res.Error)
		return res.Error
	} else {
		logrus.Infof("%d tenants deleted", res.RowsAffected)
	}
	return nil
}
