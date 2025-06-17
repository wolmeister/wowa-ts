package gui

// #cgo linux pkg-config: gtk+-3.0
// #include "probe.h"
import "C"

import (
	"os"
	"runtime"
	"unsafe"

	webview "github.com/webview/webview_go"
)

//export on_create_window
func on_create_window(window *C.GtkWidget) {
	w := webview.NewWindow(false, unsafe.Pointer(window))
	w.SetHtml("<body style='color: red; background-color: yellow;'>Hello world</body>")
	// fmt.Printf("Go: callback called with result = %d\n", int(window))
}

func StartGui() {
	if runtime.GOOS == "linux" {
		os.Setenv("__NV_DISABLE_EXPLICIT_SYNC", "1")
	}

	C.new_application()

}
