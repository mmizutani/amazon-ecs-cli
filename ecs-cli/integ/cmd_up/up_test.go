// +build integ

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package cmd_up tests the "ecs-cli up" command with various configurations.
package cmd_up

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	ecsCLIResourcePrefix = "amazon-ecs-cli-setup-"
)

// TestClusterCreation runs the 'ecs-cli up -c <clusterName> --capability-iam --force' command.
//
// If there is no CloudFormation stack created, then the test fails.
func TestClusterCreation(t *testing.T) {
	clusterName := fmt.Sprintf("%s-%d", integ.GetBuildId(), time.Now().Unix())
	cfn, err := newCFNClient()
	// Fail the test immediately if we won't be able to evaluate it
	assert.NoError(t, err)

	runTest(t, cfn, clusterName)

	// Cleanup the created stack
	deleteStack(cfn, clusterName)
}

func runTest(t *testing.T, cfn *cloudformation.CloudFormation, clusterName string) {
	// Given
	cmd := integ.GetCommand([]string{"up", "-c", clusterName, "--capability-iam", "--force"})

	// When
	stdout, err := cmd.Output()
	assert.NoError(t, err, fmt.Sprintf("Error running %v\nStdout: %s", cmd.Args, string(stdout)))

	// Then
	_, err = getStack(cfn, clusterName)
	assert.NoError(t, err)
}

// newCFNClient initializes the CloudFormation client.
func newCFNClient() (*cloudformation.CloudFormation, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new session")
	}
	return cloudformation.New(sess, &aws.Config{
		Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
	}), nil
}

// getStack returns the CloudFormation stack created by the up command.
func getStack(cfn *cloudformation.CloudFormation, clusterName string) (*cloudformation.Stack, error) {
	resp, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName(clusterName)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unexpected cloudformation error")
	}
	if resp.Stacks == nil || len(resp.Stacks) == 0 {
		return nil, errors.New(fmt.Sprintf("no stack named '%s' found", stackName(clusterName)))
	}
	return resp.Stacks[0], nil
}

// deleteStack best-effort deletes any resources created by the test.
func deleteStack(cfn *cloudformation.CloudFormation, clusterName string) {
	cfn.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName(clusterName)),
	})
}

func stackName(clusterName string) string {
	return ecsCLIResourcePrefix + clusterName
}
