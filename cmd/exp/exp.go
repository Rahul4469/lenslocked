package main

import (
	"fmt"

	"github.com/Rahul4469/lenslocked/models"
)

func main() {
	// // Load environment variables from .env file
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// // Get SMTP configuration from environment variables
	// host := os.Getenv("SMTP_HOST")
	// portStr := os.Getenv("SMTP_PORT")
	// port, err := strconv.Atoi(portStr)
	// if err != nil {
	// 	log.Fatalf("Invalid SMTP_PORT: %v", err)
	// }
	// username := os.Getenv("SMTP_USERNAME")
	// password := os.Getenv("SMTP_PASSWORD")

	// // Validate required environment variables
	// if host == "" || portStr == "" || username == "" || password == "" {
	// 	log.Fatal("Missing required SMTP environment variables (SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD)")
	// }

	// es, err := models.NewEmailService(models.SMTPConfig{
	// 	Host:     host,
	// 	Port:     port,
	// 	Username: username,
	// 	Password: password,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// defer es.Close() // Close the client when done

	// err = es.ForgotPassword("rahul@gmail.com", "https://lenslocked.com/reset-pw?token=abs123")
	// if err != nil {
	// 	log.Fatalf("Failed to send forgot password email: %v", err)
	// }

	// log.Println("Forgot password email sent successfully!")

	gs := models.GalleryService{}
	fmt.Println(gs.Images(1))

}
