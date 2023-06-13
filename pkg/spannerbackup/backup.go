package spannerbackup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	adminpb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	pbt "github.com/golang/protobuf/ptypes/timestamp"
)

const (
	defaultExpireDays = 14
)

type PubSubMessage struct {
	Data []byte `json:"data,omitempty"`
}

type BackupParameters struct {
	BackupID string `json:"backupId"`
	Database string `json:"database"`
	Expire   int    `json:"expire"`
	Parent   string
}

func CreateBackupByPubSub(ctx context.Context, m PubSubMessage) error {
	var params BackupParameters
	err := json.Unmarshal(m.Data, &params)
	if err != nil {
		return fmt.Errorf("failed to parse pubsub message %s: %v", string(m.Data), err)
	}

	// parse database string
	matches := regexp.MustCompile("^(projects/[^/]+/instances/[^/]+)/databases/([^/]+)$").FindStringSubmatch(params.Database)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("createBackup: invalid database id %q", params.Database)
	}
	params.Parent = matches[1]

	// set the default values
	if params.Expire == 0 {
		params.Expire = defaultExpireDays
	}
	if params.BackupID == "" {
		params.BackupID = fmt.Sprintf("backup-%s-%d", matches[2], time.Now().Unix())
	}
	return CreateBackup(ctx, params, time.Now(), false) // waitTillFinish = false since it takes over 10 mins even for a table with 3 rows
}

// createBackup creates spanner backup
// source: https://github.com/GoogleCloudPlatform/golang-samples/blob/main/spanner/spanner_snippets/spanner/spanner_create_backup.go
func CreateBackup(ctx context.Context, params BackupParameters, versionTime time.Time, waitTillFinish bool) error {
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("createBackup.NewDatabaseAdminClient: %w", err)
	}
	defer adminClient.Close()

	expireTime := time.Now().AddDate(0, 0, params.Expire)
	// Create a backup.
	req := adminpb.CreateBackupRequest{
		Parent:   params.Parent,
		BackupId: params.BackupID,
		Backup: &adminpb.Backup{
			Database:    params.Database,
			ExpireTime:  &pbt.Timestamp{Seconds: expireTime.Unix(), Nanos: int32(expireTime.Nanosecond())},
			VersionTime: &pbt.Timestamp{Seconds: versionTime.Unix(), Nanos: int32(versionTime.Nanosecond())},
		},
	}
	op, err := adminClient.CreateBackup(ctx, &req)
	if err != nil {
		return fmt.Errorf("createBackup.CreateBackup: %w", err)
	}

	if !waitTillFinish {
		return nil
	}

	// Wait for backup operation to complete.
	backup, err := op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("createBackup.Wait: %w", err)
	}

	// Get the name, create time, version time and backup size.
	backupCreateTime := time.Unix(backup.CreateTime.Seconds, int64(backup.CreateTime.Nanos))
	backupVersionTime := time.Unix(backup.VersionTime.Seconds, int64(backup.VersionTime.Nanos))
	log.Printf(
		"Backup %s of size %d bytes was created at %s with version time %s\n",
		backup.Name,
		backup.SizeBytes,
		backupCreateTime.Format(time.RFC3339),
		backupVersionTime.Format(time.RFC3339))
	return nil
}
