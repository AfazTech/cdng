package controller

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	netStat "github.com/shirou/gopsutil/net"
)

const (
	domainConfigPath = "/etc/nginx/conf.d/domains/"
	listenConfigPath = "/etc/nginx/conf.d/listen.conf"
)

type Stats struct {
	Ports       []string `json:"ports"`
	DomainCount int      `json:"domain_count"`
	CPULoad     string   `json:"cpu_load"`
	RAMUsage    string   `json:"ram_usage"`
	CPUUsage    string   `json:"cpu_usage"`
	Upload      string   `json:"upload"`
	Download    string   `json:"download"`
}

func AddDomain(domain, ip string) error {
	if !isValidDomain(domain) {
		return errors.New("Invalid domain format")
	}
	if net.ParseIP(ip) == nil {
		return errors.New("Invalid IP format")
	}

	if _, err := os.Stat(domainConfigPath + domain + ".conf"); err == nil {
		return fmt.Errorf("Domain %s already exists", domain)
	}

	config := fmt.Sprintf(`server {
    include /etc/nginx/conf.d/listen.conf;
    
    server_name %s;

    ssl_certificate /etc/nginx/ssl/dummy.crt;
    ssl_certificate_key /etc/nginx/ssl/dummy.key;

location /.well-known/acme-challenge/ {
        root /var/www/html;
        try_files $uri $uri/ =404;
    }
}`, domain)

	if err := ioutil.WriteFile(domainConfigPath+domain+".conf", []byte(config), 0644); err != nil {
		return fmt.Errorf("Failed to write domain configuration: %v", err)
	}

	if err := ReloadNginx(); err != nil {
		os.Remove(domainConfigPath + domain + ".conf")
		return fmt.Errorf("Failed to reload Nginx after adding domain: %v", err)
	}
	cmd := exec.Command("certbot", "certonly", "--webroot", "-w", "/var/www/html",
		"-d", domain, "--preferred-challenges", "http", "--non-interactive",
		"--agree-tos", "-m", "admin@"+domain)

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(domainConfigPath + domain + ".conf")
		ReloadNginx()
		return fmt.Errorf("Failed to obtain SSL certificate: %v\n%s", err, string(output))
	}
	fmt.Println("Certbot output:", string(output))

	sslConfig := fmt.Sprintf(`server {
    include /etc/nginx/conf.d/listen.conf;
    
    server_name %s;

    ssl_certificate /etc/letsencrypt/live/%s/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/%s/privkey.pem;

    location / {
        proxy_pass $scheme://%s:$server_port;
    }
}`, domain, domain, domain, ip)

	if err := ioutil.WriteFile(domainConfigPath+domain+".conf", []byte(sslConfig), 0644); err != nil {
		os.Remove(domainConfigPath + domain + ".conf")
		ReloadNginx()
		return fmt.Errorf("Failed to write SSL domain configuration: %v", err)
	}

	if err := ReloadNginx(); err != nil {
		os.Remove(domainConfigPath + domain + ".conf")
		ReloadNginx()
		return fmt.Errorf("Failed to reload Nginx with SSL configuration: %v", err)
	}

	return nil
}

func DeleteDomain(domain string) error {
	if _, err := os.Stat(domainConfigPath + domain + ".conf"); os.IsNotExist(err) {
		return fmt.Errorf("Domain %s does not exist", domain)
	}

	if err := os.Remove(domainConfigPath + domain + ".conf"); err != nil {
		return fmt.Errorf("Failed to delete domain configuration: %v", err)
	}
	return ReloadNginx()
}

func AddPort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return errors.New("Invalid port format")
	}

	data, err := ioutil.ReadFile(listenConfigPath)
	if err == nil && strings.Contains(string(data), "listen "+port+";") {
		return fmt.Errorf("Port %s is already added", port)
	}

	file, err := os.OpenFile(listenConfigPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open listen configuration: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString("listen " + port + ";\n"); err != nil {
		return fmt.Errorf("Failed to add port to configuration: %v", err)
	}

	return ReloadNginx()
}

