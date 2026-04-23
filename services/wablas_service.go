package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type wablasMessageResponse struct {
	Status  interface{} `json:"status"`
	Success interface{} `json:"success"`
	Message string      `json:"message"`
}

func IsWablasConfigured() bool {
	return strings.TrimSpace(os.Getenv("WABLAS_TOKEN")) != "" &&
		strings.TrimSpace(os.Getenv("WABLAS_SECRET_KEY")) != ""
}

func SendWhatsAppOTP(phone, otpCode string) error {
	token := strings.TrimSpace(os.Getenv("WABLAS_TOKEN"))
	secretKey := strings.TrimSpace(os.Getenv("WABLAS_SECRET_KEY"))
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("WABLAS_BASE_URL")), "/")

	if baseURL == "" {
		baseURL = "https://sby.wablas.com"
	}

	if token == "" && secretKey == "" {
		return errors.New("wablas belum dikonfigurasi")
	}
	if token == "" || secretKey == "" {
		return errors.New("WABLAS_TOKEN dan WABLAS_SECRET_KEY wajib diisi")
	}

	message := fmt.Sprintf("Kode OTP SaaS Cloud kamu: %s. Berlaku 7 menit. Jangan bagikan kode ini ke siapa pun.", otpCode)
	form := url.Values{}
	form.Set("phone", phone)
	form.Set("message", message)
	form.Set("secret", "true")
	form.Set("retry", "true")

	request, err := http.NewRequest(http.MethodPost, baseURL+"/api/send-message", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", token+"."+secretKey)

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("wablas gagal kirim OTP, status: %d, response: %s", response.StatusCode, string(body))
	}

	var parsed wablasMessageResponse
	if err := json.Unmarshal(body, &parsed); err == nil {
		status := parsed.Status
		if status == nil {
			status = parsed.Success
		}
		if !isWablasSuccess(status) {
			if parsed.Message == "" {
				parsed.Message = string(body)
			}
			return fmt.Errorf("wablas menolak kirim OTP: %s", parsed.Message)
		}

		fmt.Printf("[WABLAS] OTP request accepted for %s: %s\n", phone, string(body))
	}

	return nil
}

func isWablasSuccess(status interface{}) bool {
	switch value := status.(type) {
	case bool:
		return value
	case string:
		lowered := strings.ToLower(value)
		return lowered == "true" || lowered == "success"
	default:
		return status == nil
	}
}
