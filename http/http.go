package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brunoOchoa/api-send.git/send"
)

func StartServer() {
	// Log de inicialização
	log.Println("[SERVER] Iniciando servidor na porta 8080...")

	// Configurar rotas
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/responder", handleResponder)
	http.HandleFunc("/update-token", handleUpdateToken)

	// Criar servidor HTTP
	server := &http.Server{
		Addr: ":8080",
	}

	// Goroutine para capturar sinais de interrupção e fazer graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		log.Println("[SERVER] Recebido sinal de interrupção, iniciando graceful shutdown...")

		// Criar contexto com timeout de 30 segundos para o shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fazer graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[SERVER] Erro durante graceful shutdown: %v", err)
			os.Exit(1)
		}

		log.Println("[SERVER] Servidor finalizado com sucesso!")
		os.Exit(0)
	}()

	// Mostrar link clicável no terminal
	log.Println("[SERVER] Servidor rodando em: http://localhost:8080")

	// Iniciar servidor
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[SERVER] Erro ao iniciar servidor: %v", err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func handleResponder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	to := r.FormValue("to")
	message := r.FormValue("message")

	// fallback se vier via JSON
	isJSON := r.Header.Get("Content-Type") == "application/json"
	if to == "" && isJSON {
		var req struct {
			To      string `json:"to"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		to = req.To
		message = req.Message
	}

	log.Printf("Recebido pedido para %s: %s", to, message)
	if err := send.SendMessageToWABA(to, message); err != nil {
		log.Printf("Erro ao enviar mensagem: %v", err)
		http.Error(w, "erro ao enviar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Sempre retorna JSON válido se a chamada for API
	if isJSON || r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func handleUpdateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token         string `json:"token"`
		PhoneNumberID string `json:"phoneNumberId,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token é obrigatório", http.StatusBadRequest)
		return
	}

	// Atualizar variáveis de ambiente
	if err := send.UpdateTokens(req.Token, req.PhoneNumberID); err != nil {
		log.Printf("Erro ao atualizar tokens: %v", err)
		http.Error(w, "Erro ao atualizar tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[TOKEN] Token atualizado com sucesso!")
	w.Write([]byte("Token atualizado com sucesso!"))
}
