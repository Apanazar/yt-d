package main

import (
	"log"
	"net/url"
	"os"
	"os/exec"
	"sync"

	"github.com/gotk3/gotk3/gtk"
)

func isYouTubeLink(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		log.Println(err)
		return false
	}
	return u.Host == "www.youtube.com" || u.Host == "youtube.com" || u.Host == "youtu.be"
}

func downloadFile(dir, link, ytType string, wg *sync.WaitGroup, textBuffer *gtk.TextBuffer) {
	defer wg.Done()

	var cmd *exec.Cmd
	switch ytType {
	case "Audio":
		cmd = exec.Command("yt-dlp", "-x", "-i", "--no-playlist", "-P", dir, link)
	case "Video":
		cmd = exec.Command("yt-dlp", "-f", "best", "--no-playlist", "-P", dir, link)
	case "Playlist":
		cmd = exec.Command("yt-dlp", "-f", "best", "--only-playlist", "-P", dir, link)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Error obtaining stdout pipe:", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		log.Println("Error starting command:", err)
		return
	}

	buf := make([]byte, 1024)
	go func() {
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				log.Println("Error reading stdout:", err)
				break
			}

			text := string(buf[:n])

			textBuffer.InsertAtCursor(text)
		}
	}()

	err = cmd.Wait()
	if err != nil {
		log.Println("Error waiting for command completion:", err)
	}
}

func main() {
	gtk.Init(nil)

	window, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	window.SetTitle("YT-D")
	window.SetDefaultSize(500, 400)
	window.SetResizable(false)
	window.SetIconFromFile("./src/img/icon.png")

	vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	textView, _ := gtk.TextViewNew()
	textBuffer, _ := textView.GetBuffer()
	textView.SetSizeRequest(-1, 150)
	textView.SetSensitive(false)

	entry, _ := gtk.EntryNew()
	entry.SetPlaceholderText("Enter YouTube URL")
	entry.SetWidthChars(40)

	combo, _ := gtk.ComboBoxTextNew()
	combo.AppendText("Video")
	combo.AppendText("Audio")
	combo.AppendText("Playlist")

	button, _ := gtk.ButtonNewWithLabel("\tStart\t")
	button.SetHAlign(gtk.ALIGN_CENTER)

	button.SetMarginTop(20)
	button.SetMarginBottom(20)

	button.Connect("clicked", func() {
		urlLink, _ := entry.GetText()
		downloadsPath, err := os.Getwd()
		if err != nil {
			log.Println("Error getting current directory:", err)
			return
		}

		if isYouTubeLink(urlLink) {
			var wg sync.WaitGroup
			wg.Add(1)
			go downloadFile(downloadsPath, urlLink, combo.GetActiveText(), &wg, textBuffer)
		} else {
			dialog := gtk.MessageDialogNew(window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "The data entered is not a YouTube URL")
			dialog.Run()
			dialog.Destroy()
		}
	})

	vbox.PackStart(textView, true, true, 10)
	vbox.PackStart(entry, false, false, 5)
	vbox.PackStart(combo, false, false, 0)
	vbox.PackStart(button, false, false, 0)

	vbox.SetMarginTop(20)
	vbox.SetMarginBottom(20)
	vbox.SetMarginStart(20)
	vbox.SetMarginEnd(20)

	window.Add(vbox)
	window.ShowAll()

	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	gtk.Main()
}
