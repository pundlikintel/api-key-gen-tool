package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/apikey-gen/aws"
	"github.com/apikey-gen/database"
	"github.com/apikey-gen/model"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	ServiceId uuid.UUID `json:"service_id"`
}

func CreateTenant(ctx context.Context, tenantsCount int, tx *gorm.DB, emailDomain string, serviceOfferId, planId, serviceOfferPlanSourceId uuid.UUID) ([]Tenant, error) {
	var tenantsId []Tenant
	for i := 0; i < tenantsCount; i++ {
		tenantId := uuid.New()
		serviceId := uuid.New()
		err := database.MakeTenantEntry(ctx, tx, &model.Tenant{
			ID:      tenantId,
			Name:    "TestName",
			Company: "TestCompany",
			Email:   fmt.Sprintf("%s@%s", uuid.NewString(), emailDomain),
			Address: "address",
		})
		if err != nil {
			return nil, err
		}

		err = database.MakeServiceEntry(ctx, tx, &model.Service{
			ID:                       serviceId,
			TenantId:                 tenantId,
			ServiceOfferId:           serviceOfferId,
			Name:                     "TEE_Attestation",
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
			CreatedBy:                uuid.UUID{},
			UpdatedBy:                uuid.UUID{},
			CreatorType:              "User",
			UpdaterType:              "User",
			ExternalId:               uuid.UUID{},
			PlanId:                   planId,
			Active:                   true,
			Status:                   "Active",
			ServiceOfferPlanSourceId: serviceOfferPlanSourceId,
		})
		if err != nil {
			return nil, err
		}
		tenantsId = append(tenantsId, Tenant{
			ID:        tenantId,
			ServiceId: serviceId,
		})
	}
	return tenantsId, nil
}

func CreateAPIKey(ctx context.Context, attestationKeysPerTenant, managementKeysPerTenant int, tx *gorm.DB, tenantId, attestationProductId, managementProductId,
	serviceId uuid.UUID, attProductExtId, mgmtProductExtId, email string) ([]model.ApiKeyModel, error) {
	var apiKeyModels []model.ApiKeyModel

	conf, err := model.GetConfig(context.Background(), "properties.toml")
	if err != nil {
		panic(err)
	}

	for i := 0; i < managementKeysPerTenant; i++ {
		apiKeyInfo, err := createApiKey(ctx, tx, managementProductId, serviceId, tenantId, mgmtProductExtId, email, "")
		if err != nil {
			return nil, err
		}
		apiKeyInfo.KeyType = "management"
		apiKeyModels = append(apiKeyModels, apiKeyInfo)
	}

	logrus.Infof("Sleeping for 2 minutes to set management api keys")
	time.Sleep(2 * time.Minute)

	for i := 0; i < attestationKeysPerTenant; i++ {
		//Create policy
		policyId, err := CreatePolicy(ctx, conf.PoliciesConfig.Url, apiKeyModels[0].FullKey)
		if err != nil {
			return nil, err
		}

		apiKeyInfo, err := createApiKey(ctx, tx, attestationProductId, serviceId, tenantId, attProductExtId, email, policyId)
		if err != nil {
			return nil, err
		}
		apiKeyInfo.KeyType = "attestation"
		apiKeyInfo.PolicyId = uuid.MustParse(policyId)
		apiKeyModels = append(apiKeyModels, apiKeyInfo)
	}

	return apiKeyModels, nil
}

func createApiKey(ctx context.Context, tx *gorm.DB, productId, serviceId, tenantId uuid.UUID, prdExtId, email, policyId string) (model.ApiKeyModel, error) {
	apiKey := uuid.New()
	variableKey := uuid.NewString()
	name := fmt.Sprintf("ApiKey_Perf_%s", uuid.NewString())
	keyExtId, keyValue, err := aws.CreateApiKey(ctx, name, apiKey.String(), prdExtId, email)
	if err != nil {
		return model.ApiKeyModel{}, err
	}
	err = database.MakeSubscriptionEntry(ctx, tx, &model.Subscription{
		ID:          apiKey,
		ServiceId:   serviceId,
		ProductId:   productId,
		TenantId:    tenantId,
		Status:      "Active",
		Name:        name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   uuid.UUID{},
		UpdatedBy:   uuid.UUID{},
		CreatorType: "User",
		UpdaterType: "User",
		ExternalId:  keyExtId,
		Version:     "v1",
		VariableKey: variableKey,
		DeletedAt:   gorm.DeletedAt{},
	})
	if err != nil {
		return model.ApiKeyModel{}, err
	}

	if policyId != "" {
		err = database.MakeSubscriptionPolicyEntry(ctx, tx, &model.SubscriptionPolicy{
			TenantId:       tenantId,
			SubscriptionId: apiKey,
			PolicyId:       uuid.MustParse(policyId),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      uuid.UUID{},
			UpdatedBy:      uuid.UUID{},
			Deleted:        false,
		})
		if err != nil {
			return model.ApiKeyModel{}, err
		}
	}

	apiKeyInfo := model.ApiKeyModel{
		TenantId:    tenantId,
		ID:          apiKey,
		VariableKey: variableKey,
		ApiKey:      keyValue,
		Version:     "v1",
	}

	apiKeyInfo.FullKey = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s:%s", apiKeyInfo.Version, apiKeyInfo.VariableKey, apiKeyInfo.ApiKey)))
	return apiKeyInfo, nil
}

func CreatePolicy(ctx context.Context, url, managementKey string) (policyId string, err error) {
	conf, err := model.GetConfig(ctx, "properties.toml")
	if err != nil {
		logrus.Errorf("error in config file %v", err)
		return
	}
	policy := model.PolicyModel{
		PolicyName:      conf.PoliciesConfig.PolicyName,
		PolicyType:      conf.PoliciesConfig.PolicyType,
		AttestationType: conf.PoliciesConfig.AttestationType,
		ServiceOfferId:  conf.PoliciesConfig.ServiceOfferId,
	}
	policy.Policy = strings.ReplaceAll(conf.PoliciesConfig.Policy, "{sgx_mrenclave}", RandStringRunes(64))
	policy.Policy = strings.ReplaceAll(policy.Policy, "{sgx_mrsigner}", RandStringRunes(64))
	postBody, err := json.Marshal(policy)
	responseBody := bytes.NewBuffer(postBody)
	req, err := http.NewRequest("POST", url, responseBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", managementKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	op, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Output %s", op)
	}

	mp := map[string]interface{}{}

	if err != nil || resp.StatusCode != 201 {
		logrus.Errorf("Error in create poicy %v, status code %s, [%s]", err, resp.Status, op)
		return "", err
	}

	err = json.Unmarshal(op, &mp)
	if err != nil {
		return "", err
	}

	return mp["policy_id"].(string), nil
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
