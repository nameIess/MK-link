//go:build windows

package ui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/nameIess/MK-link/internal/link"
	"golang.org/x/sys/windows"
)

const (
	windowClass = "MKLinkNativeWindow"
	windowTitle = "MK-Link"
	windowWidth = 820
	windowHeight = 640

	idTarget = 1001
	idBrowseFile = 1002
	idBrowseDir = 1003
	idType = 1004
	idDescription = 1005
	idName = 1006
	idDestination = 1007
	idBrowseDestination = 1008
	idCreate = 1009
	idStatus = 1010
)

const (
	wmDestroy = 0x0002
	wmClose = 0x0010
	wmCommand = 0x0111
	wmSetFont = 0x0030
	wmSize = 0x0005
	wmDpiChanged = 0x02E0

	bnClicked = 0
	cbSelectionChange = 1
	cbResetContent = 0x014B
	cbAddString = 0x0143
	cbSetCurSel = 0x014E
	cbGetCurSel = 0x0147

	cbsDropDownList = 0x0003

	wsChild = 0x40000000
	wsVisible = 0x10000000
	wsTabStop = 0x00010000
	wsBorder = 0x00800000
	wsOverlapped = 0
	wsCaption = 0x00C00000
	wsSysMenu = 0x00080000
	wsThickFrame = 0x00040000
	wsMinimizeBox = 0x00020000

	esAutoHScroll = 0x0080
	esMultiline = 0x0004
	esReadOnly = 0x0800
	bsPushButton = 0

	wsExClientEdge = 0x00000200

	gwlUserData = -21
	swShow = 5

	mbOk = 0
	mbIconWarning = 0x30
	mbIconError = 0x10
)

type app struct {
	logger      *slog.Logger
	hwnd        windows.Handle
	font        windows.Handle
	boldFont    windows.Handle
	targetEdit  windows.Handle
	typeCombo   windows.Handle
	description windows.Handle
	nameEdit    windows.Handle
	destEdit    windows.Handle
	create      windows.Handle
	status      windows.Handle

	target string
	targetIsDir bool
}

type point struct {
	x int32
	y int32
}

type message struct {
	hwnd windows.Handle
	msg uint32
	wParam uintptr
	lParam uintptr
	time uint32
	pt point
	lPrivate uint32
}

type wndClassEx struct {
	size uint32
	style uint32
	wndProc uintptr
	cbClsExtra int32
	cbWndExtra int32
	instance windows.Handle
	icon windows.Handle
	cursor windows.Handle
	background windows.Handle
	menu *uint16
	className *uint16
	iconSm windows.Handle
}

type openFileName struct {
	structSize uint32
	owner windows.Handle
	instance windows.Handle
	filter *uint16
	customFilter *uint16
	maxCustFilter uint32
	filterIndex uint32
	file *uint16
	maxFile uint32
	fileTitle *uint16
	maxFileTitle uint32
	initialDir *uint16
	title *uint16
	flags uint32
	fileOffset uint16
	fileExtension uint16
	defaultExt *uint16
	custData uintptr
	hook uintptr
	templateName *uint16
	pvReserved uintptr
	reserved uint32
	flagsEx uint32
}

type browseInfo struct {
	owner windows.Handle
	root windows.Handle
	displayName *uint16
	title *uint16
	flags uint32
	callback uintptr
	param uintptr
	image int32
}

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	comdlg32 = windows.NewLazySystemDLL("comdlg32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")
	ole32 = windows.NewLazySystemDLL("ole32.dll")
	comctl32 = windows.NewLazySystemDLL("comctl32.dll")

	registerClassExW = user32.NewProc("RegisterClassExW")
	createWindowExW = user32.NewProc("CreateWindowExW")
	defWindowProcW = user32.NewProc("DefWindowProcW")
	getMessageW = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	dispatchMessage = user32.NewProc("DispatchMessageW")
	postQuitMessage = user32.NewProc("PostQuitMessage")
	showWindow = user32.NewProc("ShowWindow")
	updateWindow = user32.NewProc("UpdateWindow")
	sendMessageW = user32.NewProc("SendMessageW")
	setWindowTextW = user32.NewProc("SetWindowTextW")
	getWindowTextW = user32.NewProc("GetWindowTextW")
	getWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	moveWindow = user32.NewProc("MoveWindow")
	getClientRect = user32.NewProc("GetClientRect")
	loadImageW = user32.NewProc("LoadImageW")
	createFontW = user32.NewProc("CreateFontW")
	initCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	getOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	shBrowseForFolderW = shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	coTaskMemFree = ole32.NewProc("CoTaskMemFree")
	getModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	getModuleFileNameW = kernel32.NewProc("GetModuleFileNameW")
)

