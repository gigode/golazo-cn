package ui

import (
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/ui/design"
	"github.com/charmbracelet/lipgloss"
)

// Truncate truncates text to fit the specified width, appending "..." if truncated.
func Truncate(text string, width int) string {
	if len(text) <= width {
		return text
	}
	return text[:width-3] + "..."
}

// renderStatusBanner renders a status banner based on the specified type.
// Returns an empty string if no banner should be displayed.
// The banner is styled with cyan color, bold text, and center alignment.
// The new version banner uses a gradient effect.
func renderStatusBanner(bannerType constants.StatusBannerType, width int) string {
	var message string

	switch bannerType {
	case constants.StatusBannerDebug:
		message = "[调试模式] 日志：~/.golazo/golazo_debug.log"
	case constants.StatusBannerNewVersion:
		message = "发现新版本！运行 'golazo --update'"
	case constants.StatusBannerDev:
		message = "[开发版] 当前是开发版本"
	case constants.StatusBannerNone:
		fallthrough
	default:
		return "" // No banner for None or unknown types
	}

	var styledMessage string

	if bannerType == constants.StatusBannerNewVersion {
		// Apply gradient to new version banner (cyan → red, adaptive)
		styledMessage = design.ApplyGradientToText(message)
	} else {
		// Use simple cyan styling for other banners
		bannerStyle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)
		styledMessage = bannerStyle.Render(message)
	}

	// Center the banner in the available width
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return containerStyle.Render(styledMessage)
}
