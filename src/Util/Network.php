<?php
/**
 * Network utilities
 *
 * @category PHP
 * @package  AWS_Route53_DDNS_Client
 * @author   Robert F.P. Ludwick <rfpludwick@gmail.com>
 * @license  Apache 2.0
 */

// Namespacing
namespace RFPLudwick\AWS\DDNS\Util;

use \GuzzleHttp\Client;

/**
 * Network utilities
 *
 * @category PHP
 * @package  AWS_Route53_DDNS_Client
 * @author   Robert F.P. Ludwick <rfpludwick@gmail.com>
 * @license  Apache 2.0
 */
class Network
{
    /**
     * Returns the public IP address
     *
     * @return string The public IP address
     */
    public static function getPublicIpAddress()
    {
        return (new Client)
            ->request('GET', 'https://api.ipify.org')
            ->getBody()
            ->getContents();
    }
}
