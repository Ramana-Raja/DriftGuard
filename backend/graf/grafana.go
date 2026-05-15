package graf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const GrafanaURL = "http://grafana:3000"
const GrafanaAdminAuth = "Basic YWRtaW46YWRtaW4="

func CreateGrafanaOrg(projectID string, projectName string) (int, error) {
	uniqueOrgName := fmt.Sprintf("%s (ID: %s)", projectName, projectID)

	payload := map[string]string{"name": uniqueOrgName}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", GrafanaURL+"/api/orgs", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Authorization", GrafanaAdminAuth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("network error connecting to Grafana: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("grafana rejected org creation: %s", string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	orgId := int(result["orgId"].(float64))
	return orgId, nil
}

func CreatePrometheusDataSource(orgID int, projectID string) error {
	payload := map[string]interface{}{
		"name":      fmt.Sprintf("Mimir-Project-%s", projectID),
		"type":      "prometheus",
		"url":       "http://prometheus:9090",
		"access":    "proxy",
		"isDefault": true,
		"jsonData": map[string]interface{}{
			"httpHeaderName1": "X-Scope-OrgID",
		},
		"secureJsonData": map[string]interface{}{
			"httpHeaderValue1": projectID,
		},
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", GrafanaURL+"/api/datasources", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Authorization", GrafanaAdminAuth)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("X-Grafana-Org-Id", fmt.Sprintf("%d", orgID))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("failed to create data source for org %d", orgID)
	}

	return nil
}

func ProvisionWorkspace(projectID string, projectName string) error {
	orgID, err := CreateGrafanaOrg(projectID, projectName)
	if err != nil {
		return err
	}

	err = CreatePrometheusDataSource(orgID, projectID)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully provisioned Grafana Workspace for: %s (DB ID: %s, Grafana Org ID: %d)\n", projectName, projectID, orgID)
	return nil
}
