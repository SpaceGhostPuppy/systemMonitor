package main

import (
    "fmt"
    "log"
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

// Monitor überwacht bestimmte Verzeichnisse auf Änderungen
func Monitor(paths []string, interval time.Duration) {
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
                    log.Printf("Neue Datei erkannt: %s\n", path)
                } else {
                    // Auf Änderungen prüfen
                    if info.Size() != oldInfo.Size {
                        log.Printf("Dateigröße geändert: %s\n", path)
                    }
                    if info.ModTime() != oldInfo.LastModified {
                        log.Printf("Datei modifiziert: %s\n", path)
                    }
                    if info.Mode() != oldInfo.Permissions {
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
    
    fmt.Println("Starte Dateisystem-Monitor...")
    Monitor(paths, 5*time.Second)
}
