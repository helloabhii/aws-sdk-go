package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	var (
		instanceId string
		err        error
	)

	ctx := context.Background()
	if instanceId, err = createEC2(ctx, "us-east-1"); err != nil {
		fmt.Printf("createEC2 error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Instance id: %s", instanceId)
}

func createEC2(ctx context.Context, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %s", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	keyPairs, err := ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{"go-aws-demo"},
	})
	if err != nil && !strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
		return "", fmt.Errorf("DescribeKeyPairs error:  %s", err)
	}

	if keyPairs == nil || len(keyPairs.KeyPairs) == 0 {
		keyPair, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
			KeyName: aws.String("go-aws-demo"),
		})
		if err != nil {
			return "", fmt.Errorf("CreateKeyPair error: %s", err)
		}
		err = os.WriteFile("go-aws-ec2.pem", []byte(*keyPair.KeyMaterial), 0600)
		if err != nil {
			return "", fmt.Errorf("CreateKeyPair error: %s", err)
		}

	}

	imageOutput, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"},
			},
			{
				Name:   aws.String("virtualization-type"),
				Values: []string{"hvm"},
			},
		},
		Owners: []string{"099720109477"},
	})
	if err != nil {
		return "", fmt.Errorf("DescribeImages error : %s", err)
	}

	if len(imageOutput.Images) == 0 {
		return "", fmt.Errorf("imageOutput.Images is of 0 length")
	}

	instance, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      imageOutput.Images[0].ImageId,
		KeyName:      aws.String("go-aws-demo"),
		InstanceType: types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		return "", fmt.Errorf("RunInstances error: %s", err)
	}
	// imageOutput.Images[0].ImageId

	if len(instance.Instances) == 0 {
		return "", fmt.Errorf("Instance.Instances is of 0 length")
	}
	return *instance.Instances[0].InstanceId, nil
}
