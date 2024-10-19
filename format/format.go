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

func Marshal(v interface{}, commentsMap *map[string]string) (string, error) {
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

	// Обработка полей структуры
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		// Получаем тег из структуры
		tag := field.Tag.Get("format")

		if commentsMap != nil {
			// Получаем комментарий из мапы по тегу
			comment, ok := (*commentsMap)[tag]

			if ok && comment != "" {
				commentLines := strings.Split(comment, "\n")
				for _, commentLine := range commentLines {
					if _, err := writer.WriteString(fmt.Sprintf("# %s\n", commentLine)); err != nil {
						return "", err
					}
				}
			}
		}

		if tag != "" {
			if _, err := writer.WriteString(fmt.Sprintf("[%s]\n", tag)); err != nil {
				return "", err
			}
		}

		// Проверяем, является ли поле массивом строк
		switch fieldValue.Kind() {
		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() == reflect.String {
				for j := 0; j < fieldValue.Len(); j++ {
					if _, err := writer.WriteString(fmt.Sprintf("%s\n", fieldValue.Index(j).String())); err != nil {
						return "", err
					}
				}
			}
		case reflect.String:
			if _, err := writer.WriteString(fmt.Sprintf("%s\n", fieldValue.String())); err != nil {
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

		// Добавляем разделитель между полями
		if _, err := writer.WriteString("\n"); err != nil {
			return "", err
		}
	}

	if err := writer.Flush(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func Unmarshal(data string, v interface{}) error {
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
			// Игнорируем пустые строки и комментарии
			continue
		}

		// Если строка начинается с '[', то это заголовок поля
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Сохраняем значение предыдущего поля, если есть
			if inMultiline && currentField != "" {
				setFieldValue(val, currentField, currentValue.String())
				currentValue.Reset()
			}

			currentField = strings.Trim(line, "[]")
			inMultiline = true
			continue
		}

		// Если продолжается многострочное значение, добавляем строку
		if inMultiline {
			if currentValue.Len() > 0 {
				currentValue.WriteString("\n")
			}
			currentValue.WriteString(line)
		}
	}

	// Устанавливаем последнее поле, если оно было многострочным
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
			// Преобразуем строковые значения в bool
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "1", "yes", "true", "on":
				field.SetBool(true)
			case "0", "no", "false", "off":
				field.SetBool(false)
			}
		}
	}
}
