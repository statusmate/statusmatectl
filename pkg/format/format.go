package format

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func Marshal(v any, commentsMap *map[string]string) (string, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return "", errors.New("expected a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return "", errors.New("expected a pointer to a struct")
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		tag := field.Tag.Get("format")

		if commentsMap != nil {
			comment, ok := (*commentsMap)[tag]

			if ok && comment != "" {
				commentLines := strings.Split(comment, "\n")
				for _, commentLine := range commentLines {
					if _, err := fmt.Fprintf(writer, "# %s\n", commentLine); err != nil {
						return "", err
					}
				}
			}
		}

		if tag != "" {
			if _, err := fmt.Fprintf(writer, "[%s]\n", tag); err != nil {
				return "", err
			}
		}

		switch fieldValue.Kind() {
		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() == reflect.String {
				for j := 0; j < fieldValue.Len(); j++ {
					if _, err := fmt.Fprintf(writer, "%s\n", fieldValue.Index(j).String()); err != nil {
						return "", err
					}
				}
			}
		case reflect.String:
			if _, err := fmt.Fprintf(writer, "%s\n", fieldValue.String()); err != nil {
				return "", err
			}

		case reflect.Bool:
			if fieldValue.Bool() {
				if _, err := writer.WriteString("yes\n"); err != nil {
					return "", err
				}
			} else {
				if _, err := writer.WriteString("no\n"); err != nil {
					return "", err
				}
			}
		}

		if _, err := writer.WriteString("\n"); err != nil {
			return "", err
		}
	}

	if err := writer.Flush(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func Unmarshal(data string, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("expected a non-nil pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("expected a pointer to a struct")
	}

	reader := bufio.NewReader(strings.NewReader(data))
	currentField := ""
	var currentValue strings.Builder
	inMultiline := false

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if inMultiline && currentField != "" {
				setFieldValue(val, currentField, currentValue.String())
				currentValue.Reset()
			}

			currentField = strings.Trim(line, "[]")
			inMultiline = true
			continue
		}

		if inMultiline {
			if currentValue.Len() > 0 {
				currentValue.WriteString("\n")
			}
			currentValue.WriteString(line)
		}
	}

	if inMultiline && currentField != "" {
		setFieldValue(val, currentField, currentValue.String())
	}

	return nil
}

func setFieldValue(val reflect.Value, fieldName, value string) {
	field := val.FieldByNameFunc(func(name string) bool {
		fieldTag, _ := val.Type().FieldByName(name)
		return fieldTag.Tag.Get("format") == fieldName
	})

	if field.IsValid() && field.CanSet() {
		switch field.Kind() {
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				lines := strings.Split(value, "\n")
				for _, line := range lines {
					field.Set(reflect.Append(field, reflect.ValueOf(line)))
				}
			}
		case reflect.String:
			field.SetString(value)

		case reflect.Bool:
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "1", "yes", "true", "on":
				field.SetBool(true)
			case "0", "no", "false", "off":
				field.SetBool(false)
			}
		}
	}
}
