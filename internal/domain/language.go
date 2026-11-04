package domain

import "errors"

// NormalizeReportLanguage keeps API defaults compatible with existing clients.
func NormalizeReportLanguage(language string) (string, error) {
	if language == "" {
		return "id", nil
	}
	if language != "id" && language != "en" {
		return "", errors.New("report_language harus id atau en")
	}
	return language, nil
}
