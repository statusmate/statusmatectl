package format

import (
	"reflect"
	"strings"
	"testing"
)

var CommentsMap = map[string]string{
	"name": "Название инцидента",
	"desc": "Описание инцидента",
	"status": `Available statuses:
- incident_investigating
- incident_identified
- incident_monitoring
- incident_resolved`,
	"affected": `Impacts:
- o, op - operational
- u, um - under maintenance
- d, dp - degraded performance
- p, po - partial outage
- m, mo - major outage`,
}

type Incident struct {
	Name        string   `format:"name"`
	Description string   `format:"desc"`
	Status      string   `format:"status"`
	Notify      bool     `format:"notify"`
	Affected    []string `format:"affected"`
}

// Тест на проверку корректного маршалинга структуры с комментариями
func TestMarshal(t *testing.T) {
	incident := Incident{
		Name: "Test Incident",
		Description: `Test Description
Second string`,
		Notify:   true,
		Status:   "incident_identifying",
		Affected: []string{"op cloud", "um cdn"},
	}

	expected := `# Название инцидента
[name]
Test Incident

# Описание инцидента
[desc]
Test Description
Second string

# Available statuses:
# - incident_investigating
# - incident_identified
# - incident_monitoring
# - incident_resolved
[status]
incident_identifying

[notify]
yes

# Impacts:
# - o, op - operational
# - u, um - under maintenance
# - d, dp - degraded performance
# - p, po - partial outage
# - m, mo - major outage
[affected]
op cloud
um cdn

`

	marshaled, err := Marshal(&incident, &CommentsMap)
	if err != nil {
		t.Errorf("unexpected error during marshaling: %v", err)
	}

	if marshaled != expected {
		t.Errorf("unexpected output: got %v, expected %v", marshaled, expected)
	}
}

// Тест на проверку корректного анмаршалинга строки с комментариями
func TestUnmarshal(t *testing.T) {
	data := `
# Название инцидента
[name]
Test Incident

# Описание инцидента
[desc]
Test Description
Test second string

# Available statuses:
# - incident_investigating
# - incident_identified
# - incident_monitoring
# - incident_resolved
[status]
incident_investigating

# Impacts:
# - o, op - operational
# - u, um - under maintenance
# - d, dp - degraded performance
# - p, po - partial outage
# - m, mo - major outage
[affected]
op cloud
um cdn
`

	expected := Incident{
		Name:        "Test Incident",
		Description: "Test Description\nTest second string",
		Status:      "incident_investigating",
		Notify:      false,
		Affected:    []string{"op cloud", "um cdn"},
	}

	var incident Incident
	err := Unmarshal(data, &incident)
	if err != nil {
		t.Errorf("unexpected error during unmarshaling: %v", err)
	}

	if !reflect.DeepEqual(incident, expected) {
		t.Errorf("unexpected unmarshaled struct: got %+v, expected %+v", incident, expected)
	}
}

// Тест на проверку корректной обработки комментариев
func TestUnmarshalWithComments(t *testing.T) {
	data := `
# This is a comment
[name]
Test Incident

# Another comment here
[desc]
Test Description

[status]
incident_investigating

# Comment for affected section
[affected]
op cloud
um cdn
`

	expected := Incident{
		Name:        "Test Incident",
		Description: "Test Description",
		Status:      "incident_investigating",
		Affected:    []string{"op cloud", "um cdn"},
	}

	var incident Incident
	err := Unmarshal(data, &incident)
	if err != nil {
		t.Errorf("unexpected error during unmarshaling: %v", err)
	}

	if !reflect.DeepEqual(incident, expected) {
		t.Errorf("unexpected unmarshaled struct: got %+v, expected %+v", incident, expected)
	}
}

// Тест на пустые поля при анмаршалинге
func TestUnmarshalEmptyFields(t *testing.T) {
	data := `
[name]

[desc]

[status]

[affected]
`

	expected := Incident{
		Name:        "",
		Description: "",
		Status:      "",
		Affected:    []string{""},
	}

	var incident Incident
	err := Unmarshal(data, &incident)
	if err != nil {
		t.Errorf("unexpected error during unmarshaling: %v", err)
	}

	if !reflect.DeepEqual(incident, expected) {
		t.Errorf("unexpected unmarshaled struct: got %+v, expected %+v", incident, expected)
	}
}

// Тест на обработку пустого массива Affected
func TestUnmarshalEmptyAffected(t *testing.T) {
	data := `
[name]
Test Incident

[desc]
Test Description

[status]
incident_resolved

[affected]
`

	expected := Incident{
		Name:        "Test Incident",
		Description: "Test Description",
		Status:      "incident_resolved",
		Affected:    []string{""},
	}

	var incident Incident
	err := Unmarshal(data, &incident)
	if err != nil {
		t.Errorf("unexpected error during unmarshaling: %v", err)
	}

	if !reflect.DeepEqual(incident, expected) {
		t.Errorf("unexpected unmarshaled struct: got %+v, expected %+v", incident, expected)
	}
}

// Тест на корректную обработку всех комментариев при маршалинге
func TestMarshalWithComments(t *testing.T) {
	incident := Incident{
		Name:        "Test Incident",
		Description: "Test Description",
		Status:      "incident_identifying",
		Affected:    []string{"op cloud", "um cdn"},
	}

	marshaled, err := Marshal(&incident, &CommentsMap)
	if err != nil {
		t.Errorf("unexpected error during marshaling: %v", err)
	}

	if !strings.Contains(marshaled, "# Название инцидента") ||
		!strings.Contains(marshaled, "# Описание инцидента") {
		t.Errorf("comments not found in output: got %v", marshaled)
	}
}

// Тест на корректную обработку без комментариев при маршалинге
func TestMarshalWithoutComments(t *testing.T) {
	incident := Incident{
		Name:        "Test Incident",
		Description: "Test Description",
		Status:      "incident_identifying",
		Affected:    []string{"op cloud", "um cdn"},
	}

	expected := `[name]
Test Incident

[desc]
Test Description

[status]
incident_identifying

[notify]
no

[affected]
op cloud
um cdn

`

	marshaled, err := Marshal(&incident, nil)
	if err != nil {
		t.Errorf("unexpected error during marshaling: %v", err)
	}

	if !reflect.DeepEqual(marshaled, expected) {
		t.Errorf("unexpected unmarshaled struct: got %+v, expected %+v", marshaled, expected)
	}
}
