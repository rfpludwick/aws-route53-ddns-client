package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"gopkg.in/ini.v1"
)

func main() {
	os.Exit(exec())
}

func exec() int {
	// Read configuration
	config, err := ini.ShadowLoad("config/config.ini")

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to read configuration ini:", err)

		return 1
	}

	// Setup logging
	logfile, err := os.OpenFile(config.Section("logger").Key("file").Value(), (os.O_CREATE | os.O_APPEND | os.O_WRONLY), 0644)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to open logfile:", err)

		return 1
	}

	defer logfile.Close()

	log.SetOutput(logfile)

	// Get our IP address
	c := make(chan string)

	go func() {
		httpResponse, err := http.Get("https://api.ipify.org")

		if err != nil {
			log.Println("Error retriving IP address:", err)
		}

		defer httpResponse.Body.Close()

		responseBodyIPAddress, err := ioutil.ReadAll(httpResponse.Body)

		if err != nil {
			log.Println("Error reading IP address response:", err)
		}

		c <- string(responseBodyIPAddress)
	}()

	// Prepare the Route53 changes
	configRoute53 := config.Section("aws_route53")
	ipAddress := <-c

	var route53Changes []*route53.Change

	route53Action := aws.String("UPSERT")
	route53Ttl := aws.Int64(configRoute53.Key("ttl").MustInt64())
	route53IpAddress := aws.String(ipAddress)

	var route53Type *string

	if configRoute53.Key("ipv6").MustBool() {
		route53Type = aws.String("AAAA")
	} else {
		route53Type = aws.String("A")
	}

	for _, recordName := range configRoute53.Key("record_names[]").ValueWithShadows() {
		route53Changes = append(route53Changes, &route53.Change{
			Action: route53Action,
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(recordName),
				TTL:  route53Ttl,
				Type: route53Type,
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: route53IpAddress,
					},
				},
			},
		})
	}

	route53Request := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: route53Changes,
			Comment: aws.String(configRoute53.Key("comment").String()),
		},
		HostedZoneId: aws.String(configRoute53.Key("hosted_zone_id").String()),
	}

	// Make & handle the Route53 request (after the IP address has been collected)
	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"), // Route 53 requires this region,
		Credentials: credentials.NewSharedCredentials("config/aws_credentials.ini", config.Section("aws").Key("profile").String()),
	})

	if err != nil {
		log.Println("Error instantiating AWS session:", err)

		return 1
	}

	route53Client := route53.New(awsSession)

	_, err = route53Client.ChangeResourceRecordSets(route53Request)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				log.Println(route53.ErrCodeNoSuchHostedZone, aerr.Error())
			case route53.ErrCodeNoSuchHealthCheck:
				log.Println(route53.ErrCodeNoSuchHealthCheck, aerr.Error())
			case route53.ErrCodeInvalidChangeBatch:
				log.Println(route53.ErrCodeInvalidChangeBatch, aerr.Error())
			case route53.ErrCodeInvalidInput:
				log.Println(route53.ErrCodeInvalidInput, aerr.Error())
			case route53.ErrCodePriorRequestNotComplete:
				log.Println(route53.ErrCodePriorRequestNotComplete, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}

		return 1
	}

	log.Println("IP address successfully updated to:", ipAddress)
	fmt.Fprintln(os.Stdout, "IP address successfully updated to:", ipAddress)

	return 0
}
