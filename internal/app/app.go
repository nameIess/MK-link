//go:build windows

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/nameIess/MK-link/internal/linker"
)

const className = "MKLinkNativeWindow"
const windowTitle = "MK-Link"

const (
	wsOverlapped = 0x00000000
	wsCaption = 0x00C00000
	wsSysMenu = 0x00080000
	wsMinimizeBox = 0x00020000
	wsClipChildren = 0x02000000
	wsVisible = 0x10000000
	wsChild = 0x40000000
	wsTabStop = 0x00010000
	esAutoHScroll = 0x0080
	cbsDropDownList = 0x0003
	bsPushButton = 0x00000000
	swShow = 5
	cwUseDefault = 0x80000000
	wmCreate = 0x0001
	wmDestroy = 0x0002
	wmCommand = 0x0111
	wmClose = 0x0010
	wmCtlColorStatic = 0x0138
	wmCtlColorEdit = 0x0133
	cbAddString = 0x0143
	cbResetContent = 0x014B
	cbGetCurSel = 0x0147
	cbSetCurSel = 0x014E
	enChange = 0x0300
	bnClicked = 0
	colorWindow = 5
)

const (
	idTarget = 101
	idBrowseTarget = 102
	idType = 103
	idName = 104
	idLocation = 105
	idBrowseLocation = 106
	idCreate = 107
	idClear = 108
	idDescription = 109
	idStatus = 110
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32 = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW = user32.NewProc("CreateWindowExW")
	procDefWindowProcW = user32.NewProc("DefWindowProcW")
	procDestroyWindow = user32.NewProc("DestroyWindow")
	procShowWindow = user32.NewProc("ShowWindow")
	procUpdateWindow = user32.NewProc("UpdateWindow")
	procGetMessageW = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procPostQuitMessage = user32.NewProc("PostQuitMessage")
	procSetWindowTextW = user32.NewProc("SetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW = user32.NewProc("GetWindowTextW")
	procSendMessageW = user32.NewProc("SendMessageW")
	procEnableWindow = user32.NewProc("EnableWindow")
	procMessageBoxW = user32.NewProc("MessageBoxW")
	procLoadCursorW = user32.NewProc("LoadCursorW")
	procLoadIconW = user32.NewProc("LoadIconW")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	procSHBrowseForFolderW = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree = shell32.NewProc("CoTaskMemFree")
)

type point struct{ X, Y int32 }
type msg struct { Hwnd uintptr; Message uint32; WParam uintptr; LParam uintptr; Time uint32; Pt point }
type wndClassEx struct {
	Size uint32; Style uint32; WndProc uintptr; ClsExtra int32; WndExtra int32
	Instance uintptr; Icon uintptr; Cursor uintptr; Background uintptr; MenuName *uint16
	ClassName *uint16; IconSm uintptr
}
type openFileName struct {
	StructSize uint32; Owner uintptr; Instance uintptr; Filter *uint16; CustomFilter *uint16
	MaxCustFilter uint32; FilterIndex uint32; File *uint16; MaxFile uint32; FileTitle *uint16
	MaxFileTitle uint32; InitialDir *uint16; Title *uint16; Flags uint32; FileOffset uint16
	FileExtension uint16; DefExt *uint16; CustData uintptr; Hook uintptr; TemplateName *uint16
	Reserved1 uintptr; Reserved2 uint32; FlagsEx uint32
}

const (
	ofnPathMustExist = 0x00000800
	ofnFileMustExist = 0x00001000
	ofnNoChangeDir = 0x00000008
)

var current *window

type window struct {
	hwnd, target, typ, name, location, description, status, create uintptr
}

func Run() {
	hinst, _, _ := procGetModuleHandleW.Call(0)
	class := syscall.StringToUTF16Ptr(className)
	title := syscall.StringToUTF16Ptr(windowTitle)
	cursor, _, _ := procLoadCursorW.Call(0, 32512)
	icon := loadAppIcon(hinst)

	wc := wndClassEx{
		Size: uint32(unsafe.Sizeof(wndClassEx{})), WndProc: syscall.NewCallback(wndProc),
		Instance: hinst, Icon: icon, Cursor: cursor, Background: colorWindow + 1,
		ClassName: class, IconSm: icon,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, err := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)),
		wsOverlapped|wsCaption|wsSysMenu|wsMinimizeBox|wsClipChildren|wsVisible,
		cwUseDefault, cwUseDefault, 720, 560, 0, 0, hinst, 0)
	if hwnd == 0 { panic(fmt.Errorf("create window: %w", err)) }

	current = &window{hwnd: hwnd}
	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 { break }
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		current = &window{hwnd: hwnd}; buildUI(current); return 0
	case wmCommand:
		id := uint16(wParam & 0xffff); code := uint16((wParam >> 16) & 0xffff)
		if code == bnClicked {
			switch uintptr(id) {
			case idBrowseTarget: browseTarget(current)
			case idBrowseLocation: browseFolder(current)
			case idCreate: createLink(current)
			case idClear: clearForm(current)
			}
		} else if code == enChange && id == idTarget {
			syncTarget(current)
		} else if code == enChange && id == idType {
			updateDescription(current)
		}
		return 0
	case wmCtlColorStatic, wmCtlColorEdit:
		return 6
	case wmClose:
		procDestroyWindow.Call(hwnd); return 0
	case wmDestroy:
		procPostQuitMessage.Call(0); return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func buildUI(w *window) {
	addStatic(w.hwnd, "Target path", 28, 24, 120, 22, 0)
	w.target = addEdit(w.hwnd, "", 28, 48, 560, 32, idTarget)
	addButton(w.hwnd, "Browse...", 600, 48, 92, 32, idBrowseTarget)

	addStatic(w.hwnd, "Link type", 28, 96, 120, 22, 0)
	w.typ = addCombo(w.hwnd, 28, 120, 664, 120, idType)

	addStatic(w.hwnd, "Link name", 28, 166, 120, 22, 0)
	w.name = addEdit(w.hwnd, "", 28, 190, 664, 32, idName)

	addStatic(w.hwnd, "Destination folder", 28, 238, 160, 22, 0)
	w.location = addEdit(w.hwnd, "", 28, 262, 560, 32, idLocation)
	addButton(w.hwnd, "Browse...", 600, 262, 92, 32, idBrowseLocation)

	addStatic(w.hwnd, "About this link type", 28, 312, 180, 22, 0)
	w.description = addStatic(w.hwnd, "", 28, 338, 664, 70, idDescription)

	w.status = addStatic(w.hwnd, "Choose a target to begin.", 28, 430, 664, 40, idStatus)
	addButton(w.hwnd, "Clear", 390, 480, 96, 38, idClear)
	w.create = addButton(w.hwnd, "Create link", 500, 480, 192, 38, idCreate)
	setTypes(w, false)
	updateDescription(w)
	procEnableWindow.Call(w.create, 0)
}

