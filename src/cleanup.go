package main

import (
	"context"
	"fmt"
	"github.com/apikey-gen/aws"
	"github.com/apikey-gen/database"
	"github.com/apikey-gen/model"
	"github.com/sirupsen/logrus"
)

func CleanUp(ctx context.Context, count ...int) {
	conf, err := model.GetConfig(ctx, "properties.toml")
	clanupCont := -1
	if err != nil {
		logrus.Errorf("error in config file %v", err)
		return
	}

	if len(count) > 0 {
		clanupCont = count[0]
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

	var ers []error
	for _, id := range ids {
		err = aws.CleanupApiKeys(ctx, id)
		if err != nil {
			ers = append(ers, err)
		}
	}
	if len(ids) == len(ers) && len(ers) > 0 {
		var resp string
		fmt.Println("All delete api key calls failed, Do you want to terminate? (y/n)")
		for {
			_, err := fmt.Scanln(&resp)
			if err != nil {
				return
			}
			if resp == "y" {
				return
			} else if resp == "n" {
				break
			}
		}
	}

	err = database.DeleteSubscriptions(ctx, tx, conf.RequiredDetail.EmailDomain, clanupCont)
	if err != nil {
		return
	}

	err = database.DeletePolicies(ctx, tx, conf.RequiredDetail.EmailDomain, clanupCont)
	if err != nil {
		return
	}

	err = database.DeleteService(ctx, tx, conf.RequiredDetail.EmailDomain, clanupCont)
	if err != nil {
		return
	}

	err = database.DeleteTenants(ctx, tx, conf.RequiredDetail.EmailDomain, clanupCont)
	if err != nil {
		return
	}

	var resp string
	fmt.Println("Do you want to COMMIT transaction. All deleted data will can not be restored once deleted? (yes/no)")
	for {
		_, err := fmt.Scanln(&resp)
		if err != nil {
			return
		}
		if resp == "yes" {
			break
		} else if resp == "no" {
			return
		} else {
			logrus.Info("Type yes/no")
		}
	}

	fmt.Println("Reconfirm. All deleted data will can not be restored once deleted? (yes/no)")
	for {
		_, err := fmt.Scanln(&resp)
		if err != nil {
			return
		}
		if resp == "yes" {
			break
		} else if resp == "no" {
			return
		} else {
			logrus.Info("Type yes/no")
		}
	}

	tx.Commit()
}
