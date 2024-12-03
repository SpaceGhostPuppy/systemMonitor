package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

// FileInfo speichert Metadaten über überwachte Dateien
type FileInfo struct {
    Path         string
    Size         int64
    LastModified time.Time
    Permissions  os.FileMode
}

// Notification repräsentiert eine Änderungsbenachrichtigung
type Notification struct {
    EventType    string    `json:"event_type"`
    Path         string    `json:"path"`
    Size         int64     `json:"size"`
    LastModified time.Time `json:"last_modified"`
    Permissions  string    `json:"permissions"`
    Timestamp    time.Time `json:"timestamp"`
}

// NotificationService verwaltet das Senden von Benachrichtigungen
type NotificationService struct {
    WebhookURL string
    Client     *http.Client
}

// NewNotificationService erstellt einen neuen Benachrichtigungsdienst
func NewNotificationService(webhookURL string) *NotificationService {
    return &NotificationService{
        WebhookURL: webhookURL,
        Client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// SendNotification sendet eine Benachrichtigung an den Webhook
func (ns *NotificationService) SendNotification(notification Notification) error {
    jsonData, err := json.Marshal(notification)
    if err != nil {
        return fmt.Errorf("error marshaling notification: %v", err)
    }

    req, err := http.NewRequest("POST", ns.WebhookURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := ns.Client.Do(req)
    if err != nil {
        return fmt.Errorf("error sending notification: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
    }

    return nil
}

// Monitor überwacht bestimmte Verzeichnisse auf Änderungen
func Monitor(paths []string, interval time.Duration, notificationService *NotificationService) {
    fileMap := make(map[string]FileInfo)
    
    // Erster Scan der Dateien
    for _, path := range paths {
        err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            fileMap[path] = FileInfo{
                Path:         path,
                Size:         info.Size(),
                LastModified: info.ModTime(),
                Permissions:  info.Mode(),
            }
            return nil
        })
        if err != nil {
            log.Printf("Fehler beim Scannen des Pfads %s: %v\n", path, err)
        }
    }
    
    // Überwachungsschleife
    ticker := time.NewTicker(interval)
    for range ticker.C {
        for _, path := range paths {
            err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }
                
                // Prüfen, ob die Datei neu ist
                oldInfo, exists := fileMap[path]
                if !exists {
                    notification := Notification{
                        EventType:    "file_created",
                        Path:         path,
                        Size:         info.Size(),
                        LastModified: info.ModTime(),
                        Permissions:  info.Mode().String(),
                        Timestamp:    time.Now(),
                    }
                    if err := notificationService.SendNotification(notification); err != nil {
                        log.Printf("Failed to send notification: %v\n", err)
                    }
                    log.Printf("Neue Datei erkannt: %s\n", path)
                } else {
                    // Auf Änderungen prüfen
                    if info.Size() != oldInfo.Size {
                        notification := Notification{
                            EventType:    "size_changed",
                            Path:         path,
                            Size:         info.Size(),
                            LastModified: info.ModTime(),
                            Permissions:  info.Mode().String(),
                            Timestamp:    time.Now(),
                        }
                        if err := notificationService.SendNotification(notification); err != nil {
                            log.Printf("Failed to send notification: %v\n", err)
                        }
                        log.Printf("Dateigröße geändert: %s\n", path)
                    }
                    if info.ModTime() != oldInfo.LastModified {
                        notification := Notification{
                            EventType:    "file_modified",
                            Path:         path,
                            Size:         info.Size(),
                            LastModified: info.ModTime(),
                            Permissions:  info.Mode().String(),
                            Timestamp:    time.Now(),
                        }
                        if err := notificationService.SendNotification(notification); err != nil {
                            log.Printf("Failed to send notification: %v\n", err)
                        }
                        log.Printf("Datei modifiziert: %s\n", path)
                    }
                    if info.Mode() != oldInfo.Permissions {
                        notification := Notification{
                            EventType:    "permissions_changed",
                            Path:         path,
                            Size:         info.Size(),
                            LastModified: info.ModTime(),
                            Permissions:  info.Mode().String(),
                            Timestamp:    time.Now(),
                        }
                        if err := notificationService.SendNotification(notification); err != nil {
                            log.Printf("Failed to send notification: %v\n", err)
                        }
                        log.Printf("Dateiberechtigungen geändert: %s\n", path)
                    }
                }
                
                // Dateiinformationen aktualisieren
                fileMap[path] = FileInfo{
                    Path:         path,
                    Size:         info.Size(),
                    LastModified: info.ModTime(),
                    Permissions:  info.Mode(),
                }
                return nil
            })
            if err != nil {
                log.Printf("Fehler bei der Überwachung des Pfads %s: %v\n", path, err)
            }
        }
    }
}

func main() {
    // Zu überwachende Verzeichnisse
    paths := []string{
        ".", // Aktuelles Verzeichnis
        // Hier können weitere Pfade hinzugefügt werden
    }
    
    // Webhook URL für Benachrichtigungen
    webhookURL := "http://localhost:8080/webhook" // Ändern Sie dies zu Ihrer Webhook-URL
    notificationService := NewNotificationService(webhookURL)
    
    fmt.Println("Starte Dateisystem-Monitor mit Netzwerk-Benachrichtigungen...")
    Monitor(paths, 5*time.Second, notificationService)
}
