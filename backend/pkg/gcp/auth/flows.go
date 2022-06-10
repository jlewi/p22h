// Code originally based on https://github.com/kubeflow/internal-acls/blob/master/google_groups/cmd/main.go#L129
package auth

import (
	"fmt"
	"strings"
)

func getWebFlowSecretManager() *gcp.CachedCredentialHelper {
	webFlow, err := gcp.NewWebFlowHelper(opts.CredentialsFile, scopes)

	if err != nil {
		log.Error(err, "Failed to create a WebFlowHelper credential helper")
		return nil
	}

	pieces := strings.Split(opts.Secret, "/")

	if len(pieces) != 2 {
		log.Error(fmt.Errorf("Secret %v not in form {project}/{secret}", opts.Secret), "Incorrectly specified secret", "secret", opts.Secret)
		return nil
	}

	cache, err := gcp.NewSecretCache(pieces[0], pieces[1], "latest")

	if err != nil {
		log.Error(err, "Could not create cache for secret manager")
		return nil
	}

	cache.Log = log

	h := &gcp.CachedCredentialHelper{
		CredentialHelper: webFlow,
		TokenCache:       cache,
		Log:              log,
	}

	return h
}
