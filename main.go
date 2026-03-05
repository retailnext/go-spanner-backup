package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/retailnext/spanner-backup/pkg/spannerbackup"
)

func main() {
	app := &cli.Command{
		Name:  "spanner-backup",
		Usage: "Backup spanner",
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

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

type PubSubMessage struct {
	Message      spannerbackup.PubSubMessage `json:"message"`
	Subscription string                      `json:"subscription"`
}

// sanitizeLog removes control characters from a string to
// prevent log injection attacks.
func sanitizeLog(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 0 && r < 32) || r == 127 {
			return -1
		}
		return r
	}, s)
}

func ServiceMain(ctx context.Context, cmd *cli.Command) error {
	http.HandleFunc("/", BackupSpanner)
	// Determine port for HTTP service.
	port := cmd.String("port")
	// Start HTTP server.
	log.Printf("Listening on port %s", sanitizeLog(port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("ListenAndServe: %s", sanitizeLog(err.Error()))
	}
	return nil
}

// BackupSpanner receives a Pub/Sub push message and process it.
func BackupSpanner(w http.ResponseWriter, r *http.Request) {
	var m PubSubMessage
	log.Printf("Going to sleep")
	time.Sleep(30 * time.Second)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// byte slice unmarshalling handles base64 decoding.
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %s", sanitizeLog(err.Error()))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err = spannerbackup.CreateBackupByPubSub(context.Background(), m.Message); err != nil {
		log.Printf("Spanner backup failed: %s", sanitizeLog(err.Error()))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func JobMain(ctx context.Context, cmd *cli.Command) error {
	params := spannerbackup.BackupParameters{
		BackupID: cmd.String("backupId"),
		Database: cmd.String("databaseId"),
		Expire:   cmd.Int("expire"),
	}
	messageData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	psMessage := spannerbackup.PubSubMessage{
		Data: messageData,
	}

	return spannerbackup.CreateBackupByPubSub(ctx, psMessage)
}
