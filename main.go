package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go"
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

		defer func() {
			if err := httpResponse.Body.Close(); err != nil {
				log.Fatal("Error closing HTTP response body:", err)
			}
		}()

		responseBodyIPAddress, err := io.ReadAll(httpResponse.Body)

		if err != nil {
			log.Fatal("Error reading IP address response:", err)
		}

		c <- string(responseBodyIPAddress)
	}()

	ipAddress := <-c

	// Prepare the Route53 changes
	for _, cfs := range cf.Sites {
		var route53Changes []types.Change
		var route53Type types.RRType

		if cfs.Ipv6 {
			route53Type = types.RRTypeAaaa
		} else {
			route53Type = types.RRTypeA
		}

		for _, recordName := range cfs.RecordNames {
			route53Changes = append(route53Changes, types.Change{
				Action: types.ChangeActionUpsert,
				ResourceRecordSet: &types.ResourceRecordSet{
					Name: aws.String(recordName),
					TTL:  aws.Int64(cfs.Ttl),
					Type: route53Type,
					ResourceRecords: []types.ResourceRecord{
						{
							Value: aws.String(ipAddress),
						},
					},
				},
			})
		}

		route53Request := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &types.ChangeBatch{
				Changes: route53Changes,
				Comment: aws.String(cfs.Comment),
			},
			HostedZoneId: aws.String(cfs.HostedZoneId),
		}

		// Make & handle the Route53 request (after the IP address has been collected)
		ctx := context.Background()
		sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1")) // global region

		if err != nil {
			log.Fatal("Error loading AWS configuration:", err)
		}

		route53Client := route53.NewFromConfig(sdkConfig)

		if _, err = route53Client.ChangeResourceRecordSets(ctx, route53Request); err != nil {
			var apiErr smithy.APIError

			if errors.As(err, &apiErr) {
				log.Fatal(apiErr.ErrorMessage())
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
