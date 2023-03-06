package main

import (
	"math"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func isYouTubeLink(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Host == "www.youtube.com" || u.Host == "youtube.com" || u.Host == "youtu.be"
}

func downloadFile(dir, link, ytType string, wg *sync.WaitGroup, pb *widgets.QProgressBar) {
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

	re := regexp.MustCompile(`\[download\]\s+(\d+.\d)%\s+of`)

	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	cmd.Run()

	buf := make([]byte, 1024)
	go func() {
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}

			text := string(buf[:n])
			match := re.FindStringSubmatch(text)

			if len(match) > 0 {
				percent := match[1]
				valf, _ := strconv.ParseFloat(percent, 64)
				value := int(math.Round(valf*10) / 10)

				pb.SetValue(value)
			}
		}
	}()
}

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	icon := gui.NewQIcon5("./src/img/icon.png")

	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowIcon(icon)
	window.SetFixedSize2(500, 300)
	window.SetWindowTitle("YT-D")

	label := widgets.NewQLabel2("Status:", nil, 0)

	progressBar := widgets.NewQProgressBar(nil)
	progressBar.SetContentsMargins(0, 0, 0, 30)
	progressBar.SetFixedHeight(20)
	progressBar.SetMinimum(0)
	progressBar.SetMaximum(100)
	progressBar.SetValue(0)

	entry := widgets.NewQLineEdit(nil)
	entry.SetPlaceholderText("URL")
	entry.SetFixedHeight(25)

	combo := widgets.NewQComboBox(nil)
	combo.SetFixedHeight(35)
	combo.AddItem("Video", core.NewQVariant1("Video"))
	combo.AddItem("Audio", core.NewQVariant1("Audio"))
	combo.AddItem("Playlist", core.NewQVariant1("Playlist"))

	button := widgets.NewQPushButton2("OK", nil)
	button.ConnectClicked(func(bool) {
		urlLink := entry.Text()

		downloadsPath, err := os.Getwd()
		if err != nil {
			return
		}

		if isYouTubeLink(entry.Text()) {
			var wg sync.WaitGroup
			wg.Add(1)
			go downloadFile(downloadsPath, urlLink, combo.CurrentText(), &wg, progressBar)
			wg.Wait()
		} else {
			widgets.QMessageBox_Information(
				nil,
				"YT-D: Info",
				"The data entered is not a YouTube URL",
				widgets.QMessageBox__Ok,
				widgets.QMessageBox__Ok,
			)
		}
	})

	widget1 := widgets.NewQFrame(nil, 0)
	widget1Layout := widgets.NewQVBoxLayout2(widget1)
	widget1Layout.AddWidget(label, 0, 0)
	widget1Layout.AddSpacing(-60)
	widget1Layout.AddWidget(progressBar, 0, 0)

	widget2 := widgets.NewQFrame(nil, 0)
	widget2Layout := widgets.NewQVBoxLayout2(widget2)
	widget2Layout.AddWidget(entry, 0, 0)
	widget2Layout.AddWidget(combo, 0, 0)
	widget2Layout.AddSpacing(25)
	widget2Layout.AddWidget(button, 0, 0)

	mainLayout := widgets.NewQVBoxLayout()
	mainLayout.AddWidget(widget1, 0, 0)
	mainLayout.AddSpacing(15)
	mainLayout.AddWidget(widget2, 0, 0)

	centralWidget := widgets.NewQWidget(nil, 0)
	centralWidget.SetLayout(mainLayout)

	window.SetCentralWidget(centralWidget)
	window.Show()
	app.Exec()
}
