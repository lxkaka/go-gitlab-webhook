package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

//GitlabRepository represents repository information from the webhook
type GitlabRepository struct {
	Name        string
	URL         string
	Description string
	Home        string
}

//Commit represents commit information from the webhook
type Commit struct {
	ID        string
	Message   string
	Timestamp string
	URL       string
	Author    Author
}

//Author represents author information from the webhook
type Author struct {
	Name  string
	Email string
}

//Webhook represents push information from the webhook
type PushEvent struct {
	Before            string
	After             string
	Ref               string
	Username          string
	UserID            int
	ProjectID         int
	Repository        GitlabRepository
	Commits           []Commit
	TotalCommitsCount int
}

// assignee information from merge request webhook
type GitlabUser struct {
	Name     string
	UserName string
}

type ObjectAttributes struct {
	LastCommit Commit `json:"last_commit"`
	Url        string
	Assignee   GitlabUser
	State      string
}

// merge request information from the webhook
type MergeRequestEvent struct {
	User             GitlabUser
	Repository       GitlabRepository
	ObjectAttributes ObjectAttributes `json:"object_attributes"`
}

//ConfigRepository represents a repository from the config file
type ConfigRepository struct {
	Name     string
	Commands []string
}

//Config represents the config file
type Config struct {
	Logfile     string
	Address     string
	Port        int64
	HookAddress string
}

// wewrok message content
type Content struct {
	Content string `json:"content"`
}

// wework robot message
type Message struct {
	MsgType string  `json:"msgtype"`
	Text    Content `json:"text"`
}

func PanicIf(err error, what ...string) {
	if err != nil {
		if len(what) == 0 {
			panic(err)
		}

		panic(errors.New(err.Error() + what[0]))
	}
}

var config Config
var configFile string

func main() {
	args := os.Args

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP)

	go func() {
		<-sigc
		var err error
		config, err = loadConfig(configFile)
		if err != nil {
			log.Fatalf("Failed to read config: %s", err)
		}
		log.Println("config reloaded")
	}()

	//if we have a "real" argument we take this as conf path to the config file
	if len(args) > 1 {
		configFile = args[1]
	} else {
		configFile = "config.json"
	}

	//load config
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	//open log file
	//writer, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	//if err != nil {
	//	log.Printf("Failed to open log file: %s", err)
	//	os.Exit(1)
	//}
	//
	////close logfile on exit
	//defer func() {
	//	writer.Close()
	//}()
	//
	////setting logging output
	//log.SetOutput(writer)

	//setting handler
	http.HandleFunc("/webhook", hookHandler)

	address := config.Address + ":" + strconv.FormatInt(config.Port, 10)

	log.Println("Listening on " + address)

	//starting server
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Println(err)
	}
}

func loadConfig(configFile string) (Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	count := 0

	count, err = file.Read(buffer)
	if err != nil {
		return Config{}, err
	}

	err = json.Unmarshal(buffer[:count], &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	var hook MergeRequestEvent

	//read request body
	var data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request: %s", err)
		return
	}
	//unmarshal request body
	err = json.Unmarshal(data, &hook)
	if err != nil {
		log.Printf("Failed to parse request: %s", err)
		return
	}
	log.Printf("request data: %s", data)
	log.Printf("hook info: %+v", hook)
	//assignee := hook.ObjectAttributes.Assignee
	//assigneeName := ""
	//if assignee == (GitlabUser{}) {
	//	assigneeName = "缺少 assignee"
	//} else {
	//	assigneeName = assignee.Name
	//}
	//content := fmt.Sprintf(
	//	"%s 有新的 PR(%s):\n%s\n%s\nAuthor: %s\nAssignee: %s",
	//	hook.Repository.Name,
	//	hook.ObjectAttributes.State,
	//	hook.ObjectAttributes.Url,
	//	hook.ObjectAttributes.LastCommit.Message,
	//	hook.User.Name,
	//	assigneeName)
	//message := Message{
	//	MsgType: "text",
	//	Text: Content{
	//		Content: content,
	//	},
	//}
	//sendMessageToWework(message)
	res, err := json.Marshal(hook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func sendMessageToWework(message Message) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	_, err = http.Post(
		config.HookAddress,
		"application/json",
		bytes.NewBuffer(msgBytes))

	if err != nil {
		log.Printf("Failed to send message: %s", err)
	}

}
