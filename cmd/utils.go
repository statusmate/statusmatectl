package cmd

import (
	"regexp"
)

const (
	TypeUUID    = "UUID"
	TypeID      = "ID"
	TypeDomain  = "Domain"
	TypeUnknown = "Unknown"
)

func isUUID(s string) bool {
	uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidRegex, s)
	return matched
}

func isID(s string) bool {
	idRegex := `^\d+$`
	matched, _ := regexp.MatchString(idRegex, s)
	return matched
}

func isDomain(s string) bool {
	domainRegex := `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(domainRegex, s)
	return matched
}

func IdentifyType(s string) string {
	if isUUID(s) {
		return "UUID"
	} else if isID(s) {
		return "ID"
	} else if isDomain(s) {
		return "Domain"
	}
	return "Unknown"
}
