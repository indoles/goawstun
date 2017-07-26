Create two AWS SQS Fifo queues.
Compile this program

make sure to go get the dependencies:

golang.org/x/net/ipv4
github.com/songgao/water
github.com/aws/aws-sdk-go/aws
github.com/aws/aws-sdk-go/aws/session
github.com/aws/aws-sdk-go/service/sqs
github.com/nu7hatch/gouuid


On two different hosts, run these commands:

goawstun -local 10.0.1.1 -remote 10.0.2.1 -sendqueue << First AWS SQS URL >> -receivequeue << Second AWS SQS URL >>

goawstun -local 10.0.2.1 -remote 10.0.1.1 -sendqueue << Second AWS SQS URL >> -receivequeue << First AWS SQS URL >>

On either of the hosts, in a different terminal, ssh to the other ip address. (on the host where you executed the second command, ssh <user>@10.0.1.1)
