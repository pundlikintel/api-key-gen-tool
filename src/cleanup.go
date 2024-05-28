package main

import (
	"context"
	"github.com/apikey-gen/aws"
	"github.com/apikey-gen/database"
	"github.com/apikey-gen/model"
	"github.com/sirupsen/logrus"
)

func CleanUp(ctx context.Context) {
	conf, err := model.GetConfig(ctx, "properties.toml")
	if err != nil {
		logrus.Errorf("error in config file %v", err)
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

	ids, err := database.GetSubscriptionIds(ctx, tx, conf.RequiredDetail.EmailDomain)
	if err != nil {
		return
	}

	for _, id := range ids {
		aws.CleanupApiKeys(ctx, id)
	}
	err = database.DeleteSubscriptions(ctx, tx, conf.RequiredDetail.EmailDomain)
	if err != nil {
		return
	}
	tx.Commit()
}
