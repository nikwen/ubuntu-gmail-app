package main

import (
        "gopkg.in/qml.v1"
        "log"
        "net/http"
        "math"
        "strings"
        "time"

        gmail "code.google.com/p/google-api-go-client/gmail/v1"
        "code.google.com/p/goauth2/oauth"
)

var (
        context qml.Context
        threadModel ThreadModel
)

func main() {
        err := qml.Run(run)
        if (err != nil) {
                log.Fatal(err)
        }
}

func run() error {
        engine := qml.NewEngine()
        component, err := engine.LoadFile("share/ubuntu-gmail-app/main.qml")
        if err != nil {
                return err
        }

        context := engine.Context()
        backend := GmailBackend{}
        context.SetVar("gmailBackend", &backend)
        context.SetVar("threadModel", &threadModel)

        win := component.CreateWindow(nil)

        win.Show()
        win.Wait()
        return nil
}

type GmailBackend struct {
        Handler      OAuthHandler
        GmailService *gmail.Service
}

func (backend *GmailBackend) InstantiateGmailService(ClientId string, ClientSecret string, Scope string, Token string) {
        //Clear old threadModel
        threadModel = ThreadModel{}

        backend.Handler.SetOAuthData(ClientId, ClientSecret, Scope, Token)
        httpClient := backend.Handler.CreateHttpClient()

        svc, err := gmail.New(httpClient)
        if err != nil {
                log.Fatalf("Unable to create Gmail service: %v", err)
        } else {
                backend.GmailService = svc
        }

        log.Print("New gmail")
}

func (backend *GmailBackend) ListThreadsAsynchronously() {
        go func() {
                req := backend.GmailService.Users.Threads.List("me")
                r, err := req.Do()
                if err != nil {
                        log.Fatalf("Unable to retrieve threads: %v", err)
                        return
                }

                numberOfThreadsToLoad := int(math.Min(20.0, float64(len(r.Threads))))

                queue := make(chan Thread, numberOfThreadsToLoad)

                for i := 0; i < numberOfThreadsToLoad; i++ {
                        go func(t *gmail.Thread) {
                                messages, err := backend.FetchMessage(t)
                                if err != nil {
                                        log.Fatalf("Unable to retrieve thread: %v", err)
                                        return
                                }

                                thread := Thread{
                                        MessageCount: len(messages),
                                }

                                for messageNo, m := range messages {
                                        for _, label := range m.LabelIds {
                                                if label == "UNREAD" {
                                                        thread.Unread = true
                                                        if thread.Snippet == "" {
                                                                thread.Snippet = m.Snippet
                                                        }
                                                }
                                        }
                                        payload := *m.Payload
                                        for _, header := range payload.Headers {
                                                if header.Name == "From" {
                                                        if !sliceContains(thread.Senders, header.Value) {
                                                                thread.Senders = append(thread.Senders, header.Value)
                                                        }
                                                } else if header.Name == "Subject" && messageNo == 0 {
                                                        thread.Subject = header.Value
                                                } else if header.Name == "Date" && messageNo == thread.MessageCount - 1 {
                                                        thread.Date = header.Value
                                                }
                                        }
                                }

                                if (thread.Snippet == "") {
                                        thread.Snippet = messages[len(messages) - 1].Snippet
                                }

                                queue <- thread
                        }(r.Threads[i])
                }

                go func() { //TODO: Order!
                        for t := range queue {
                                threadModel.AppendThread(&t)
                        }
                        log.Print("Ready")
                }()
        }()
}

func sliceContains(slice []string, s string) bool {
        for _, a := range slice {
                if a == s {
                        return true
                }
        }
        return false
}

func (backend *GmailBackend) FetchMessage(thread *gmail.Thread) ([]*gmail.Message, error) {
        re, err := backend.GmailService.Users.Threads.Get("me", thread.Id).Do()
        if err != nil {
                return nil, err
        }

        return re.Messages, nil
}

type ThreadModel struct {
        Threads []Thread
        Len int
}

type Thread struct {
        Subject string
        Snippet string
        Senders []string
        MessageCount int
        Date string
        Unread bool
}

func (tm *ThreadModel) AppendThread(t *Thread) { //TODO: Order
        tm.Threads = append(tm.Threads, *t)
        tm.Len = len(tm.Threads)
        qml.Changed(tm, &tm.Len)
}