var windowProc = windows.NewCallback(wndProc)
var currentApp *app

var errDialogCancelled = errors.New("dialog cancelled")

func Run(logger *slog.Logger) error {
	if logger == nil {
		return errors.New("logger is required")
	}
	if err := initControls(); err != nil {
		return err
	}
	return runMessageLoop(logger)
}

func initControls() error {
	type initCommonControlsExData struct {
		size uint32
		classes uint32
	}
	data := initCommonControlsExData{
		size: uint32(unsafe.Sizeof(initCommonControlsExData{})),
		classes: 0x00004000,
	}
	r1, _, callErr := initCommonControlsEx.Call(uintptr(unsafe.Pointer(&data)))
	if r1 == 0 {
		return fmt.Errorf("initialize common controls: %w", callErr)
	}
	return nil
}

func runMessageLoop(logger *slog.Logger) error {
	instanceRaw, _, callErr := getModuleHandleW.Call(0)
	if instanceRaw == 0 {
		return fmt.Errorf("get module handle: %w", callErr)
	}
	instance := windows.Handle(instanceRaw)

	className, err := windows.UTF16PtrFromString(windowClass)
	if err != nil {
		return fmt.Errorf("encode window class: %w", err)
	}
	title, err := windows.UTF16PtrFromString(windowTitle)
	if err != nil {
		return fmt.Errorf("encode window title: %w", err)
	}

	icon := loadAppIcon(instance)
	wc := wndClassEx{
		size: uint32(unsafe.Sizeof(wndClassEx{})),
		style: 0x0003,
		wndProc: windowProc,
		instance: instance,
		icon: icon,
		className: className,
		iconSm: icon,
	}
	if r1, _, callErr := registerClassExW.Call(uintptr(unsafe.Pointer(&wc))); r1 == 0 {
		return fmt.Errorf("register window class: %w", callErr)
	}

	a := &app{logger: logger}
	currentApp = a
	hwnd, _, callErr := createWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsOverlapped|wsCaption|wsSysMenu|wsThickFrame|wsMinimizeBox|wsVisible,
		0x80000000,
		0x80000000,
		windowWidth,
		windowHeight,
		0,
		0,
		uintptr(instance),
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("create main window: %w", callErr)
	}
	a.hwnd = windows.Handle(hwnd)

	showWindow.Call(hwnd, swShow)
	updateWindow.Call(hwnd)

	for {
		var msg message
		r1, _, callErr := getMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		switch int32(r1) {
		case -1:
			return fmt.Errorf("get windows message: %w", callErr)
		case 0:
			return nil
		default:
			translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}

func wndProc(hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmDestroy:
		postQuitMessage.Call(0)
		return 0
	case wmClose:
		result, _, _ := messageBoxW.Call(
			uintptr(hwnd),
			uintptr(unsafe.Pointer(mustUTF16("Close MK-Link?"))),
			uintptr(unsafe.Pointer(mustUTF16(windowTitle))),
			mbOk|mbIconWarning,
		)
		_ = result
		user32.NewProc("DestroyWindow").Call(uintptr(hwnd))
		return 0
	case wmCommand:
		if currentApp != nil {
			currentApp.onCommand(wParam)
		}
		return 0
	case wmSize, wmDpiChanged:
		if currentApp != nil {
			currentApp.layout()
		}
		return 0
	default:
		return defWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	}
}

func (a *app) onCommand(wParam uintptr) {
	id := uint16(wParam & 0xFFFF)
	notify := uint16((wParam >> 16) & 0xFFFF)
	switch {
	case id == idBrowseFile && notify == bnClicked:
		a.pickFile()
	case id == idBrowseDir && notify == bnClicked:
		a.pickFolder("Choose the target directory")
	case id == idBrowseDestination && notify == bnClicked:
		a.pickFolder("Choose the destination folder")
	case id == idType && notify == cbSelectionChange:
		a.updateDescription()
	case id == idCreate && notify == bnClicked:
		a.createLink()
	}
}

