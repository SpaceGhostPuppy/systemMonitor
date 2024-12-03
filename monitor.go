package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
)

// FileInfo stores metadata about monitored files
type FileInfo struct {
    Path         string
    Size         int64
    LastModified time.Time
    Permissions  os.FileMode
}

// Monitor watches specified directories for changes
func Monitor(paths []string, interval time.Duration) {
    fileMap := make(map[string]FileInfo)
    
    // Initial scan of files
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
            log.Printf("Error scanning path %s: %v\n", path, err)
        }
    }
    
    // Monitoring loop
    ticker := time.NewTicker(interval)
    for range ticker.C {
        for _, path := range paths {
            err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }
                
                // Check if file is new
                oldInfo, exists := fileMap[path]
                if !exists {
                    log.Printf("New file detected: %s\n", path)
                } else {
                    // Check for modifications
                    if info.Size() != oldInfo.Size {
                        log.Printf("File size changed: %s\n", path)
                    }
                    if info.ModTime() != oldInfo.LastModified {
                        log.Printf("File modified: %s\n", path)
                    }
                    if info.Mode() != oldInfo.Permissions {
                        log.Printf("File permissions changed: %s\n", path)
                    }
                }
                
                // Update file info
                fileMap[path] = FileInfo{
                    Path:         path,
                    Size:         info.Size(),
                    LastModified: info.ModTime(),
                    Permissions:  info.Mode(),
                }
                return nil
            })
            if err != nil {
                log.Printf("Error monitoring path %s: %v\n", path, err)
            }
        }
    }
}

func main() {
    // Directories to monitor
    paths := []string{
        ".", // Current directory
        // Add more paths as needed
    }
    
    fmt.Println("Starting file system monitor...")
    Monitor(paths, 5*time.Second)
}
