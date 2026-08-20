package actions

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "charm.land/bubbletea/v2"
)

var browserCommand = func(url string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url)
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return exec.Command("xdg-open", url)
	}
}

func OpenURL(url string) tea.Cmd {
	return track(func() tea.Msg {
		if url == "" {
			return FinishedMsg{Status: "Nothing to open", Err: ErrUnsupported{Reason: "no URL recorded"}}
		}
		if err := browserCommand(url).Start(); err != nil {
			return FinishedMsg{Status: "Could not open the browser", Err: fmt.Errorf("opening %s: %w", url, err)}
		}
		return FinishedMsg{Status: "Opened in your browser"}
	})
}