func addStatic(parent uintptr, text string, x, y, width, height int32, id int) uintptr {
	return createControl("STATIC", text, wsChild|wsVisible, x, y, width, height, parent, uintptr(id))
}
func addEdit(parent uintptr, text string, x, y, width, height int32, id int) uintptr {
	return createControl("EDIT", text, wsChild|wsVisible|wsTabStop|esAutoHScroll, x, y, width, height, parent, uintptr(id))
}
func addButton(parent uintptr, text string, x, y, width, height int32, id int) uintptr {
	return createControl("BUTTON", text, wsChild|wsVisible|wsTabStop|bsPushButton, x, y, width, height, parent, uintptr(id))
}
func addCombo(parent uintptr, x, y, width, height int32, id int) uintptr {
	return createControl("COMBOBOX", "", wsChild|wsVisible|wsTabStop|cbsDropDownList, x, y, width, height, parent, uintptr(id))
}
func createControl(class, text string, style uint32, x, y, width, height int32, parent, id uintptr) uintptr {
	hinst, _, _ := procGetModuleHandleW.Call(0)
	c := syscall.StringToUTF16Ptr(class); t := syscall.StringToUTF16Ptr(text)
	h, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(t)), uintptr(style),
		uintptr(x), uintptr(y), uintptr(width), uintptr(height), parent, id, hinst, 0)
	return h
}
func addComboString(hwnd uintptr, text string) {
	t := syscall.StringToUTF16Ptr(text)
	procSendMessageW.Call(hwnd, cbAddString, 0, uintptr(unsafe.Pointer(t)))
}
func setTypes(w *window, isDir bool) {
	procSendMessageW.Call(w.typ, cbResetContent, 0, 0)
	if isDir {
		addComboString(w.typ, "Symbolic link (directory)")
		addComboString(w.typ, "Junction")
	} else {
		addComboString(w.typ, "Symbolic link (file)")
		addComboString(w.typ, "Hard link")
	}
	procSendMessageW.Call(w.typ, cbSetCurSel, 0, 0)
}
func updateDescription(w *window) {
	sel, _, _ := procSendMessageW.Call(w.typ, cbGetCurSel, 0, 0)
	dir := isDirectory(getText(w.target))
	var text string
	if dir {
		if sel == 0 {
			text = "A directory symbolic link redirects access to another directory. It can cross volumes and is a filesystem link."
		} else {
			text = "A junction is a Windows directory reparse point. Its target and link must be on the same volume."
		}
	} else if sel == 0 {
		text = "A file symbolic link redirects access to another file. The target can be on another volume."
	} else {
		text = "A hard link is another directory entry for the same file data. It is for files only and must stay on the same volume."
	}
	setText(w.description, text)
}
func syncTarget(w *window) {
	target := getText(w.target)
	if target == "" {
		procEnableWindow.Call(w.create, 0)
		return
	}
	setTypes(w, isDirectory(target))
	updateDescription(w)
	procEnableWindow.Call(w.create, 1)
}

