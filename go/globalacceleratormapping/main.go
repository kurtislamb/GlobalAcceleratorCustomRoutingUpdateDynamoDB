package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
)

type routingMap struct {
	ExternalIP      []string
	ExternalPort    int32
	DestinationIP   string
	DestinationPort int32
}

func main() {

	var (
		dynamoTableName   = flag.String("dynamoTableName", "", "The name of the dynamoDB table to use")
		dynamoTableRegion = flag.String("dynamoTableRegion", "", "The region of the dynamoDB table to use")
		acceleratorArn    = flag.String("acceleratorArn", "", "The ARN of the Global Accelerator Custom Router to read from")
		endpointGroupArn  = flag.String("endpointGroupArn", "", "The ARN of the Global Accelerator endpoint to read from")
	)

	flag.Parse()

	if *dynamoTableName == "" || *dynamoTableRegion == "" || *acceleratorArn == "" || *endpointGroupArn == "" {
		flag.Usage()
	}

	var MaxResults int32 = 200
	var ipAddresses []string

	ctx := context.Background()

	cfgUSWest2, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-west-2"))
	if err != nil {
		errorExit("Loading AWS config: %s", err)
	}

	cfgDDB, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(*dynamoTableRegion))
	if err != nil {
		errorExit("Loading AWS config: %s", err)
	}

	globalAcceleratorSVC := globalaccelerator.NewFromConfig(cfgUSWest2)
	dynamodbClient := dynamodb.NewFromConfig(cfgDDB)

	DescribeCustomRoutingAcceleratorInput := globalaccelerator.DescribeCustomRoutingAcceleratorInput{
		AcceleratorArn: acceleratorArn,
	}
	customRouter, err := globalAcceleratorSVC.DescribeCustomRoutingAccelerator(ctx, &DescribeCustomRoutingAcceleratorInput)
	if err != nil {
		errorExit("Check Existing Global Accelerators: %s", err)
	}

	for _, ip := range customRouter.Accelerator.IpSets {
		ipAddresses = ip.IpAddresses
	}

	ListCustomRoutingPortMappingsInput := globalaccelerator.ListCustomRoutingPortMappingsInput{
		AcceleratorArn:   acceleratorArn,
		EndpointGroupArn: endpointGroupArn,
		MaxResults:       &MaxResults,
	}

	loop := 0
	for loop != 1 {
		log.Printf("Updating %d records", MaxResults)
		mappings, err := globalAcceleratorSVC.ListCustomRoutingPortMappings(ctx, &ListCustomRoutingPortMappingsInput)
		if err != nil {
			errorExit("Check Existing Global Accelerators: %s", err)
		}
		for _, v := range mappings.PortMappings {

			routingMap := routingMap{
				ExternalIP:      ipAddresses,
				ExternalPort:    *v.AcceleratorPort,
				DestinationIP:   *v.DestinationSocketAddress.IpAddress,
				DestinationPort: *v.DestinationSocketAddress.Port,
			}

			item, err := attributevalue.MarshalMap(routingMap)
			if err != nil {
				errorExit("creating item: %s", err)
			}

			DynamoDBPutItemInput := dynamodb.PutItemInput{
				TableName: aws.String(*dynamoTableName),
				Item:      item,
			}

			_, err = dynamodbClient.PutItem(ctx, &DynamoDBPutItemInput)
			if err != nil {
				errorExit("putting mapping into dynamodb: %s", err)
			}
		}
		ListCustomRoutingPortMappingsInput.NextToken = mappings.NextToken

		if mappings.NextToken == nil {
			loop = 1
		}
	}
}

func errorExit(format string, v ...any) {
	log.Printf(format, v...)
	os.Exit(1)
}
