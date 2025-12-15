package main

import (
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		fmt.Printf("Error generating VAPID keys: %v\n", err)
		return
	}

	fmt.Println("# VAPID Keys Generated")
	fmt.Println("# Add these to your .env file or environment variables")
	fmt.Println()
	fmt.Printf("VAPID_PUBLIC_KEY=%s\n", publicKey)
	fmt.Printf("VAPID_PRIVATE_KEY=%s\n", privateKey)
	fmt.Println("# Note: Do NOT include 'mailto:' prefix - webpush-go adds it automatically")
	fmt.Println("VAPID_SUBJECT=admin@bcc.media")
	fmt.Println()
	fmt.Println("# For frontend (Nuxt):")
	fmt.Printf("NUXT_PUBLIC_VAPID_PUBLIC_KEY=%s\n", publicKey)
}
