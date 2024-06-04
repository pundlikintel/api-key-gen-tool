package model

import (
	"context"
)
import "github.com/spf13/viper"

type DBConf struct {
	Host     string `json:"host" mapstructure:"host"`
	User     string `json:"user" mapstructure:"user"`
	Password string `json:"password" mapstructure:"password"`
	DBName   string `json:"db_name" mapstructure:"db_name"`
	SSLMode  string `json:"ssl_mode" mapstructure:"ssl_mode"`
	Port     int    `json:"port" mapstructure:"port"`
}

type RequiredDetail struct {
	TenantsCount         int    `json:"tenants_count" mapstructure:"tenants_count"`
	AttKeyPerTenant      int    `json:"att_key_per_tenant" mapstructure:"att_keys_per_tenant"`
	MagtKeyPerTenant     int    `json:"mgmt_key_per_tenant" mapstructure:"mgmt_key_per_tenant"`
	MaintainerEmail      string `json:"maintainer-email" mapstructure:"maintainer_email"`
	AttestationProductId string `json:"attestation_product_id" mapstructure:"attestation_product_id"`
	ManagementProductId  string `json:"management_product_id" mapstructure:"management_product_id"`
	EmailDomain          string `json:"email_domain" mapstructure:"email_domain"`
	ReportTmpl           string `json:"report_tmpl" mapstructure:"report_tmpl"`
	ReportFileName       string `json:"report_file" mapstructure:"report_file"`
	TenantSource         string `json:"tenant_source" mapstructure:"tenant_source"`
}

type AwsConf struct {
	AccessKeyId     string `json:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" mapstructure:"secret_access_key"`
	SessionToken    string `json:"session_token" mapstructure:"session_token"`
	AWSRegion       string `json:"aws_region" mapstructure:"aws_region"`
}

type PoliciesConfig struct {
	Policy                   string `json:"policy" mapstructure:"policy"`
	PolicyName               string `json:"policy_name" mapstructure:"policy_name"`
	PolicyType               string `json:"policy_type" mapstructure:"policy_type"`
	AttestationType          string `json:"attestation_type" mapstructure:"attestation_type"`
	ServiceOfferId           string `json:"service_offer_id" mapstructure:"service_offer_id"`
	Url                      string `json:"url" mapstructure:"ap_url"`
	PlanId                   string `json:"plan_id" mapstructure:"plan_id"`
	ServiceOfferPlanSourceId string `json:"service_offer_plan_source_id" mapstructure:"service_offer_plan_source_id"`
	PolicyCount              int    `json:"policy_count" mapstructure:"policies_per_tennant"`
}

type Config struct {
	DbConf         DBConf         `json:"db_conf" mapstructure:"db_conf"`
	RequiredDetail RequiredDetail `json:"required_detail" mapstructure:"required_detail"`
	AwsConf        AwsConf        `json:"aws_conf" mapstructure:"aws_conf"`
	PoliciesConfig PoliciesConfig `json:"policies_config" mapstructure:"policies_config"`
}

func GetConfig(ctx context.Context, fileName string) (Config, error) {
	viper.SetConfigFile(fileName)
	err := viper.ReadInConfig()

	if err != nil {
		return Config{}, err
	}

	c := new(Config)
	err = viper.Unmarshal(c)
	if err != nil {
		return Config{}, err
	}
	return *c, nil
}
