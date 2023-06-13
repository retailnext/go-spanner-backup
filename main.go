package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/retailnext/spanner-backup/pkg/spannerbackup"
)

func main() {
	app := &cli.App{
		Name:     "spanner-backup",
		HelpName: "spanner-backup",
		Usage:    "Backup spanner",
		Commands: []*cli.Command{
			{
				Name:  "service",
				Usage: "Run backup as a service",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "port", Aliases: []string{"p"}, Value: "8080"},
				},
				Action: ServiceMain,
			},
			{
				Name:  "job",
				Usage: "Run backup as a job",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "backupId", Aliases: []string{"b"}},
					&cli.StringFlag{Name: "databaseId", Aliases: []string{"d"}, Required: true},
					&cli.IntFlag{Name: "expire", Aliases: []string{"e"}, Value: 1},
				},
				Action: JobMain,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type PubSubMessage struct {
	Message      spannerbackup.PubSubMessage `json:"message"`
	Subscription string                      `json:"subscription"`
}

func ServiceMain(cCtx *cli.Context) error {
	http.HandleFunc("/", BackupSpanner)
	// Determine port for HTTP service.
	port := cCtx.String("port")
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	return nil
}

// BackupSpanner receives a Pub/Sub push message and process it.
func BackupSpanner(w http.ResponseWriter, r *http.Request) {
	var m PubSubMessage
	log.Printf("Going to sleep")
	time.Sleep(30 * time.Second)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// byte slice unmarshalling handles base64 decoding.
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err = spannerbackup.CreateBackupByPubSub(context.Background(), m.Message); err != nil {
		log.Printf("Spanner backup failed: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func JobMain(cCtx *cli.Context) error {
	params := spannerbackup.BackupParameters{
		BackupID: cCtx.String("backupId"),
		Database: cCtx.String("databaseId"),
		Expire:   cCtx.Int("expire"),
	}
	messageData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	psMessage := spannerbackup.PubSubMessage{
		Data: messageData,
	}

	return spannerbackup.CreateBackupByPubSub(cCtx.Context, psMessage)
}
