package api

import (
	"fmt"
	"net/url"
	"reflect"
)

func (c *Client) stringifyQueryParams(endpoint string, queryParams QueryParams) (*url.URL, error) {
	fullURL, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, err
	}

	q := fullURL.Query()

	for key, value := range queryParams {
		switch v := value.(type) {
		case string:
			q.Add(key, v)
		case []string:
			for _, val := range v {
				q.Add(key, val)
			}
		default:
			switch reflect.TypeOf(value).Kind() {
			case reflect.Slice:
				valSlice := reflect.ValueOf(value)
				for i := 0; i < valSlice.Len(); i++ {
					q.Add(key, fmt.Sprintf("%v", valSlice.Index(i).Interface()))
				}
			default:
				q.Add(key, fmt.Sprintf("%v", value))
			}
		}
	}

	fullURL.RawQuery = q.Encode()

	return fullURL, nil
}

func BuildAffectedComponents(affected []string, components []Component) ([]AffectedComponent, error) {
	var affectedComponents []AffectedComponent

	for _, ci := range affected {
		parsedCI, err := ParseComponentImpact(ci)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s': %w", ci, err)
		}

		var foundComponent *Component
		for _, cmp := range components {
			if cmp.Name == parsedCI.Component {
				foundComponent = &cmp
				break
			}
		}

		if foundComponent == nil {
			return nil, fmt.Errorf("component '%s' not found", parsedCI.Component)
		}

		ac := NewAffectedComponent(*foundComponent.ID, parsedCI.Impact)
		affectedComponents = append(affectedComponents, *ac)
	}

	return affectedComponents, nil
}
