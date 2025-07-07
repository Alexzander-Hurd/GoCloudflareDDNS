package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type DnsRecord struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Content    string   `json:"content"`
	Proxiable  bool     `json:"proxiable"`
	Proxied    bool     `json:"proxied"`
	TTL        int      `json:"ttl"`
	Settings   struct{} `json:"settings"`
	Meta       struct{} `json:"meta"`
	Comment    string   `json:"comment"`
	Tags       []string `json:"tags"`
	CreatedOn  string   `json:"created_on"`
	ModifiedOn string   `json:"modified_on"`
}

type DnsRecordResponse struct {
	Success    bool          `json:"success"`
	Errors     []interface{} `json:"errors"`
	Messages   []interface{} `json:"messages"`
	Result     []DnsRecord   `json:"result"`
	ResultInfo struct {
		Count int `json:"count"`
	} `json:"result_info"`
}

func main() {
	loadEnv()
	currentIP := getCurrentIP()
	recordID, recordIP := getRecord()

	current, _ := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, currentIP)
	record, _ := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, recordIP)

	if !current || !record {
		log.Fatal("Error getting current IP or record IP")
	}

	if recordID == "" {
		log.Fatal("Error getting record ID")
	}

	fmt.Println("Current IP:", currentIP)
	fmt.Println("Record IP:", recordIP)
	if currentIP != recordIP {
		updateRecord(recordID, currentIP)
	}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	requiredEnv := []string{"CF_Token", "ZONE_ID", "DOMAIN"}
	for _, key := range requiredEnv {
		if os.Getenv(key) == "" {
			log.Fatalf("Missing required env variable: %s", key)
		}
	}
}

func getCurrentIP() string {
	resp, err := http.Get("http://ipv4.icanhazip.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(body))
}

func getRecord() (string, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones/"+os.Getenv("ZONE_ID")+"/dns_records", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CF_Token"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("API request failed with status: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var data DnsRecordResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range data.Result {
		if v.Name == os.Getenv("DOMAIN") {
			return v.ID, v.Content
		}
	}
	return "", ""
}

func updateRecord(id string, ip string) {
	client := &http.Client{}
	payload := map[string]interface{}{
		"type":    "A",
		"name":    os.Getenv("DOMAIN"),
		"content": ip,
		"ttl":     3600,
		"proxied": true,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("PATCH", "https://api.cloudflare.com/client/v4/zones/"+os.Getenv("ZONE_ID")+"/dns_records/"+id, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CF_Token"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("API request failed with status: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}
