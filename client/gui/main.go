package gui

import (
	"bytes"
	_ "embed"
	"image/png"
	"wowa/core"

	g "github.com/AllenDang/giu"
)

// var searchTerm = ""
// var selectedTab = 0

var addons []core.LocalAddon

func buildAddonRows() []*g.TableRowWidget {
	var rows []*g.TableRowWidget
	for _, addon := range addons {
		// if searchTerm != "" && !giu.HasSubstringInsensitive(addon.Name, searchTerm) {
		// 	continue
		// }
		rows = append(rows, g.TableRow(
			// g.Label(addon.Icon),
			g.Label(addon.Name),
			g.Label(addon.Version),
			g.Label(string(addon.GameVersion)),
			// g.Label(addon.Latest),
			// g.Label(addon.GameVersion),
		))
	}
	return rows
}

func loop() {

	// g.Style().SetStyleFloat(g.StyleVarWindowPadding, 64).To()

	g.SingleWindow().Layout(
		g.Style().
			SetFontSize(24).
			To(g.Label("wowa addon manager")),

		g.Dummy(0, 32),
		g.Separator(),
		g.Dummy(0, 32),

		g.Table().FastMode(true).
			Columns(
				g.TableColumn("Name"),
				g.TableColumn("Version"),
				g.TableColumn("Game Version"),
			).
			Rows(
				buildAddonRows()...,
			),
	)
}

//go:embed icon.png
var iconBytes []byte

func Start(localAddonRepository *core.LocalAddonRepository) error {
	go func() {
		addons, _ = localAddonRepository.GetAll(nil)
	}()

	// Create the window
	wnd := g.NewMasterWindow("wowa", 1024, 900, 0)

	// Load icon
	iconReader := bytes.NewReader(iconBytes)
	iconPng, err := png.Decode(iconReader)
	if err != nil {
		return err
	}
	iconRgba := g.ImageToRgba(iconPng)
	wnd.SetIcon(iconRgba)

	// Set style
	style := g.Style()
	// style.SetColor(g.StyleColorWindowBg, color.RGBA{0x55, 0x55, 0x55, 255})
	style.SetStyle(g.StyleVarWindowPadding, 32, 32)
	wnd.SetStyle(style)

	// Run
	wnd.Run(loop)

	return nil
}
