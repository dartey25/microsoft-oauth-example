package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

//go:embed .env
var envBytes []byte

func main() {
	if err := os.WriteFile(".env", envBytes, 0o644); err != nil {
		panic("Error writing .env file")
	}

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	graph := NewGraphHelper()

	initializeGraph(graph)

	var (
		choice int64 = -1
		err    error
	)

	for {
		fmt.Println("Please choose one of the following options:")
		fmt.Println("0. Exit")
		fmt.Println("1. Display access token")
		fmt.Println("2. List users")
		fmt.Println("3. Read mail")
		fmt.Println("4. Send mail")

		_, err = fmt.Scanf("%d", &choice)
		if err != nil {
			choice = -1
		}

		ident := "    "

		switch choice {
		case 0:
			fmt.Println("Goodbye...")
		case 1:
			displayAccessToken(graph)
		case 2:
			users, err := getUsers(graph)
			if err != nil {
				log.Panic(err)
			}

			listUsers(users, "")
		case 3:
			listInbox(graph, ident)
		case 4:
			sendMail(graph, ident)
		default:
			fmt.Println("Invalid choice! Please try again.")
		}

		if choice == 0 {
			break
		}
	}
}

func initializeGraph(graphHelper *GraphHelper) {
	err := graphHelper.InitializeGraphForAppAuth()
	if err != nil {
		log.Panicf("Error initializing Graph for app auth: %v\n", err)
	}
}

func displayAccessToken(graphHelper *GraphHelper) {
	token, err := graphHelper.GetAppToken()
	if err != nil {
		log.Panicf("Error getting user token: %v\n", err)
	}

	fmt.Printf("App-only token: %s", *token)
	fmt.Println()
}

type Users struct {
	ID    string
	Name  string
	Email string
}

func getUsers(g *GraphHelper) ([]Users, error) {
	u, err := g.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("Error getting users: %v", err)
	}

	var users []Users
	for _, user := range u.GetValue() {
		id := *user.GetId()
		name := *user.GetDisplayName()
		email := *user.GetMail()
		if email == "" {
			email = "NO EMAIL"
		}

		users = append(users, Users{ID: id, Name: name, Email: email})
	}

	return users, nil
}

func listUsers(users []Users, ident string) {
	for idx := range users {
		fmt.Printf("%s%d. User: %s\n", ident, idx+1, users[idx].Name)
		fmt.Printf("%s  ID: %s\n", ident, users[idx].ID)
		fmt.Printf("%s  Email: %s\n", ident, users[idx].Email)
	}
}

func listInbox(graphHelper *GraphHelper, ident string) {
	choice := -1
	users, err := getUsers(graphHelper)
	if err != nil {
		log.Panic(err)
	}

	for {
		fmt.Println()
		fmt.Printf("%sPlease choose one of the following users:\n", ident)
		listUsers(users, ident)

		_, err = fmt.Scanf("%d", &choice)
		if err != nil {
			choice = -1
		}

		if choice > len(users) || choice < 1 {
			fmt.Printf("%sInvalid choice! Please try again.\n", ident)
			continue
		}

		fmt.Println()
		fmt.Println()

		messages, err := graphHelper.GetInbox(users[choice-1].ID)
		if err != nil {
			fmt.Printf("%sError getting user's inbox: %v\n", ident, err)
			break
		}

		// Load local time zone
		// Dates returned by Graph are in UTC, use this
		// to convert to local
		location, err := time.LoadLocation("Local")
		if err != nil {
			fmt.Printf("%sError getting local timezone: %v\n", ident, err)
			break
		}

		// Output each message's details
		for _, message := range messages.GetValue() {
			fmt.Printf("%sMessage: %s\n", ident, *message.GetSubject())
			fmt.Printf("%s  From: %s\n", ident, *message.GetFrom().GetEmailAddress().GetName())

			status := "Unknown"
			if *message.GetIsRead() {
				status = "Read"
			} else {
				status = "Unread"
			}
			fmt.Printf("%s  Status: %s\n", ident, status)
			fmt.Printf("%s  Received: %s\n", ident, (*message.GetReceivedDateTime()).In(location))
		}

		nextLink := messages.GetOdataNextLink()

		fmt.Println()
		fmt.Printf("%sMore messages available? %t\n", ident, nextLink != nil)
		fmt.Println()

		break
	}
}

func sendMail(graphHelper *GraphHelper, ident string) {
	choice := -1
	users, err := getUsers(graphHelper)
	if err != nil {
		log.Panic(err)
	}

	for {
		fmt.Printf("%sPlease choose one of the following users on whose behalf you want to send a mail:\n", ident)
		listUsers(users, ident)

		_, err = fmt.Scanf("%d", &choice)
		if err != nil {
			choice = -1
		}

		if choice > len(users) || choice < 1 {
			fmt.Printf("%sInvalid choice! Please try again.\n", ident)
			continue
		}

		fmt.Printf("%sPlease enter the subject of the mail:\n", ident)
		var subject string
		_, _ = fmt.Scanf("%s", &subject)

		fmt.Printf("%sPlease enter the body of the mail:\n", ident)
		var body string
		_, _ = fmt.Scanf("%s", &body)

		fmt.Printf("%sPlease enter the recipient's email address:\n", ident)
		var recipient string
		_, _ = fmt.Scanf("%s", &recipient)

		fmt.Println()
		fmt.Println()

		err := graphHelper.SendMail(users[choice-1].ID, &subject, &body, &recipient)
		if err != nil {
			fmt.Printf("%sError sending mail: %v\n", ident, err)
			break
		}

		fmt.Printf("%sMail successfully sent!\n", ident)
		break
	}
}
