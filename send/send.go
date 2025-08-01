package send

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func SendMessageToWABA(to, msg string) error {
	if to == "" || msg == "" {
		return fmt.Errorf("destinatário e mensagem são obrigatórios")
	}

	phoneNumberID := os.Getenv("WABA_PHONE_NUMBER_ID")
	token := os.Getenv("WABA_TOKEN")
	if phoneNumberID == "" || token == "" {
		return fmt.Errorf("variáveis de ambiente WABA_PHONE_NUMBER_ID e WABA_TOKEN são obrigatórias")
	}

	url := "https://graph.facebook.com/v22.0/" + phoneNumberID + "/messages"

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": msg,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Enviando mensagem para %s: %s", to, msg)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		resBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro WABA: %s", string(resBody))
	}

	return nil
}

// UpdateTokens atualiza as variáveis de ambiente WABA_TOKEN e opcionalmente WABA_PHONE_NUMBER_ID
func UpdateTokens(token, phoneNumberID string) error {
	if token == "" {
		return fmt.Errorf("token não pode estar vazio")
	}

	// Atualizar o token
	os.Setenv("WABA_TOKEN", token)
	log.Printf("Token WABA atualizado")

	// Atualizar Phone Number ID se fornecido
	if phoneNumberID != "" {
		os.Setenv("WABA_PHONE_NUMBER_ID", phoneNumberID)
		log.Printf("Phone Number ID atualizado: %s", phoneNumberID)
	}

	return nil
}
