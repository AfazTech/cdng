package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	netStat "github.com/shirou/gopsutil/net"
)

const (
	domainConfigPath = "/etc/nginx/conf.d/domains/"
	listenConfigPath = "/etc/nginx/conf.d/listen.conf"
)

func AddDomain(domain, ip string) error {
	if !isValidDomain(domain) {
		return errors.New("Invalid domain format")
	}
	if net.ParseIP(ip) == nil {
		return errors.New("Invalid IP format")
	}

	// Attempt to obtain SSL certificate
	cmd := exec.Command("certbot", "certonly", "--standalone", "-d", domain, "--non-interactive", "--agree-tos", "-m", "admin@"+domain)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to obtain SSL certificate: %v", err)
	}

	config := fmt.Sprintf(`server {
	include /etc/nginx/conf.d/listen.conf;
	
	server_name %s;

    ssl_certificate /etc/letsencrypt/live/%s/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/%s/privkey.pem;

    location / {
        proxy_pass $scheme://%s:$server_port;
    }
}`, domain, domain, domain, ip)

	if err := ioutil.WriteFile(domainConfigPath+domain+".conf", []byte(config), 0644); err != nil {
		return fmt.Errorf("Failed to write domain configuration: %v", err)
	}

	return ReloadNginx()
}

func DeleteDomain(domain string) error {
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

	newConfig := strings.Replace(string(data), "listen "+port+";\n", "", -1)
	if err := ioutil.WriteFile(listenConfigPath, []byte(newConfig), 0644); err != nil {
		return fmt.Errorf("Failed to remove port from configuration: %v", err)
	}

	return ReloadNginx()
}

func ReloadNginx() error {
	cmd := exec.Command("nginx", "-s", "reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to reload Nginx: %v", err)
	}
	return nil
}

func GetStats() (map[string]interface{}, error) {
	ports, err := ioutil.ReadFile(listenConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read ports configuration: %v", err)
	}

	domains, err := ioutil.ReadDir(domainConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read domain configurations: %v", err)
	}

	cpuLoad, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("Failed to read CPU load: %v", err)
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("Failed to read memory info: %v", err)
	}

	netStats, err := netStat.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("Failed to get network statistics: %v", err)
	}

	stats := map[string]interface{}{
		"ports":          strings.Count(string(ports), "listen "),
		"domain_count":   len(domains),
		"cpu_load":       cpuLoad.Load1,
		"total_ram":      memInfo.Total / 1024 / 1024,
		"used_ram":       memInfo.Used / 1024 / 1024,
		"cpu_usage":      memInfo.UsedPercent,
		"upload_bytes":   netStats[0].BytesSent,
		"download_bytes": netStats[0].BytesRecv,
	}

	return stats, nil
}

func RestartNginx() error {
	cmd := exec.Command("systemctl", "restart", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to restart Nginx: %v", err)
	}
	return nil
}

func isValidDomain(domain string) bool {
	return strings.Contains(domain, ".") && len(domain) > 3 && len(domain) < 255
}