func DeletePort(port string) error {
	data, err := ioutil.ReadFile(listenConfigPath)
	if err != nil {
		return fmt.Errorf("Failed to read listen configuration: %v", err)
	}

	if !strings.Contains(string(data), "listen "+port+";") {
		return fmt.Errorf("Port %s does not exist", port)
	}

	newConfig := strings.Replace(string(data), "listen "+port+";\n", "", -1)
	if err := ioutil.WriteFile(listenConfigPath, []byte(newConfig), 0644); err != nil {
		return fmt.Errorf("Failed to remove port from configuration: %v", err)
	}

	return ReloadNginx()
}

func ReloadNginx() error {
	cmd := exec.Command("nginx", "-s", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func RestartNginx() error {
	cmd := exec.Command("systemctl", "restart", "nginx")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func GetStats() (string, error) {

	ports, err := getListeningPorts()
	if err != nil {
		return "", err
	}
	domains, err := ioutil.ReadDir(domainConfigPath)
	if err != nil {
		return "", fmt.Errorf("Failed to read domain configurations: %v", err)
	}

	cpuLoad, err := load.Avg()
	if err != nil {
		return "", fmt.Errorf("Failed to read CPU load: %v", err)
	}

	cpuCores, err := cpu.Counts(true)
	if err != nil {
		return "", fmt.Errorf("Failed to get CPU core count: %v", err)
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return "", fmt.Errorf("Failed to read memory info: %v", err)
	}

	initialNetStats, err := netStat.IOCounters(false)
	if err != nil {
		return "", fmt.Errorf("Failed to get initial network statistics: %v", err)
	}
	initialSent := initialNetStats[0].BytesSent
	initialRecv := initialNetStats[0].BytesRecv

	time.Sleep(1 * time.Second)

	finalNetStats, err := netStat.IOCounters(false)
	if err != nil {
		return "", fmt.Errorf("Failed to get final network statistics: %v", err)
	}
	finalSent := finalNetStats[0].BytesSent
	finalRecv := finalNetStats[0].BytesRecv

	uploadSpeed := float64(finalSent-initialSent) * 8 / 1_000_000
	downloadSpeed := float64(finalRecv-initialRecv) * 8 / 1_000_000

	stats := Stats{
		Ports:       ports,
		DomainCount: len(domains),
		CPULoad:     fmt.Sprintf("%.2f/%d", cpuLoad.Load1, cpuCores),
		RAMUsage:    fmt.Sprintf("%d/%d MB", memInfo.Used/1024/1024, memInfo.Total/1024/1024),
		CPUUsage:    fmt.Sprintf("%d%%", int(memInfo.UsedPercent)),
		Upload:      fmt.Sprintf("%.2f/Mb", uploadSpeed),
		Download:    fmt.Sprintf("%.2f/Mb", downloadSpeed),
	}

	jsonOutput, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", fmt.Errorf("Failed to marshal stats to JSON: %v", err)
	}

	return string(jsonOutput), nil
}

func isValidDomain(domain string) bool {
	return strings.Contains(domain, ".") && len(domain) > 3 && len(domain) < 255
}

func getListeningPorts() ([]string, error) {
	file, err := os.Open(listenConfigPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []string
	reg := regexp.MustCompile(`listen\s+(\d+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := reg.FindStringSubmatch(line)
		if len(matches) > 1 {
			ports = append(ports, matches[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ports, nil
}

func StatusNginx() (string, error) {

	cmd := exec.Command("systemctl", "is-active", "nginx")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get Nginx status: %v", err)
	}
	return string(output), nil
}

func StopNginx() error {
	cmd := exec.Command("systemctl", "stop", "nginx")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop Bind: %w | output: %s", err, string(output))
	}
	return nil
}

func StartNginx() error {
	cmd := exec.Command("systemctl", "start", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to start nginx: %v", err)
	}
	return nil
}