func (a *app) createControls() {
	a.font = newFont(10, 400)
	a.boldFont = newFont(11, 700)

	a.targetEdit = a.makeControl("EDIT", "", wsChild|wsVisible|wsTabStop|esAutoHScroll|wsExClientEdge, 24, 56, 560, 28, idTarget)
	a.makeLabel("Target", 24, 28, 120, 22, a.boldFont)
	a.makeButton("Browse file…", 594, 56, 105, 28, idBrowseFile)
	a.makeButton("Browse folder…", 707, 56, 95, 28, idBrowseDir)

	a.makeLabel("Link type", 24, 104, 120, 22, a.boldFont)
	a.typeCombo = a.makeControl("COMBOBOX", "", wsChild|wsVisible|wsTabStop|cbsDropDownList, 24, 132, 778, 30, idType)

	a.makeLabel("Description", 24, 176, 120, 22, a.boldFont)
	a.description = a.makeControl("EDIT", "", wsChild|wsVisible|esMultiline|esReadOnly, wsExClientEdge, 24, 204, 778, 72, idDescription)

	a.makeLabel("Link name", 24, 298, 120, 22, a.boldFont)
	a.nameEdit = a.makeControl("EDIT", "", wsChild|wsVisible|wsTabStop|esAutoHScroll, wsExClientEdge, 24, 326, 778, 28, idName)

	a.makeLabel("Destination folder", 24, 378, 160, 22, a.boldFont)
	a.destEdit = a.makeControl("EDIT", "", wsChild|wsVisible|wsTabStop|esAutoHScroll, wsExClientEdge, 24, 406, 665, 28, idDestination)
	a.makeButton("Browse…", 699, 406, 103, 28, idBrowseDestination)

	a.makeLabel("Status", 24, 456, 120, 22, a.boldFont)
	a.status = a.makeControl("EDIT", "", wsChild|wsVisible|esMultiline|esReadOnly, wsExClientEdge, 24, 484, 778, 62, idStatus)
	a.makeButton("Create link", 642, 566, 160, 34, idCreate)

	a.layout()
	a.refreshTypes()
}

func (a *app) makeLabel(text string, x, y, w, h int32, font windows.Handle) {
	hwnd := a.makeControl("STATIC", text, wsChild|wsVisible, 0, x, y, w, h, 0)
	if font != 0 {
		sendMessageW.Call(uintptr(hwnd), wmSetFont, uintptr(font), 1)
	}
}

func (a *app) makeButton(text string, x, y, w, h int32, id uintptr) windows.Handle {
	return a.makeControl("BUTTON", text, wsChild|wsVisible|wsTabStop|bsPushButton, 0, x, y, w, h, id)
}

func (a *app) makeControl(class, text string, style, exStyle uint32, x, y, w, h int32, id uintptr) windows.Handle {
	classPtr := mustUTF16(class)
	textPtr := mustUTF16(text)
	hwnd, _, _ := createWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(classPtr)),
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(a.hwnd),
		id,
		0,
		0,
	)
	handle := windows.Handle(hwnd)
	if handle != 0 && a.font != 0 {
		sendMessageW.Call(uintptr(handle), wmSetFont, uintptr(a.font), 1)
	}
	return handle
}

func (a *app) layout() {
	var client rect
	getClientRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&client)))
	width := client.right - client.left
	if width < 820 {
		width = 820
	}

	moveWindow.Call(uintptr(a.targetEdit), 24, 56, width-260, 28, 1)
	moveWindow.Call(uintptr(a.typeCombo), 24, 132, width-48, 30, 1)
	moveWindow.Call(uintptr(a.description), 24, 204, width-48, 72, 1)
	moveWindow.Call(uintptr(a.nameEdit), 24, 326, width-48, 28, 1)
	moveWindow.Call(uintptr(a.destEdit), 24, 406, width-160, 28, 1)
	moveWindow.Call(uintptr(a.status), 24, 484, width-48, 62, 1)
	moveWindow.Call(uintptr(a.create), width-178, 566, 154, 34, 1)
}

type rect struct {
	left int32
	top int32
	right int32
	bottom int32
}

