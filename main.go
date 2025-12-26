package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"wolfscream/database"
	"wolfscream/discord"
	"wolfscream/routes"
	"wolfscream/websocket"
	websocket_handlers "wolfscream/websocket/handlers"
)

func main() {
	websocket.InitHub()
	websocket_handlers.InitHandlers()

	r := routes.NewRouter()

	r.Get("/ws", websocket.HandleWebSocket)

	srv := &http.Server {
		Addr: ":8080",
		Handler: r,
	}
	

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()
	
	<-stop
	log.Println("Shutting down server...")


	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	discord.DiscordBot.Close()

	if err := database.DB.Close(); err != nil {
		log.Printf("Error closing DB: %v", err)
	}

	log.Println("Server exited properly")
	
}
