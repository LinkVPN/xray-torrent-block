package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
	"torrents-blocker/config"

	"github.com/hpcloud/tail"
)

func StartLogMonitor() {
	t, err := tail.TailFile(config.LogFile, tail.Config{
		Follow:    true,
		ReOpen:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
	})
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	for line := range t.Lines {
		if strings.Contains(line.Text, config.TorrentTag) {
			handleLogEntry(line.Text)
		}
	}
}

func handleLogEntry(line string) {
	ip := config.IpRegex.FindString(line)
	var tid []string
	var username []string

	if config.TidRegex != nil {
		tid = config.TidRegex.FindStringSubmatch(line)
	}

	if config.UsernameRegex != nil {
		username = config.UsernameRegex.FindStringSubmatch(line)
	}

	if ip == "" || len(username) < 2 {
		log.Println("Invalid log entry format: IP or username missing")
		return
	}

	config.Mu.Lock()
	defer config.Mu.Unlock()

	if config.BlockedIPs[ip] {
		log.Printf("User %s with IP: %s is already blocked. Skipping...\n", username[1], ip)
		return
	}
	config.BlockedIPs[ip] = true

	if config.SendUserMessage {
		go SendTelegramMessage(tid[1], config.Message, config.BotToken, "HTML", true)
	}

	if config.SendAdminMessage {
		adminMsg := fmt.Sprintf(config.AdminBlockTemplate, username[1], ip, config.Hostname, username[1])
		go SendTelegramMessage(config.AdminChatID, adminMsg, config.AdminBotToken, "HTML", true)
	}

	go BlockIP(ip)
	log.Printf("User %s with IP: %s blocked for %d minutes\n", username[1], ip, config.BlockDuration)

	go UnblockIPAfterDelay(ip, time.Duration(config.BlockDuration)*time.Minute, username[1])
}

func ScheduleBlockedIPsUpdate() {
	UpdateBlockedIPs()
	go func() {
		for range time.Tick(time.Duration(config.BlockDuration) * time.Minute) {
			UpdateBlockedIPs()
		}
	}()
}

func UpdateBlockedIPs() {
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error checking ufw status: %v", err)
		return
	}

	config.Mu.Lock()
	defer config.Mu.Unlock()

	config.BlockedIPs = make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		ip := config.IpRegex.FindString(line)
		if ip != "" {
			config.BlockedIPs[ip] = true
		}
	}
}

func SendTelegramMessage(chatID string, message string, botToken string, parseMode string, disablePreview bool) {
	urlStr := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", message)
	data.Set("parse_mode", parseMode)
	if disablePreview {
		data.Set("disable_web_page_preview", "true")
	}

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
	}
}

func BlockIP(ip string) {
	var cmd *exec.Cmd

	if config.BlockMode == "iptables" {
		cmd = exec.Command("iptables", "-t", "raw", "-I", "PREROUTING", "-s", ip, "-j", "DROP")
	} else {
		cmd = exec.Command("ufw", "insert", "1", "deny", "from", ip, "to", "any")
	}

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error blocking IP: %v", err)
	}
}

func UnblockIPAfterDelay(ip string, delay time.Duration, username string) {
	time.Sleep(delay)

	var cmd *exec.Cmd

	if config.BlockMode == "iptables" {
		cmd = exec.Command("iptables", "-t", "raw", "-D", "PREROUTING", "-s", ip, "-j", "DROP")
	} else {
		cmd = exec.Command("ufw", "delete", "deny", "from", ip, "to", "any")
	}

	err := cmd.Run()
	if err != nil {
		log.Printf("Error unblocking IP: %v", err)
		return
	}

	config.Mu.Lock()
	delete(config.BlockedIPs, ip)
	config.Mu.Unlock()

	log.Printf("User %s with IP: %s has been unblocked\n", username, ip)

	if config.SendAdminMessage {
		adminMsg := fmt.Sprintf(config.AdminUnblockTemplate, username, ip, config.Hostname, username)
		go SendTelegramMessage(config.AdminChatID, adminMsg, config.AdminBotToken, "HTML", true)
	}
}
