package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"gopkg.in/yaml.v3"
)

type ConfigurationFile struct {
	GetIpAddressHost string                  `yaml:"get_ip_address_host"`
	Sites            []ConfigurationFileSite `yaml:"sites"`
}

type ConfigurationFileSite struct {
	HostedZoneId string   `yaml:"hosted_zone_id"`
	Ttl          int64    `yaml:"ttl"`
	Ipv6         bool     `yaml:"ipv6"`
	Comment      string   `yaml:"comment"`
	RecordNames  []string `yaml:"record_names"`
}

var (
	flagConfigurationFile string
)

func init() {
	flag.StringVar(&flagConfigurationFile, "config", "conf/configuration.yaml", "Configuration file to use")
}

func main() {
	flag.Parse()

	cf := processConfiguration()

	// Get our IP address
	c := make(chan string)

	go func() {
		httpResponse, err := http.Get(cf.GetIpAddressHost)

		if err != nil {
			log.Fatal("Error retrieving IP address:", err)
		}

		defer httpResponse.Body.Close()

		responseBodyIPAddress, err := io.ReadAll(httpResponse.Body)

		if err != nil {
			log.Fatal("Error reading IP address response:", err)
		}

		c <- string(responseBodyIPAddress)
	}()

	ipAddress := <-c

	// Prepare the Route53 changes
	for _, cfs := range cf.Sites {
		var route53Changes []*route53.Change
		var route53Type *string

		if cfs.Ipv6 {
			route53Type = aws.String("AAAA")
		} else {
			route53Type = aws.String("A")
		}

		for _, recordName := range cfs.RecordNames {
			route53Changes = append(route53Changes, &route53.Change{
				Action: aws.String("UPSERT"),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(recordName),
					TTL:  aws.Int64(cfs.Ttl),
					Type: route53Type,
					ResourceRecords: []*route53.ResourceRecord{
						{
							Value: aws.String(ipAddress),
						},
					},
				},
			})
		}

		route53Request := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: route53Changes,
				Comment: aws.String(cfs.Comment),
			},
			HostedZoneId: aws.String(cfs.HostedZoneId),
		}

		// Make & handle the Route53 request (after the IP address has been collected)
		awsSession, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"), // Route 53 requires this region,
		})

		if err != nil {
			log.Fatal("Error instantiating AWS session:", err)
		}

		route53Client := route53.New(awsSession)

		if _, err = route53Client.ChangeResourceRecordSets(route53Request); err != nil {
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
					log.Fatal(aerr.Error())
				}
			} else {
				log.Fatal(err.Error())
			}
		}
	}

	log.Println("IP address successfully updated to:", ipAddress)
}

func processConfiguration() *ConfigurationFile {
	configurationFileBody, err := os.ReadFile(flagConfigurationFile)

	if err != nil {
		log.Fatal("Error reading configuration file:", err)
	}

	var cf ConfigurationFile

	if err = yaml.Unmarshal(configurationFileBody, &cf); err != nil {
		log.Fatal("Error decoding configuration YAML:", err)
	}

	if cf.GetIpAddressHost == "" {
		cf.GetIpAddressHost = "https://api.ipify.org"
	}

	for index, cfs := range cf.Sites {
		if cfs.Ttl == 0 {
			cfs.Ttl = 300
		}

		if cfs.Comment == "" {
			cfs.Comment = "AWS Route53 DDNS Client"
		}

		cf.Sites[index] = cfs
	}

	return &cf
}
