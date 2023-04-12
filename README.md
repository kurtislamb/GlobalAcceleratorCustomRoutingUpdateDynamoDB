# GlobalAcceleratorCustomRoutingUpdateDynamoDB

Update a DynamoDB Table With Global Accelerator Custom Routing Mappings

This GO application is a useful tool for copying the mapping from a Global Accelerator Custom Router into a more read friendly DynamoDB Table. This code can be extended quite comfortably to support other database's such as any RDS variant.

## DynamoDB

I recommend a `paritionkey` of `DestinationIP (S)` and a `Sortkey` of `DestinationPort (N)`.
It is worth noting that while Global Accelerator must operate out of `us-west-2` the DynamoDB table does not have be in that region, as such providing dynamoTableRegion is required.

## Build go

```shell
cd go/globalacceleratormapping
go build .
```

## Run Application

```shell
./globalacceleratormapping -dynamoTableName "<Dynamodb Table Name>" \
 -dynamoTableRegion "<region>" \
 -acceleratorArn "<Global Accelerator ARN>" \
 -endpointGroupArn "<Global Accelerator Endpoint Group ARN>"
 ```
