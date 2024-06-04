package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/sirupsen/logrus"
)

var awsClient *apigateway.Client

// credentialProvider /* Credential provider */
type credentialProvider struct {
	accessKeyId, secretAccessKey, sessionToken string
}

func (cp *credentialProvider) setCredentials(accessKeyId, secretAccessKey, sessionToken string) {
	cp.secretAccessKey = secretAccessKey
	cp.accessKeyId = accessKeyId
	cp.sessionToken = sessionToken
}

func (cp *credentialProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     cp.accessKeyId,
		SecretAccessKey: cp.secretAccessKey,
		SessionToken:    cp.sessionToken,
	}, nil
}

func InitAwsClient(AccessKeyId, SecretAccessKey, sessionToken, AWSRegion string) *apigateway.Client {
	credProvider := &credentialProvider{}
	//TBD: Credential provider need to change based on discussion
	credProvider.setCredentials(AccessKeyId, SecretAccessKey, sessionToken)
	conf := aws.Config{Credentials: credProvider, Region: AWSRegion, RetryMaxAttempts: 10, RetryMode: aws.RetryModeStandard}
	//AWS client for APIGateway related functions
	awsClient = apigateway.NewFromConfig(conf)
	return awsClient
}

func CreateApiKey(ctx context.Context, name, subscriptionId, prdExtId, email string) (string, string, error) {
	tags := map[string]string{"operation": "perf_testing", "maintainer": email}
	apiKeyOut, err := awsClient.CreateApiKey(ctx, &apigateway.CreateApiKeyInput{
		Description: aws.String(name),
		Enabled:     true,
		Name:        aws.String(subscriptionId),
		Tags:        tags,
	})
	if err != nil {
		return "", "", err
	}

	_, err = awsClient.CreateUsagePlanKey(ctx, &apigateway.CreateUsagePlanKeyInput{
		KeyId:       apiKeyOut.Id,
		KeyType:     aws.String("API_KEY"),
		UsagePlanId: aws.String(prdExtId),
	})

	if err != nil {
		return "", "", err
	}

	return *apiKeyOut.Id, *apiKeyOut.Value, nil
}

func CleanupApiKeys(ctx context.Context, id string) error {
	_, err := awsClient.DeleteApiKey(ctx, &apigateway.DeleteApiKeyInput{ApiKey: aws.String(id)})
	if err != nil {
		logrus.Errorf("Error in delete key from aws %s", id)
	} else {
		logrus.Infof("Deleted api key %s", id)
	}
	return err
}
