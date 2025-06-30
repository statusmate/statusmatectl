package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type StatusPageThemeEnum string

const (
	Theme1 StatusPageThemeEnum = "theme_1"
	Theme2 StatusPageThemeEnum = "theme_2"
)

type StatusPageBgColorEnum string

const (
	Slate   StatusPageBgColorEnum = "slate"
	Zinc    StatusPageBgColorEnum = "zinc"
	Gray    StatusPageBgColorEnum = "gray"
	Neutral StatusPageBgColorEnum = "neutral"
	Stone   StatusPageBgColorEnum = "stone"
)

type StatusPageSkinEnum string

const (
	Skin2   StatusPageSkinEnum = "skin_2"
	SpSkin1 StatusPageSkinEnum = "sp_skin_1"
	SpSkin3 StatusPageSkinEnum = "sp_skin_3"
)

type StatusPageDateTimeFormatEnum string

const (
	DATETIME_FORMAT       StatusPageDateTimeFormatEnum = "DATETIME_FORMAT"
	SHORT_DATETIME_FORMAT StatusPageDateTimeFormatEnum = "SHORT_DATETIME_FORMAT"
	DotM_Y_H_i_e          StatusPageDateTimeFormatEnum = "d.m.Y H:i e"
)

type StatusPageDarkModeEnum string

const (
	Auto  StatusPageDarkModeEnum = "auto"
	Light StatusPageDarkModeEnum = "light"
	Dark  StatusPageDarkModeEnum = "dark"
)

type StatusPageCustomSMTP struct {
	EmailFrom string `json:"email_from"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	UseTLS    bool   `json:"use_tls"`
}

type StatusPageTheme struct {
	ColorBrand               string                `json:"color_brand"`
	ColorNotStarted          string                `json:"color_not_started"`
	ColorDegradedPerformance string                `json:"color_degraded_performance"`
	ColorMajorOutage         string                `json:"color_major_outage"`
	ColorOperational         string                `json:"color_operational"`
	ColorPartialOutage       string                `json:"color_partial_outage"`
	ColorUnderMaintenance    string                `json:"color_under_maintenance"`
	Theme                    StatusPageThemeEnum   `json:"theme"`
	BgColor                  StatusPageBgColorEnum `json:"bg_color"`
}

type StatusPage struct {
	UUID                   string                       `json:"uuid" tab:"UUID"`
	AbsoluteURL            string                       `json:"absolute_url" tab:"URL"`
	CreatedAt              *time.Time                   `json:"created_at,omitempty" tab:"CreatedAt"`
	Slug                   string                       `json:"slug" tab:"Domain"`
	ID                     int                          `json:"id"`
	UpdatedAt              *time.Time                   `json:"updated_at,omitempty"`
	Name                   string                       `json:"name"`
	Description            string                       `json:"description"`
	Impact                 ImpactType                   `json:"impact"`
	Team                   int                          `json:"team"`
	TeamSlug               string                       `json:"team_slug"`
	Theme                  StatusPageTheme              `json:"theme"`
	LogoLight              *Media                       `json:"logo_light,omitempty"`
	LogoDark               *Media                       `json:"logo_dark,omitempty"`
	Icon                   *Media                       `json:"icon,omitempty"`
	DateTimeFormat         StatusPageDateTimeFormatEnum `json:"datetime_format"`
	TimeZone               string                       `json:"timezone"`
	HTMLInHeader           *string                      `json:"html_in_header,omitempty"`
	HTMLInBody             *string                      `json:"html_in_body,omitempty"`
	ShowLastDays           int                          `json:"show_last_days"`
	MaxUptimeDays          int                          `json:"max_uptime_days"`
	MaintenanceStartText   string                       `json:"maintenance_start_text"`
	MaintenanceEndText     string                       `json:"maintenance_end_text"`
	ShowAffectedComponents bool                         `json:"show_affected_components"`
	Lang                   LanguageEnum                 `json:"lang"`
	Skin                   StatusPageSkinEnum           `json:"skin"`
	CustomDomain           string                       `json:"custom_domain"`
	CustomDomainVerifiedAt *time.Time                   `json:"custom_domain_verified_at,omitempty"`
	CustomSMTP             *StatusPageCustomSMTP        `json:"custom_smtp"`
	SSLReady               bool                         `json:"ssl_ready"`
	ApprovedAt             *time.Time                   `json:"approved_at,omitempty"`
	CustomLogoLink         string                       `json:"custom_logo_link"`
	CustomSupportLink      string                       `json:"custom_support_link"`
	CustomDocsLink         string                       `json:"custom_docs_link"`
	DarkMode               StatusPageDarkModeEnum       `json:"dark_mode"`
	PaymentsPaid           bool                         `json:"payments_paid"`
	PaymentsNextAt         *time.Time                   `json:"payments_next_at"`
	PaymentsDailyAmount    int                          `json:"payments_daily_amount"`
	HidePoweredBy          bool                         `json:"hide_powered_by"`
}

type ProtoStatusPage struct {
	CreateComponents []struct {
		Name   string     `json:"name"`
		Impact ImpactType `json:"impact"`
	} `json:"create_components"`
	Name     string          `json:"name"`
	TeamSlug *string         `json:"team_slug,omitempty"`
	Theme    StatusPageTheme `json:"theme"`
	Slug     string          `json:"slug"`
	Lang     LanguageEnum    `json:"lang"`
}

func (c *Client) GetPaginatedStatusPages(payload PaginatedRequest) (*Paginated[StatusPage], error) {
	queryParams := ConvertToQueryParams(payload)

	resp, err := c.Get("/api/pages/", queryParams)

	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve pages: status " + resp.Status)
	}

	var result Paginated[StatusPage]
	err = parseResponseBody(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &result, nil
}

func (c *Client) GetStatusPageBySlug(slug string) (*StatusPage, error) {
	url := fmt.Sprintf("/api/pages/%s/", slug)

	resp, err := c.Get(url, nil)
	if err != nil {
		return nil, errors.New("failed to perform GET request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve page: status " + resp.Status)
	}

	var statusPage StatusPage
	if err := parseResponseBody(resp, &statusPage); err != nil {
		return nil, errors.New("failed to parse response body: " + err.Error())
	}

	return &statusPage, nil
}

func (c *Client) CreateStatusPage(statusPage *ProtoStatusPage) (*StatusPage, error) {
	resp, err := c.Post("/api/pages/", statusPage)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create page: status " + resp.Status)
	}

	var newStatusPage StatusPage
	err = parseResponseBody(resp, &newStatusPage)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	return &newStatusPage, nil
}

func (c *Client) UpdateStatusPage(slug string, partial *StatusPage) (*StatusPage, error) {
	url := fmt.Sprintf("/api/pages/%s/", slug)

	resp, err := c.Patch(url, partial)
	if err != nil {
		return nil, errors.New("failed to perform PATCH request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to update page: status " + resp.Status)
	}

	var updatedStatusPage StatusPage
	if err := parseResponseBody(resp, &updatedStatusPage); err != nil {
		return nil, errors.New("failed to parse response body: " + err.Error())
	}

	return &updatedStatusPage, nil
}
