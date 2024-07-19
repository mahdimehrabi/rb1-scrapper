package main

import (
	"fmt"
	"log/slog"
	"os"
	"rb-scrapper/utils"
)

func main() {
	drs := utils.NewDownloadResizer(2, slog.New(slog.NewTextHandler(os.Stdin, nil)), []string{"cat"})
	iuc := make(chan string, 10)

	go drs.Download(iuc)
	for i := range iuc {
		fmt.Println(i)
	}
}
