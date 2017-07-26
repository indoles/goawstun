package main

import (
	"fmt"

	"encoding/base64"
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/nu7hatch/gouuid"
)


type queue struct {
	URL string
	svc *sqs.SQS
}

func newQ(url string) *queue {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	
	return &queue{url, sqs.New(sess)}
}

func (q *queue) list() ([]*string, error) {
	result, err := q.svc.ListQueues(nil)
	if err != nil {
		return nil, err
	}

	return result.QueueUrls, nil
}

func (q *queue) send(b []byte) error {
	msg := base64.StdEncoding.EncodeToString(b)
	uuid, err := uuid.NewV4()
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	
	_, err = q.svc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(msg),
		MessageDeduplicationId: aws.String(uuid.String()),
		MessageGroupId: aws.String("TUN"),
		QueueUrl:    &q.URL,
	})

	if err != nil {
		fmt.Println("Error", err)
		return err
	}

	return nil
}

func (q *queue) receive() (bool, [][]byte, error) {
	bodies := make([][]byte, 0)
	result, err := q.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &q.URL,
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(20),
	})

	if err != nil {
		fmt.Println("Error", err)
		return false, nil, err
	}

	if len(result.Messages) == 0 {
		return false, nil, nil
	} else {
		for _, message := range result.Messages {
			_, err := q.svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl: &q.URL,
				ReceiptHandle: message.ReceiptHandle,
			})

			if err != nil {
				fmt.Println("Delete message error:", err)
			}
			body, err := base64.StdEncoding.DecodeString(*message.Body)
			if err != nil {
				fmt.Println("Decoding message error:", err)
				return false, nil, err
			}
			bodies = append(bodies, body)
		}
		return true, bodies, nil
	}
	
}