func (a *app) refreshTypes() {
	sendMessageW.Call(uintptr(a.typeCombo), cbResetContent, 0, 0)
	if a.target == "" {
		a.addType("Select a target first", "")
		setWindowTextW.Call(uintptr(a.description), uintptr(unsafe.Pointer(mustUTF16("Choose a file or folder to see the compatible link types."))))
		return
	}

	if a.targetIsDir {
		a.addType("Symbolic link", string(link.Symbolic))
		a.addType("Junction", string(link.Junction))
	} else {
		a.addType("Symbolic link", string(link.Symbolic))
		a.addType("Hard link", string(link.Hardlink))
	}
	sendMessageW.Call(uintptr(a.typeCombo), cbSetCurSel, 0, 0)
	a.updateDescription()
	if a.text(a.nameEdit) == "" {
		if name := filepath.Base(a.target); name != "." {
			setWindowTextW.Call(uintptr(a.nameEdit), uintptr(unsafe.Pointer(mustUTF16(name))))
		}
	}
}

func (a *app) addType(label, value string) {
	sendMessageW.Call(uintptr(a.typeCombo), cbAddString, 0, uintptr(unsafe.Pointer(mustUTF16(label))))
	_ = value
}

func (a *app) selectedType() (link.Type, error) {
	idxRaw, _, _ := sendMessageW.Call(uintptr(a.typeCombo), cbGetCurSel, 0, 0)
	if idxRaw == 0xFFFFFFFF {
		return "", errors.New("select a link type")
	}
	if a.targetIsDir {
		switch idxRaw {
		case 0:
			return link.Symbolic, nil
		case 1:
			return link.Junction, nil
		}
	} else {
		switch idxRaw {
		case 0:
			return link.Symbolic, nil
		case 1:
			return link.Hardlink, nil
		}
	}
	return "", errors.New("select a link type")
}

func (a *app) updateDescription() {
	if a.target == "" {
		return
	}
	typ, err := a.selectedType()
	if err != nil {
		return
	}
	setWindowTextW.Call(uintptr(a.description), uintptr(unsafe.Pointer(mustUTF16(typ.Description()))))
}

func (a *app) createLink() {
	target := strings.TrimSpace(a.text(a.targetEdit))
	name := strings.TrimSpace(a.text(a.nameEdit))
	destinationDir := strings.TrimSpace(a.text(a.destEdit))
	typ, err := a.selectedType()
	if err != nil {
		a.setStatus("Select a valid link type.")
		return
	}
	if target == "" || name == "" || destinationDir == "" {
		a.setStatus("Target, link name, and destination folder are required.")
		return
	}

	destination := filepath.Join(destinationDir, name)
	plan, err := link.PlanLink(target, destination, typ)
	if err != nil {
		a.logger.Warn("link validation failed", "error", err)
		a.setStatus(userMessage(err))
		return
	}

	a.setStatus("Creating link…")
	if err := link.Create(plan); err != nil {
		a.logger.Error("link creation failed", "type", typ.Label(), "error", err)
		a.setStatus(userMessage(err))
		return
	}

	a.logger.Info("link created", "type", typ.Label(), "destination", plan.Destination)
	a.setStatus(fmt.Sprintf("%s created successfully at %s", typ.Label(), plan.Destination))
}

func (a *app) setTarget(path string) {
	info, err := os.Stat(path)
	if err != nil {
		a.setStatus(userMessage(fmt.Errorf("inspect target: %w", err)))
		return
	}
	a.target = filepath.Clean(path)
	a.targetIsDir = info.IsDir()
	setWindowTextW.Call(uintptr(a.targetEdit), uintptr(unsafe.Pointer(mustUTF16(a.target))))
	a.refreshTypes()
}

func (a *app) pickFile() {
	path, err := openFile(a.hwnd)
	if err == nil {
		a.setTarget(path)
		return
	}
	if !errors.Is(err, errDialogCancelled) {
		a.setStatus(userMessage(err))
	}
}

func (a *app) pickFolder(title string) {
	path, err := browseFolder(a.hwnd, title)
	if err == nil {
		if a.target == "" {
			a.setTarget(path)
		} else {
			setWindowTextW.Call(uintptr(a.destEdit), uintptr(unsafe.Pointer(mustUTF16(path))))
		}
		return
	}
	if !errors.Is(err, errDialogCancelled) {
		a.setStatus(userMessage(err))
	}
}

