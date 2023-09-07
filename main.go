package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotState struct {
	Knowledge map[string]string
}

const knowledgeFile = "bot_knowledge.txt"

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	botState := BotState{
		Knowledge: make(map[string]string),
	}

	loadKnowledge(&botState)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Menerapkan logika belajar bot di sini
			response := HandleMessage(&botState, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)

			// Simpan pesan pengguna ke dalam berkas pengetahuan hanya jika belum ada
			saveMessageToKnowledgeIfNotExists(update.Message.Text, &botState)
		}
	}

	saveKnowledge(&botState)
}
func HandleMessage(botState *BotState, message string) string {
	if existingResponse, found := botState.Knowledge[message]; found {
		return existingResponse
	} else {
		// Jika pesan belum ada dalam pengetahuan, tambahkan ke berkas pengetahuan
		botState.Knowledge[message] = "Terima kasih, saya telah mempelajari pesan ini."
		return "Terima kasih atas pesan Anda."
	}

}

func loadKnowledge(botState *BotState) {
	file, err := os.Open(knowledgeFile)
	if err != nil {
		log.Printf("Error reading knowledge file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			botState.Knowledge[parts[0]] = parts[1]
		}
	}
}

func saveKnowledge(botState *BotState) {
	file, err := os.Create(knowledgeFile)
	if err != nil {
		log.Printf("Error creating knowledge file: %v", err)
		return
	}
	defer file.Close()

	for key, value := range botState.Knowledge {
		fmt.Fprintf(file, "%s:%s\n", key, value)
	}
}

func saveMessageToKnowledgeIfNotExists(message string, botState *BotState) {
	// Periksa apakah pesan pengguna sudah ada dalam pengetahuan
	if _, found := botState.Knowledge[message]; !found {
		// Simpan pesan pengguna ke berkas pengetahuan hanya jika belum ada
		saveMessageToKnowledge(message, botState)
	}
}

func saveMessageToKnowledge(message string, botState *BotState) {
	file, err := os.OpenFile(knowledgeFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Printf("Error opening knowledge file for appending: %v", err)
		return
	}
	defer file.Close()

	// Menambahkan pesan pengguna ke berkas pengetahuan
	fmt.Fprintf(file, "%s\n", message)

	// Juga perbarui pengetahuan botState
	botState.Knowledge[message] = "Terima kasih, saya telah mempelajari pesan ini."
}