func boolToUintptr(v bool) uintptr {
	if v { return 1 }
	return 0
}

func browseTarget(w *window) {
	path := chooseFile(w.hwnd)
	if path == "" { path = chooseFolder(w.hwnd) }
	if path == "" { return }
	setText(w.target, path); setTypes(w, isDirectory(path)); updateDescription(w)
	if getText(w.location) == "" { setText(w.location, filepath.Dir(path)) }
	procEnableWindow.Call(w.create, 1)
	setText(w.status, "Target selected. Review the link name and destination.")
}
func browseFolder(w *window) {
	if path := chooseFolder(w.hwnd); path != "" { setText(w.location, path) }
}
func createLink(w *window) {
	target := getText(w.target); name := strings.TrimSpace(getText(w.name)); location := getText(w.location)
	if target == "" || name == "" || location == "" { setText(w.status, "Target, link name, and destination folder are required."); return }
	if filepath.Base(name) != name || name == "." || name == ".." { setText(w.status, "Link name must be a single file or directory name."); return }
	linkPath := filepath.Join(location, name)
	dir := isDirectory(target)
	sel, _, _ := procSendMessageW.Call(w.typ, cbGetCurSel, 0, 0)
	var typ linker.LinkType
	if dir { if sel == 0 { typ = linker.DirectorySymbolicLink } else { typ = linker.Junction } } else { if sel == 0 { typ = linker.FileSymbolicLink } else { typ = linker.HardLink } }
	if err := linker.Create(target, linkPath, typ); err != nil { setText(w.status, "Failed: "+err.Error()); return }
	setText(w.status, "Created successfully: "+linkPath)
	procMessageBoxW.Call(w.hwnd, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("The link was created successfully."))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("MK-Link"))), 0x00000040)
}
func clearForm(w *window) {
	setText(w.target, ""); setText(w.name, ""); setText(w.location, "")
	setText(w.status, "Choose a target to begin."); setTypes(w, false); updateDescription(w)
	procEnableWindow.Call(w.create, 0)
}
func getText(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(hwnd); buf := make([]uint16, n+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}
func setText(hwnd uintptr, text string) {
	t := syscall.StringToUTF16Ptr(text); procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(t)))
}
func isDirectory(path string) bool { info, err := os.Stat(path); return err == nil && info.IsDir() }

func chooseFile(owner uintptr) string {
	buf := make([]uint16, 32768)
	filter := syscall.StringToUTF16Ptr("All files\x00*.*\x00\x00")
	title := syscall.StringToUTF16Ptr("Select target file")
	ofn := openFileName{StructSize: uint32(unsafe.Sizeof(openFileName{})), Owner: owner, Filter: filter,
		File: &buf[0], MaxFile: uint32(len(buf)), Title: title, Flags: ofnPathMustExist|ofnFileMustExist|ofnNoChangeDir}
	r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 { return "" }
	return syscall.UTF16ToString(buf)
}
func chooseFolder(owner uintptr) string {
	info := struct {
		Owner uintptr; Root uintptr; DisplayName *uint16; Title *uint16; Flags uint32; Callback uintptr; Param uintptr; Image int32
	}{Owner: owner, Title: syscall.StringToUTF16Ptr("Select folder"), Flags: 0x0001 | 0x0040}
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&info)))
	if pidl == 0 { return "" }
	defer procCoTaskMemFree.Call(pidl)
	buf := make([]uint16, 32768)
	ok, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
	if ok == 0 { return "" }
	return syscall.UTF16ToString(buf)
}
func loadAppIcon(hinst uintptr) uintptr {
	icon, _, _ := procLoadIconW.Call(hinst, 101)
	return icon
}
