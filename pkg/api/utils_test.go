package api

import "testing"

func TestStringifyQueryParams(t *testing.T) {
	client := &Client{BaseURL: "https://api.example.com"}

	tests := []struct {
		endpoint    string
		queryParams QueryParams
		expectedURL string
		expectError bool
	}{
		{
			endpoint:    "/endpoint",
			queryParams: QueryParams{"param1": "value1"},
			expectedURL: "https://api.example.com/endpoint?param1=value1",
			expectError: false,
		},
		{
			endpoint:    "/another-endpoint",
			queryParams: QueryParams{"param2": []string{"value2a", "value2b"}},
			expectedURL: "https://api.example.com/another-endpoint?param2=value2a&param2=value2b",
			expectError: false,
		},
		{
			endpoint:    "/incident",
			queryParams: QueryParams{"status": IncidentActiveStatusList()},
			expectedURL: "https://api.example.com/incident?status=incident_investigating&status=incident_identified&status=incident_monitoring",
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.endpoint, func(t *testing.T) {
			url, err := client.stringifyQueryParams(test.endpoint, test.queryParams)

			if (err != nil) != test.expectError {
				t.Errorf("expected error: %v, got: %v", test.expectError, err)
			}
			if err == nil && url.String() != test.expectedURL {
				t.Errorf("expected URL: %s, got: %s", test.expectedURL, url.String())
			}
		})
	}
}
