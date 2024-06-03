package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apikey-gen/aws"
	"github.com/apikey-gen/database"
	"github.com/apikey-gen/model"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
)

func Create(ctx context.Context) {
	conf, err := model.GetConfig(ctx, "properties.toml")
	if err != nil {
		logrus.Errorf("error in config file %v", err)
		return
	}

	attestationProductId, err := uuid.Parse(conf.RequiredDetail.AttestationProductId)
	if err != nil {
		logrus.Errorf("error in parsing attestation product id %s, %v", conf.RequiredDetail.AttestationProductId, err)
		return
	}
	managementProductId, err := uuid.Parse(conf.RequiredDetail.ManagementProductId)
	if err != nil {
		logrus.Errorf("error in parsing management product id %s, %v", conf.RequiredDetail.AttestationProductId, err)
		return
	}
	/************** AWS init *******************/
	cli := aws.InitAwsClient(conf.AwsConf.AccessKeyId, conf.AwsConf.SecretAccessKey, conf.AwsConf.SessionToken, conf.AwsConf.AWSRegion)
	if cli == nil {
		logrus.Errorf("error in creating aws client")
		return
	}

	/************** Database ******************/
	connection, err := database.GetConnection(ctx, model.DBConf{
		Host:     conf.DbConf.Host,
		Port:     conf.DbConf.Port,
		User:     conf.DbConf.User,
		Password: conf.DbConf.Password,
		DBName:   conf.DbConf.DBName,
		SSLMode:  conf.DbConf.SSLMode,
	})
	if err != nil {
		return
	}

	tx := connection.Begin()
	defer tx.Rollback()

	tenantsCount := conf.RequiredDetail.TenantsCount
	keysPerTenant := conf.RequiredDetail.AttKeyPerTenant
	mgmtkeysPerTenant := conf.RequiredDetail.MagtKeyPerTenant
	policiesCount := conf.PoliciesConfig.PolicyCount
	tenants, err := CreateTenant(ctx, tenantsCount, tx, conf.RequiredDetail.EmailDomain,
		uuid.MustParse(conf.PoliciesConfig.ServiceOfferId), uuid.MustParse(conf.PoliciesConfig.PlanId), uuid.MustParse(conf.PoliciesConfig.ServiceOfferPlanSourceId))
	if err != nil {
		return
	}
	logrus.Infof("%d Tenants created successfully", tenantsCount)
	/************** Create API keys ******************/
	attestationProductExtId, err := database.GetProductExtId(ctx, tx, attestationProductId)
	managementProductExtId, err := database.GetProductExtId(ctx, tx, managementProductId)
	if err != nil {
		logrus.Errorf("error in getting product external id %v", err)
		return
	}

	tx.Commit() //has to commit otherwise create policy will fail
	apiKeysInfos := make([]model.ApiKeyModel, 0)
	wg := sync.WaitGroup{}
	wg.Add(len(tenants))
	for _, tenant := range tenants {
		go func(wgPtr *sync.WaitGroup) {
			defer wg.Done()
			logrus.Infof("Creating api keys for tenant %s", tenant.ID)
			apiKeyInfo, err := CreateAPIKey(ctx, keysPerTenant, mgmtkeysPerTenant, policiesCount, connection, tenant.ID, attestationProductId, managementProductId, tenant.ServiceId,
				attestationProductExtId, managementProductExtId, conf.RequiredDetail.MaintainerEmail)
			if err != nil {
				logrus.Errorf("error in create api key %v", err)
				return
			}
			apiKeysInfos = append(apiKeysInfos, apiKeyInfo...)
		}(&wg)
	}
	wg.Wait()
	ExportToFile(ctx, conf.RequiredDetail.ReportFileName, conf.RequiredDetail.ReportTmpl, apiKeysInfos)
}

func ExportToFile(ctx context.Context, fileName string, template string, apiKeys []model.ApiKeyModel) {
	fileName = fmt.Sprintf(fileName, time.Now().UnixNano())
	templants := []string{}

	if len(apiKeys) == 0 {
		return
	}

	apiKey := apiKeys[0]
	apiKeyMap := map[string]string{}
	byt, _ := json.Marshal(apiKey)
	_ = json.Unmarshal(byt, &apiKeyMap)
	tmplStr := template
	for key, _ := range apiKeyMap {
		tmplStr = strings.ReplaceAll(tmplStr, fmt.Sprintf("{{%s}}", key), key)
	}
	templants = append(templants, tmplStr)

	for _, apiKey := range apiKeys {
		apiKeyMap := map[string]string{}
		byt, _ := json.Marshal(apiKey)
		_ = json.Unmarshal(byt, &apiKeyMap)
		tmplStr := template
		for key, value := range apiKeyMap {
			tmplStr = strings.ReplaceAll(tmplStr, fmt.Sprintf("{{%s}}", key), value)
		}
		templants = append(templants, tmplStr)
	}

	tmp := strings.Join(templants, "\n")
	err := os.WriteFile(fileName, []byte(tmp), fs.ModeAppend)
	if err != nil {
		return
	}
}
