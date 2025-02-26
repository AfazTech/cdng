<?php

use Exception;

class CdngApi {
    private $baseUrl;
    private $apiKey;

    public function __construct($baseUrl, $apiKey) {
        $this->baseUrl = rtrim($baseUrl, '/');
        $this->apiKey = $apiKey;
    }

    private function request($method, $endpoint, $data = []) {
        $url = $this->baseUrl . '/' . ltrim($endpoint, '/');
        $ch = curl_init($url);
        
        $headers = [
            'Content-Type: application/json',
            'Authorization: Bearer ' . $this->apiKey
        ];
        
        $options = [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_CUSTOMREQUEST => strtoupper($method),
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_SSL_VERIFYPEER => false,
            CURLOPT_SSL_VERIFYHOST => false,
            CURLOPT_TIMEOUT => 3,
        ];
        
        if ($method === 'POST' || $method === 'PUT') {
            $options[CURLOPT_POSTFIELDS] = json_encode($data);
        }
        
        curl_setopt_array($ch, $options);
        
        $response = curl_exec($ch);
        $statusCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);
        
        if ($error) {
           throw new Exception($error);
        }
        
        $decodedResponse = json_decode($response, true);
        if (!$decodedResponse['ok']){
            throw new Exception(message: $decodedResponse['message']);
        }
        return $decodedResponse;
    }

    public function addDomain($domain, $ip) {
        return $this->request('POST', 'add-domain', [
            'domain' => $domain,
            'ip' => $ip
        ]);
    }

    public function deleteDomain($domain) {
        return $this->request('DELETE', "delete-domain/$domain");
    }

    public function addPort($port) {
        return $this->request('POST', 'add-port', [
            'port' => $port
        ]);
    }

    public function deletePort($port) {
        return $this->request('DELETE', "delete-port/$port");
    }

    public function getStatus() {
        return $this->request('GET', 'status');
    }

    public function reloadNginx() {
        return $this->request('POST', 'reload');
    }

    public function stopNginx() {
        return $this->request('POST', 'stop');
    }

    public function restartNginx() {
        return $this->request('POST', 'restart');
    }

    public function getStats() {
        return $this->request('GET', 'stats');
    }
}

