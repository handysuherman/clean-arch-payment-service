package helper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type PDFDownloadConfig struct {
	WorkDir    string
	SavePath   string
	Filename   string
	Site       string
	UrlPath    string
	KodeUser   string
	KodeAsesor string
}

func DownloadPDF(cfg *PDFDownloadConfig) error {
	file := filepath.Join(cfg.WorkDir, cfg.SavePath, cfg.Filename)
	url := fmt.Sprintf("%s%s?asesi=%s&asesor=%s", cfg.Site, cfg.UrlPath, cfg.KodeUser, cfg.KodeAsesor)

	err := DownloadFile(file, url)
	if err != nil {
		return err
	}

	return nil
}

func DownloadPDFFile(cfg *PDFDownloadConfig, wg *sync.WaitGroup) error {
	defer wg.Done()
	_, err := os.Stat(filepath.Join(cfg.WorkDir, cfg.SavePath, cfg.Filename))
	if err != nil {
		if os.IsNotExist(err) {
			// Get the data
			resp, err := http.Get(fmt.Sprintf("%s%s?asesi=%s&asesor=%s", cfg.Site, cfg.UrlPath, cfg.KodeUser, cfg.KodeAsesor))
			if err != nil {
				log.Println(err)
				return err
			}
			defer resp.Body.Close()

			out, err := os.Create(filepath.Join(cfg.WorkDir, cfg.SavePath, cfg.Filename))
			if err != nil {
				log.Println(err)
				return err
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				log.Println(err)
			}
			return err
		}
	}

	return nil
}

// Download file from remote server, it will check whether file is exist or not. If it exist, request/download operation wont be executed.
func DownloadFile(filepath, url string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// Get the data
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			out, err := os.Create(filepath)
			if err != nil {
				return err
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, resp.Body)
			return err
		}
	}

	return err
}
