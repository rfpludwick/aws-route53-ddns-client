<?php
/**
 * AWS DDNS client
 *
 * @category PHP
 * @package  AWS_Route53_DDNS_Client
 * @author   Robert F.P. Ludwick <rfpludwick@gmail.com>
 * @license  Apache 2.0
 */

// Namespacing
namespace RFPLudwick\AWS\DDNS;

use Aws\Credentials\CredentialProvider;
use Aws\Exception\CredentialsException;
use Aws\Route53\Route53Client;
use Aws\Route53\Exception\Route53Exception;
use Monolog\Handler\StreamHandler;
use Monolog\Logger;

/**
 * AWS DDNS client
 *
 * @category PHP
 * @package  AWS_Route53_DDNS_Client
 * @author   Robert F.P. Ludwick <rfpludwick@gmail.com>
 * @license  Apache 2.0
 */
class Client
{
    public function execute()
    {
        // Config setup
        $configDir       = dirname(__DIR__) . '/config/';
        $config          = parse_ini_file($configDir . 'config.ini', true);
        $credentialsPath = $configDir . 'aws_credentials.ini';

        // Logging setup
        $logger = new Logger('ddns');

        try {
            $logger->pushHandler(new StreamHandler($config['logger']['file']));
        } catch (\Exception $exception) {
            exit('Logging setup failed');
        }

        // Collect together Route 53 changes
        $ipAddress = Util\Network::getPublicIpAddress();

        $route53BaseChange = [
            'Action' => 'UPSERT', // UPSERT will update or insert as required
            'ResourceRecordSet' => [
                'Type' => 'A' . ($config['aws_route53']['ipv6'] ? 'AAA' : null),
                'TTL' => $config['aws_route53']['ttl'],
                'ResourceRecords' => [
                    0 => [ // Keyed temporarily until PHPCS' Standard.Arrays.ArrayIndent properly handles keyless nests
                        'Value' => $ipAddress
                    ]
                ]
            ]
        ];

        $route53Changes = [];

        foreach ($config['aws_route53']['record_names'] as $recordName) {
            $route53Change = $route53BaseChange;

            $route53Change['ResourceRecordSet']['Name'] = $recordName;

            $route53Changes[] = $route53Change;
        }

        // Execute AWS Route 53 client
        $provider = CredentialProvider::ini($config['aws']['profile'], $credentialsPath);
        $provider = CredentialProvider::memoize($provider);

        $route53Client = new Route53Client([
            'version' => 'latest',
            'region' => 'us-east-1', // Route 53 requires this region
            'credentials' => $provider,
        ]);

        try {
            $route53Client->changeResourceRecordSets([
                'ChangeBatch' => [
                    'Changes' => $route53Changes,
                    'Comment' => $config['aws_route53']['comment']
                ],
                'HostedZoneId' => $config['aws_route53']['hosted_zone_id']
            ]);
        } catch (Route53Exception $e) {
            $logger->addError($e->getAwsErrorCode() . ': ' . $e->getMessage());

            exit('A "' . $e->getAwsErrorCode() . '" error occurred. Please check the log file for details.');
        } catch (CredentialsException $e) {
            $logger->addError('Invalid AWS credentials provided.');

            exit('Invalid AWS credentials provided: ' . $e->getMessage());
        }

        $logger->addInfo('IP Address updated to ' . $ipAddress);

        exit('IP Address updated successfully to ' . $ipAddress);
    }
}
