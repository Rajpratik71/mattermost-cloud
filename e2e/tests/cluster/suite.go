// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

//+build e2e

package cluster

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-cloud/clusterdictionary"

	"github.com/mattermost/mattermost-cloud/e2e/pkg"
	"github.com/mattermost/mattermost-cloud/e2e/workflow"
	"github.com/mattermost/mattermost-cloud/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

// TODO: we can further parametrize the test according to our needs

// TestConfig is test configuration coming from env vars.
type TestConfig struct {
	CloudURL                  string `envconfig:"default=http://localhost:8075"`
	InstallationDBType        string `envconfig:"default=mysql-operator"`
	InstallationFileStoreType string `envconfig:"default=minio-operator"`
	Environment               string `envconfig:"default=dev"`
	WebhookAddress            string `envconfig:"default=http://localhost:11111"`
	Cleanup                   bool   `envconfig:"default=true"`
}

// Test holds all data required for a db migration test.
type Test struct {
	Logger            logrus.FieldLogger
	Workflow          *workflow.Workflow
	ClusterSuite      *workflow.ClusterSuite
	InstallationSuite *workflow.InstallationSuite
	WebhookCleanup    func() error
	Cleanup           bool
}

// SetupClusterLifecycleTest sets up cluster lifecycle test.
func SetupClusterLifecycleTest() (*Test, error) {
	testID := model.NewID()
	logger := logrus.WithFields(map[string]interface{}{
		"test":   "cluster-lifecycle",
		"testID": testID,
	})

	config, err := readConfig(logger)
	if err != nil {
		return nil, err
	}

	client := model.NewClient(config.CloudURL)

	createClusterReq := &model.CreateClusterRequest{
		AllowInstallations: true,
		Annotations:        testAnnotations(testID),
	}
	err = clusterdictionary.ApplyToCreateClusterRequest("SizeAlef1000", createClusterReq)
	if err != nil {
		return nil, err
	}

	clusterParams := workflow.ClusterSuiteParams{
		CreateRequest: *createClusterReq,
	}
	installationParams := workflow.InstallationSuiteParams{
		DBType:        config.InstallationDBType,
		FileStoreType: config.InstallationFileStoreType,
		Annotations:   testAnnotations(testID),
	}

	kubeClient, err := pkg.GetK8sClient()
	if err != nil {
		return nil, err
	}

	// We need to be cautious with introducing some parallelism for tests especially on step level
	// as webhook event will be delivered to only one channel.
	webhookChan, cleanup, err := pkg.SetupTestWebhook(client, config.WebhookAddress, testID, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup webhook")
	}

	clusterSuite := workflow.NewClusterSuite(clusterParams, config.Environment, client, webhookChan, logger)
	installationSuite := workflow.NewInstallationSuite(installationParams, config.Environment, client, kubeClient, logger)

	return &Test{
		Logger:            logger,
		WebhookCleanup:    cleanup,
		Workflow:          workflow.NewWorkflow(clusterLifecycleSteps(clusterSuite, installationSuite)),
		ClusterSuite:      clusterSuite,
		InstallationSuite: installationSuite,
		Cleanup:           config.Cleanup,
	}, nil
}

func testAnnotations(testID string) []string {
	return []string{"e2e-test-cluster-lifecycle", fmt.Sprintf("test-id-%s", testID)}
}

func readConfig(logger logrus.FieldLogger) (TestConfig, error) {
	var config TestConfig
	err := envconfig.Init(&config)
	if err != nil {
		return TestConfig{}, errors.Wrap(err, "unable to read environment configuration")
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return TestConfig{}, errors.Wrap(err, "failed to marshal config to json")
	}

	logger.Infof("Test Config: %s", configJSON)

	return config, nil
}

// Run runs the test workflow.
func (w *Test) Run() error {
	err := workflow.RunWorkflow(w.Workflow, w.Logger)
	if err != nil {
		return errors.Wrap(err, "error running workflow")
	}
	return nil
}