func (tm *ThreadModel) GetThread(index int) *Thread {
        return &tm.Threads[index]
}

func (t *Thread) GetSendersString() string {
        if t.MessageCount == 1 {
                return GetFullNameFromEmail(t.Senders[0])
        }

        var senders string
        if len := len(t.Senders); len <= 3 {
                for _, name := range t.Senders {
                        if (senders == "") {
                                senders = GetFirstNameFromEmail(name)
                        } else {
                                senders += ", " + GetFirstNameFromEmail(name)
                        }
                }
        } else {
                senders = GetFirstNameFromEmail(t.Senders[0]) + " … " + GetFirstNameFromEmail(t.Senders[len - 2]) + ", " + GetFirstNameFromEmail(t.Senders[len - 1])
        }
        return senders
}

func (t *Thread) GetDateString() string { //TODO: Localize time formats (and "Yesterday")
        layout := "Mon, 2 Jan 2006 15:04:05 -0700 (MST)"
        parsed, err := time.Parse(layout, t.Date)
        if err != nil {
                layout = "Mon, 2 Jan 2006 15:04:05 -0700"
                parsed, err = time.Parse(layout, t.Date)
                if err != nil {
                        return t.Date
                }
        }
        now := time.Now()
        if sameDay(now, parsed) {
                return parsed.Format("15:04")
        } else if dayBefore(now, parsed) {
                return "Yesterday"
        } else if weekBefore(now, parsed) {
                return parsed.Format("Monday")
        } else if sameYear(now, parsed) {
                return parsed.Format("2 Jan")
        } else {
                return parsed.Format("2 Jan 2006")
        }
}

func sameDay(first, second time.Time) bool {
        return (first.Format("2 Jan 2006") == second.Format("2 Jan 2006"))
}

func dayBefore(first, second time.Time) bool {
        return (first.Add(- 24 * time.Hour).Format("2 Jan 2006") == second.Format("2 Jan 2006"))
}

func weekBefore(first, second time.Time) bool {
        year, month, day := first.Add(- 24 * 7 * time.Hour).Date()
        weekAgo := time.Date(year, month, day, 0, 0, 0, 0, first.Location())
        return weekAgo.Before(second)
}

func sameYear(first, second time.Time) bool {
        return (first.Year() == second.Year())
}

func GetFullNameFromEmail(email string) string {
        bracketClosePos := strings.LastIndex(email, "<")
        if bracketClosePos > 0 && bracketClosePos < len(email) {
                return strings.TrimSpace(email[0:bracketClosePos])
        } else {
                atPos := strings.Index(email, "@")
                if atPos > 0 && atPos < len(email) {
                        return strings.TrimSpace(email[0:atPos])
                } else {
                        return email
                }
        }
}

func GetFirstNameFromEmail(email string) string { //TODO: Improve (e.g. display "me" and drop full mail adress in first case, full name if just one sender)
        name := GetFullNameFromEmail(email)
        if strings.Contains(name, ", ") { //Assume: last name, first name
                splitted := strings.Split(name, ", ")
                return splitted[1]
        } else { //Assume: First name, last name
                splitted := strings.Split(name, " ")
                return splitted[0]
        }
}

type OAuthHandler struct {
        ClientId     string
        ClientSecret string
        Scope        string
        Token        string
}

func (handler *OAuthHandler) SetOAuthData(ClientId string, ClientSecret string, Scope string, Token string) {
        handler.ClientId     = ClientId
        handler.ClientSecret = ClientSecret
        handler.Scope        = Scope
        handler.Token        = Token
}

func (handler *OAuthHandler) CreateHttpClient() (*http.Client) {
        var config = &oauth.Config{
                ClientId:     handler.ClientId, // from https://code.google.com/apis/console/
                ClientSecret: handler.ClientSecret, // from https://code.google.com/apis/console/
                Scope:        handler.Scope,
                AuthURL:      "https://accounts.google.com/o/oauth2/auth",
                TokenURL:     "https://accounts.google.com/o/oauth2/token",
        }

        var token = &oauth.Token{
                AccessToken: handler.Token,
        }

        transport := &oauth.Transport{
                Token:     token,
                Config:    config,
                Transport: http.DefaultTransport,
        }

        httpClient := transport.Client()

        log.Print("Return httpClient")

        return httpClient
}