func (a *app) text(hwnd windows.Handle) string {
	lengthRaw, _, _ := getWindowTextLengthW.Call(uintptr(hwnd))
	length := int(lengthRaw)
	if length <= 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	writtenRaw, _, _ := getWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf[:writtenRaw])
}

func openFile(owner windows.Handle) (string, error) {
	buffer := make([]uint16, 32768)
	filter := mustUTF16("All files\x00*.*\x00\x00")
	title := mustUTF16("Select target file")
	dialog := openFileName{
		structSize: uint32(unsafe.Sizeof(openFileName{})),
		owner: owner,
		filter: filter,
		file: &buffer[0],
		maxFile: uint32(len(buffer)),
		title: title,
		flags: 0x00001000 | 0x00000800 | 0x00000004,
	}
	r1, _, callErr := getOpenFileNameW.Call(uintptr(unsafe.Pointer(&dialog)))
	if r1 == 0 {
		if callErr == windows.ERROR_CANCELLED {
			return "", errDialogCancelled
		}
		return "", fmt.Errorf("open file dialog: %w", callErr)
	}
	return windows.UTF16ToString(buffer), nil
}

func browseFolder(owner windows.Handle, title string) (string, error) {
	displayName := make([]uint16, 32768)
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return "", fmt.Errorf("encode folder dialog title: %w", err)
	}
	info := browseInfo{
		owner: owner,
		displayName: &displayName[0],
		title: titlePtr,
		flags: 0x0001 | 0x0040 | 0x0008,
	}
	r1, _, _ := shBrowseForFolderW.Call(uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		return "", errDialogCancelled
	}
	defer coTaskMemFree.Call(r1)
	path := make([]uint16, 32768)
	ok, _, callErr := shGetPathFromIDListW.Call(r1, uintptr(unsafe.Pointer(&path[0])))
	if ok == 0 {
		return "", fmt.Errorf("read selected folder: %w", callErr)
	}
	return windows.UTF16ToString(path), nil
}

func newFont(size, weight int32) windows.Handle {
	hwnd, _, _ := createFontW.Call(
		uintptr(size), 0, 0, 0, uintptr(weight),
		0, 0, 0, 1,
		0, 0, 0, 0,
		uintptr(unsafe.Pointer(mustUTF16("Segoe UI"))),
	)
	return windows.Handle(hwnd)
}

func loadAppIcon(instance windows.Handle) windows.Handle {
	exeDir := executableDir()
	if exeDir == "" {
		return 0
	}
	path := filepath.Join(exeDir, "ui", "icon.ico")
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}
	const (
		imageIcon = 1
		lrLoadFromFile = 0x00000010
		lrDefaultSize = 0x00000040
	)
	r1, _, _ := loadImageW.Call(
		uintptr(instance),
		uintptr(unsafe.Pointer(pathPtr)),
		imageIcon,
		32,
		32,
		lrLoadFromFile|lrDefaultSize,
	)
	return windows.Handle(r1)
}

func executableDir() string {
	buffer := make([]uint16, 32768)
	r1, _, _ := getModuleFileNameW.Call(0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if r1 == 0 {
		return ""
	}
	return filepath.Dir(windows.UTF16ToString(buffer[:r1]))
}

func mustUTF16(text string) *uint16 {
	ptr, _ := windows.UTF16PtrFromString(text)
	return ptr
}

func userMessage(err error) string {
	switch {
	case errors.Is(err, link.ErrTargetNotFound):
		return "The selected target does not exist."
	case errors.Is(err, link.ErrDestinationExists):
		return "The destination already exists. Choose a different name or folder."
	case errors.Is(err, link.ErrInvalidLinkName):
		return "The link name is not a valid Windows name."
	case errors.Is(err, link.ErrTargetTypeMismatch):
		return "That link type is not valid for the selected target."
	case errors.Is(err, link.ErrDifferentVolume):
		return "Hard links require the target and destination to be on the same volume."
	case errors.Is(err, link.ErrNetworkJunction):
		return "Junctions require local Windows volumes."
	case errors.Is(err, link.ErrNestedDestination):
		return "The destination folder cannot be inside the target directory."
	default:
		return fmt.Sprintf("Operation failed: %v", err)
	}
}
