package gcp

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func GetSecret(secretName string) (interface{}, error) {
	log.Debugf("Downloading secret %s", secretName)
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", err
	}

	return utils.ParseSecretData(string(result.Payload.Data), secretName)
}

func CreateSecret(projectID, secretName string, secretData interface{}) (err error) {

	secretDataStr, err := utils.MarshallSecretData(secretData)
	if err != nil {
		return
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return
	}

	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectID),
		SecretId: secretName,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	secret, err := client.CreateSecret(ctx, createSecretReq)
	if err != nil {
		return
	}

	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretDataStr),
		},
	}

	_, err = client.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return
	}

	return
}

func UpdateSecret(secretName string, secretData interface{}) (err error) {

	secretDataStr, err := utils.MarshallSecretData(secretData)
	if err != nil {
		return
	}

	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return
	}

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return
	}

	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: result.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretDataStr),
		},
	}

	_, err = client.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return
	}

	return
}
