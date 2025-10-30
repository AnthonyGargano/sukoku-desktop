package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func initLogger() (*os.File, error) {
	exe, err := os.Executable()
	if err != nil {
		exe = "."
	}
	base := filepath.Dir(exe)
	logDir := filepath.Join(base, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("make log dir: %w", err)
	}
	today := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("sudoku-%s.log", today))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	go func() {
		_ = maintainLogs(logDir, 7*24*time.Hour, 365*24*time.Hour)
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for range t.C {
			_ = maintainLogs(logDir, 7*24*time.Hour, 365*24*time.Hour)
		}
	}()
	return f, nil
}

func maintainLogs(dir string, compressAfter, deleteAfter time.Duration) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p := filepath.Join(dir, e.Name())
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		age := now.Sub(info.ModTime())
		if age > deleteAfter {
			_ = os.Remove(p)
			continue
		}
		if filepath.Ext(p) == ".log" && age > compressAfter {
			if err := gzipFile(p); err == nil {
				_ = os.Remove(p)
			}
		}
	}
	return nil
}

func gzipFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(path + ".gz")
	if err != nil {
		return err
	}
	defer out.Close()
	gw, err := gzip.NewWriterLevel(out, gzip.BestCompression)
	if err != nil {
		return err
	}
	gw.Name = filepath.Base(path)
	defer gw.Close()
	_, err = io.Copy(gw, in)
	return err
}
