package event

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

var logger *slog.Logger

const FileReadyEventType = "FileReady"

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type Event struct {
	ID        string
	Type      string
	LockToken string
}

type FileReadyEvent struct {
	Event
	Manifest      map[string]string
	DeliverTarget string
}

func NewFileReadyEvent(id string, manifest map[string]string, target string) FileReadyEvent {
	return FileReadyEvent{
		Event: Event{
			ID:   id,
			Type: FileReadyEventType,
		},
		Manifest:      manifest,
		DeliverTarget: target,
	}
}

type Publisher interface {
	Publish(ctx context.Context, event FileReadyEvent) error
}

type MemoryPublisher struct {
	Dir              string
	FileReadyChannel chan FileReadyEvent
}

type AzurePublisher struct {
	Client *aznamespaces.SenderClient
}

func (mp *MemoryPublisher) Publish(_ context.Context, event FileReadyEvent) error {
	err := os.Mkdir(mp.Dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	filename := mp.Dir + "/" + event.ID
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// write event to file.
	encoder := json.NewEncoder(f)
	err = encoder.Encode(event)
	if err != nil {
		return err
	}

	mp.FileReadyChannel <- event
	return nil
}

func (ap *AzurePublisher) Publish(ctx context.Context, event FileReadyEvent) error {
	logger.Info("publishing file ready event")
	evt, err := messaging.NewCloudEvent("upload", "fileReady", event, nil)
	if err != nil {
		return err
	}

	_, err = ap.Client.SendEvent(ctx, &evt, nil)

	return err
}
