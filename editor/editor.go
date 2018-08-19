package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	// "time"

	clipb "github.com/atotto/clipboard"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/neovim/go-client/nvim"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

var editor *Editor

// Highlight is
type Highlight struct {
	foreground *RGBA
	background *RGBA
	bold       bool
	italic     bool
}

// Char is
type Char struct {
	normalWidth bool
	char        string
	highlight   Highlight
}

type NotifyButton struct {
	action func()
	text   string
}

type Notify struct {
	level   NotifyLevel
	period  int
	message string
	buttons []*NotifyButton
}

// Editor is the editor
type Editor struct {
	signal  *editorSignal
	version string
	app     *widgets.QApplication

	activity       *Activity
	splitter       *widgets.QSplitter
	notifyStartPos *core.QPoint
	notify         chan *Notify

	workspaces []*Workspace
	active     int
	nvim       *nvim.Nvim
	window     *widgets.QMainWindow
	wsWidget   *widgets.QWidget
	wsSide     *WorkspaceSide
	deinSide   *DeinSide

	statuslineHeight int
	width            int
	height           int
	tablineHeight    int
	selectedBg       *RGBA
	matchFg          *RGBA
	bgcolor          *RGBA
	fgcolor          *RGBA

	stop     chan struct{}
	stopOnce sync.Once

	specialKeys     map[core.Qt__Key]string
	controlModifier core.Qt__KeyboardModifier
	cmdModifier     core.Qt__KeyboardModifier
	shiftModifier   core.Qt__KeyboardModifier
	altModifier     core.Qt__KeyboardModifier
	metaModifier    core.Qt__KeyboardModifier
	keyControl      core.Qt__Key
	keyCmd          core.Qt__Key
	keyAlt          core.Qt__Key
	keyShift        core.Qt__Key

	config               gonvimConfig
	notifications        []*Notification
	displayNotifications bool
}

type editorSignal struct {
	core.QObject
	_ func() `signal:"notifySignal"`
}

func (hl *Highlight) copy() Highlight {
	highlight := Highlight{}
	if hl.foreground != nil {
		highlight.foreground = hl.foreground.copy()
	}
	if hl.background != nil {
		highlight.background = hl.background.copy()
	}
	highlight.bold = hl.bold
	highlight.italic = hl.italic

	return highlight
}

// InitEditor is
func InitEditor() {
	home, err := homedir.Dir()
	if err != nil {
		home = "~"
	}
	config := newGonvimConfig(home)
	editor = &Editor{
		version:    "v0.2.3",
		signal:     NewEditorSignal(nil),
		notify:     make(chan *Notify, 10),
		selectedBg: newRGBA(81, 154, 186, 0.5),
		matchFg:    newRGBA(81, 154, 186, 1),
		bgcolor:    nil,
		fgcolor:    nil,
		stop:       make(chan struct{}),
		config:     config,
	}
	e := editor
	e.signal.ConnectNotifySignal(func() {
		// Is there a smarter implementation ?
		notify := <-e.notify
		if notify.message == "" {
			return
		}
		if notify.buttons == nil {
			e.popupNotification(notify.level, notify.period, notify.message)
		} else {
			e.popupNotification(notify.level, notify.period, notify.message, notifyOptionArg(notify.buttons))
		}
	})
	e.app = widgets.NewQApplication(0, nil)
	e.app.ConnectAboutToQuit(func() {
		editor.cleanup()
	})

	e.width = 800
	e.height = 600

	//create a window
	e.window = widgets.NewQMainWindow(nil, 0)
	e.window.SetWindowTitle("Gonvim")
	e.window.SetContentsMargins(0, 0, 0, 0)
	e.window.SetMinimumSize2(e.width, e.height)

	e.initSpecialKeys()
	e.window.ConnectKeyPressEvent(e.keyPress)
	//e.window.ConnectKeyReleaseEvent(e.keyRelease)
	e.window.SetAcceptDrops(true)

	// output log
	// tfile, terr := os.OpenFile("/Users/akiyoshi/test.log", os.O_WRONLY | os.O_CREATE, 0666)
	// if terr != nil {
	//     panic(terr)
	// }
	// defer tfile.Close()

	widget := widgets.NewQWidget(nil, 0)
	widget.SetContentsMargins(0, 0, 0, 0)

	layout := widgets.NewQBoxLayout(widgets.QBoxLayout__RightToLeft, widget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)

	e.wsWidget = widgets.NewQWidget(nil, 0)
	e.wsSide = newWorkspaceSide()

	sideArea := widgets.NewQScrollArea(nil)
	sideArea.SetWidgetResizable(true)
	sideArea.SetVerticalScrollBarPolicy(core.Qt__ScrollBarAsNeeded)
	sideArea.SetFocusPolicy(core.Qt__ClickFocus)
	sideArea.SetWidget(e.wsSide.widget)
	sideArea.SetFrameShape(widgets.QFrame__NoFrame)
	// sideArea.SetMaximumWidth(e.config.sideWidth)
	// sideArea.SetMinimumWidth(e.config.sideWidth)
	// sideArea.SetSizePolicy2(widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Expanding)
	e.wsSide.scrollarea = sideArea

	e.workspaces = []*Workspace{}
	sessionExists := false
	if err == nil {
		if e.config.Workspace.RestoreSession == true {
			for i := 0; i < 20; i++ {
				path := filepath.Join(home, ".gonvim", "sessions", strconv.Itoa(i)+".vim")
				_, err := os.Stat(path)
				if err != nil {
					break
				}
				sessionExists = true
				ws, err := newWorkspace(path)
				if err != nil {
					break
				}
				e.workspaces = append(e.workspaces, ws)
			}
		}
	}
	if !sessionExists {
		ws, err := newWorkspace("")
		if err != nil {
			return
		}
		e.workspaces = append(e.workspaces, ws)
	}

	activityWidget := widgets.NewQWidget(nil, 0)
	activity := newActivity()
	activity.widget = activityWidget
	activityWidget.SetLayout(activity.layout)
	e.activity = activity
	e.activity.sideArea.AddWidget(e.wsSide.scrollarea)
	e.activity.sideArea.SetCurrentWidget(e.wsSide.scrollarea)

	// layout.AddWidget(e.activity.sideArea, 0, 0)

	splitter := widgets.NewQSplitter2(core.Qt__Horizontal, nil)
	splitter.AddWidget(e.activity.sideArea)
	splitter.AddWidget(e.wsWidget)
	splitter.SetSizes([]int{editor.config.SideBar.Width, editor.width - editor.config.SideBar.Width})
	splitter.SetStretchFactor(1, 100)
	splitter.SetObjectName("splitter")
	e.splitter = splitter

	layout.AddWidget(splitter, 1, 0)
	layout.AddWidget(e.activity.widget, 0, 0)

	e.workspaceUpdate()
	e.notifyStartPos = core.NewQPoint2(e.width-400-10, e.height-30)
	var notifications []*Notification
	e.notifications = notifications
	e.window.ConnectResizeEvent(func(*gui.QResizeEvent) {
		e.width = e.window.Width()
		e.height = e.window.Height()
		e.notifyStartPos = core.NewQPoint2(e.width-400-10, e.height-30)
		x := e.notifyStartPos.X()
		y := e.notifyStartPos.Y()
		var newNotifications []*Notification
		for _, item := range e.notifications {
			x = e.notifyStartPos.X()
			y = e.notifyStartPos.Y() - item.widget.Height() - 4
			item.widget.Move2(x, y)
			e.notifyStartPos = core.NewQPoint2(x, y)
			newNotifications = append(newNotifications, item)
		}
		e.notifications = newNotifications
	})

	// Drop shadow to Side Bar
	if e.config.SideBar.DropShadow == true {
		go func() {
			shadow := widgets.NewQGraphicsDropShadowEffect(nil)
			shadow.SetBlurRadius(60)
			shadow.SetColor(gui.NewQColor3(0, 0, 0, 35))
			shadow.SetOffset3(6, 2)
			e.activity.sideArea.SetGraphicsEffect(shadow)
		}()
	}

	// Drop shadow for Activity Bar
	if e.config.ActivityBar.DropShadow == true {
		go func() {
			shadow := widgets.NewQGraphicsDropShadowEffect(nil)
			shadow.SetBlurRadius(60)
			shadow.SetColor(gui.NewQColor3(0, 0, 0, 35))
			shadow.SetOffset3(6, 2)
			e.activity.widget.SetGraphicsEffect(shadow)
		}()
	}
	//

	if e.config.ActivityBar.Visible == false {
		e.activity.widget.Hide()
	}
	if e.config.SideBar.Visible == false {
		e.activity.sideArea.Hide()
	}

	e.wsWidget.ConnectResizeEvent(func(event *gui.QResizeEvent) {
		for _, ws := range e.workspaces {
			ws.updateSize()
		}
	})

	// for macos, open file via Finder
	var macosArg string
	if runtime.GOOS == "darwin" {
		e.app.ConnectEvent(func(event *core.QEvent) bool {
			switch event.Type() {
			case core.QEvent__FileOpen:
				fileOpenEvent := gui.NewQFileOpenEventFromPointer(event.Pointer())
				macosArg = fileOpenEvent.File()
				go e.workspaces[e.active].nvim.Command(fmt.Sprintf(":tabe %s", macosArg))
			}
			return true
		})
	}

	e.window.SetCentralWidget(widget)

	go func() {
		<-editor.stop
		if runtime.GOOS == "darwin" {
			e.app.DisconnectEvent()
		}
		e.app.Quit()
	}()

	// // notification test
	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	e.pushNotification(NotifyInfo, "hoge hoge!")

	// 	time.Sleep(2 * time.Second)
	// 	e.pushNotification(NotifyInfo, "hoge hoge!")
	// 	opts0 := []*NotifyButton{}
	// 	optArg0 := &NotifyButton{
	// 		action: func() {
	// 			fmt.Println("Yes")
	// 		},
	// 		text: "yes!",
	// 	}
	// 	opts0 = append(opts0, optArg0)
	// 	e.pushNotification(NotifyWarn, "vim !", notifyOptionArg(opts0))

	// 	time.Sleep(2 * time.Second)
	// 	opts := []*NotifyButton{}
	// 	optArg := &NotifyButton{
	// 		action: func() {
	// 			fmt.Println("pushed!")
	// 		},
	// 		text: "push!",
	// 	}
	// 	opts = append(opts, optArg)

	// 	optArg2 := &NotifyButton{
	// 		action: func() {
	// 			fmt.Println("popped!")
	// 		},
	// 		text: "pop!",
	// 	}
	// 	opts = append(opts, optArg2)
	// 	e.pushNotification(NotifyWarn, "hoge hoge fuga fuga !", notifyOptionArg(opts))

	// 	time.Sleep(6 * time.Second)
	// 	go e.showNotifications()
	// 	// time.Sleep(4 * time.Second)
	// 	// go e.hideNotifications()
	// }()

	e.window.Show()
	e.wsWidget.SetFocus2()
	widgets.QApplication_Exec()
}

func (e *Editor) pushNotification(level NotifyLevel, p int, message string, opt ...NotifyOptionArg) {
	opts := NotifyOptions{}
	for _, o := range opt {
		o(&opts)
	}
	n := &Notify{
		level:   level,
		period:  p,
		message: message,
		buttons: opts.buttons,
	}
	e.notify <- n
	e.signal.NotifySignal()
}

func (e *Editor) popupNotification(level NotifyLevel, p int, message string, opt ...NotifyOptionArg) {
	notification := newNotification(level, p, message, opt...)
	notification.widget.SetParent(e.window)
	notification.widget.AdjustSize()
	x := e.notifyStartPos.X()
	y := e.notifyStartPos.Y() - notification.widget.Height() - 4
	notification.widget.Move2(x, y)
	e.notifyStartPos = core.NewQPoint2(x, y)
	e.notifications = append(e.notifications, notification)
	notification.show()
}

func hexToRGBA(hex string) *RGBA {
	format := "#%02x%02x%02x"
	if len(hex) == 4 {
		format = "#%1x%1x%1x"
	}
	var r, g, b uint8
	n, err := fmt.Sscanf(hex, format, &r, &g, &b)
	if err != nil {
		return nil
	}
	if n != 3 {
		return nil
	}
	rgba := &RGBA{
		R: (int)(r),
		G: (int)(g),
		B: (int)(b),
		A: 1,
	}

	return rgba
}

func darkenHex(hex string) string {
	c := hexToRGBA(hex)
	d := shiftColor(c, 20)
	return fmt.Sprintf("#%02x%02x%02x", (int)(d.R*255.0), (int)(d.G*255.0), (int)(d.B*255.0))
}

func shiftHex(hex string, v int) string {
	c := hexToRGBA(hex)
	d := shiftColor(c, v)
	return fmt.Sprintf("#%02x%02x%02x", (int)(d.R*255.0), (int)(d.G*255.0), (int)(d.B*255.0))
}

func isFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (e *Editor) copyClipBoard() {
	go func() {
		var yankedText string
		yankedText, _ = e.workspaces[e.active].nvim.CommandOutput("echo getreg()")
		if yankedText != "" {
			clipb.WriteAll(yankedText)
		}
	}()

}

func (e *Editor) workspaceNew() {
	ws, err := newWorkspace("")
	if err != nil {
		return
	}

	//e.active++
	//e.workspaces = append(e.workspaces, nil)
	//copy(e.workspaces[e.active+1:], e.workspaces[e.active:])

	e.workspaces = append(e.workspaces, nil)
	e.active = len(e.workspaces) - 1

	e.workspaces[e.active] = ws
	e.workspaceUpdate()
}

func (e *Editor) workspaceSwitch(index int) {
	index--
	if index < 0 || index >= len(e.workspaces) {
		return
	}
	e.active = index
	e.workspaceUpdate()
}

func (e *Editor) workspaceNext() {
	e.active++
	if e.active >= len(e.workspaces) {
		e.active = 0
	}
	e.workspaceUpdate()
}

func (e *Editor) workspacePrevious() {
	e.active--
	if e.active < 0 {
		e.active = len(e.workspaces) - 1
	}
	e.workspaceUpdate()
}

func (e *Editor) workspaceUpdate() {

	for i, ws := range e.workspaces {
		if i == e.active {
			ws.hide()
			ws.show()
		} else {
			ws.hide()
		}
	}

	for i := 0; i < len(e.wsSide.items) && i < len(e.workspaces); i++ {
		if i == e.active {
			e.wsSide.items[i].setActive()
			//e.wsSide.title.Show()
			//}
		} else {
			e.wsSide.items[i].setInactive()
		}
		//e.wsSide.scrollarea.Show()
		e.wsSide.items[i].setText(e.workspaces[i].cwdlabel)
		e.wsSide.items[i].show()
	}

	for i := len(e.workspaces); i < len(e.wsSide.items); i++ {
		e.wsSide.items[i].hide()
	}

	//if len(e.workspaces) == 1 || len(e.wsSide.items) == 1 {
	//	if e.config.showSide == false {
	//		//e.wsSide.scrollarea.Hide()
	//		e.wsSide.items[0].hide()
	//		//e.wsSide.title.Hide()
	//	} else {
	//		//e.wsSide.scrollarea.Show()
	//		//e.wsSide.title.Show()
	//		e.wsSide.items[0].setActive()
	//		e.wsSide.items[0].show()
	//	}
	//}

}

func (e *Editor) keyPress(event *gui.QKeyEvent) {
	input := e.convertKey(event.Text(), event.Key(), event.Modifiers())
	if input != "" {
		e.workspaces[e.active].nvim.Input(input)
		e.workspaces[e.active].detectTerminalMode()
	}
}

func (e *Editor) convertKey(text string, key int, mod core.Qt__KeyboardModifier) string {
	if mod&core.Qt__KeypadModifier > 0 {
		switch core.Qt__Key(key) {
		case core.Qt__Key_Home:
			return fmt.Sprintf("<%sHome>", e.modPrefix(mod))
		case core.Qt__Key_End:
			return fmt.Sprintf("<%sEnd>", e.modPrefix(mod))
		case core.Qt__Key_PageUp:
			return fmt.Sprintf("<%sPageUp>", e.modPrefix(mod))
		case core.Qt__Key_PageDown:
			return fmt.Sprintf("<%sPageDown>", e.modPrefix(mod))
		case core.Qt__Key_Plus:
			return fmt.Sprintf("<%sPlus>", e.modPrefix(mod))
		case core.Qt__Key_Minus:
			return fmt.Sprintf("<%sMinus>", e.modPrefix(mod))
		case core.Qt__Key_multiply:
			return fmt.Sprintf("<%sMultiply>", e.modPrefix(mod))
		case core.Qt__Key_division:
			return fmt.Sprintf("<%sDivide>", e.modPrefix(mod))
		case core.Qt__Key_Enter:
			return fmt.Sprintf("<%sEnter>", e.modPrefix(mod))
		case core.Qt__Key_Period:
			return fmt.Sprintf("<%sPoint>", e.modPrefix(mod))
		case core.Qt__Key_0:
			return fmt.Sprintf("<%s0>", e.modPrefix(mod))
		case core.Qt__Key_1:
			return fmt.Sprintf("<%s1>", e.modPrefix(mod))
		case core.Qt__Key_2:
			return fmt.Sprintf("<%s2>", e.modPrefix(mod))
		case core.Qt__Key_3:
			return fmt.Sprintf("<%s3>", e.modPrefix(mod))
		case core.Qt__Key_4:
			return fmt.Sprintf("<%s4>", e.modPrefix(mod))
		case core.Qt__Key_5:
			return fmt.Sprintf("<%s5>", e.modPrefix(mod))
		case core.Qt__Key_6:
			return fmt.Sprintf("<%s6>", e.modPrefix(mod))
		case core.Qt__Key_7:
			return fmt.Sprintf("<%s7>", e.modPrefix(mod))
		case core.Qt__Key_8:
			return fmt.Sprintf("<%s8>", e.modPrefix(mod))
		case core.Qt__Key_9:
			return fmt.Sprintf("<%s9>", e.modPrefix(mod))
		}
	}

	if text == "<" {
		return "<lt>"
	}

	specialKey, ok := e.specialKeys[core.Qt__Key(key)]
	if ok {
		return fmt.Sprintf("<%s%s>", e.modPrefix(mod), specialKey)
	}

	if text == "\\" {
		return fmt.Sprintf("<%s%s>", e.modPrefix(mod), "Bslash")
	}

	c := ""
	if mod&e.controlModifier > 0 || mod&e.cmdModifier > 0 {
		if int(e.keyControl) == key || int(e.keyCmd) == key || int(e.keyAlt) == key || int(e.keyShift) == key {
			return ""
		}
		c = string(key)
		if !(mod&e.shiftModifier > 0) {
			c = strings.ToLower(c)
		}
	} else {
		c = text
	}

	if c == "" {
		return ""
	}

	char := core.NewQChar11(c)
	if char.Unicode() < 0x100 && !char.IsNumber() && char.IsPrint() {
		mod &= ^e.shiftModifier
	}

	prefix := e.modPrefix(mod)
	if prefix != "" {
		return fmt.Sprintf("<%s%s>", prefix, c)
	}

	return c
}

func (e *Editor) modPrefix(mod core.Qt__KeyboardModifier) string {
	prefix := ""
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if mod&e.cmdModifier > 0 {
			prefix += "D-"
		}
	}

	if mod&e.controlModifier > 0 {
		prefix += "C-"
	}

	if mod&e.shiftModifier > 0 {
		prefix += "S-"
	}

	if mod&e.altModifier > 0 {
		prefix += "A-"
	}

	return prefix
}

func (e *Editor) initSpecialKeys() {
	e.specialKeys = map[core.Qt__Key]string{}
	e.specialKeys[core.Qt__Key_Up] = "Up"
	e.specialKeys[core.Qt__Key_Down] = "Down"
	e.specialKeys[core.Qt__Key_Left] = "Left"
	e.specialKeys[core.Qt__Key_Right] = "Right"

	e.specialKeys[core.Qt__Key_F1] = "F1"
	e.specialKeys[core.Qt__Key_F2] = "F2"
	e.specialKeys[core.Qt__Key_F3] = "F3"
	e.specialKeys[core.Qt__Key_F4] = "F4"
	e.specialKeys[core.Qt__Key_F5] = "F5"
	e.specialKeys[core.Qt__Key_F6] = "F6"
	e.specialKeys[core.Qt__Key_F7] = "F7"
	e.specialKeys[core.Qt__Key_F8] = "F8"
	e.specialKeys[core.Qt__Key_F9] = "F9"
	e.specialKeys[core.Qt__Key_F10] = "F10"
	e.specialKeys[core.Qt__Key_F11] = "F11"
	e.specialKeys[core.Qt__Key_F12] = "F12"
	e.specialKeys[core.Qt__Key_F13] = "F13"
	e.specialKeys[core.Qt__Key_F14] = "F14"
	e.specialKeys[core.Qt__Key_F15] = "F15"
	e.specialKeys[core.Qt__Key_F16] = "F16"
	e.specialKeys[core.Qt__Key_F17] = "F17"
	e.specialKeys[core.Qt__Key_F18] = "F18"
	e.specialKeys[core.Qt__Key_F19] = "F19"
	e.specialKeys[core.Qt__Key_F20] = "F20"
	e.specialKeys[core.Qt__Key_F21] = "F21"
	e.specialKeys[core.Qt__Key_F22] = "F22"
	e.specialKeys[core.Qt__Key_F23] = "F23"
	e.specialKeys[core.Qt__Key_F24] = "F24"
	e.specialKeys[core.Qt__Key_Backspace] = "BS"
	e.specialKeys[core.Qt__Key_Delete] = "Del"
	e.specialKeys[core.Qt__Key_Insert] = "Insert"
	e.specialKeys[core.Qt__Key_Home] = "Home"
	e.specialKeys[core.Qt__Key_End] = "End"
	e.specialKeys[core.Qt__Key_PageUp] = "PageUp"
	e.specialKeys[core.Qt__Key_PageDown] = "PageDown"

	e.specialKeys[core.Qt__Key_Return] = "Enter"
	e.specialKeys[core.Qt__Key_Enter] = "Enter"
	e.specialKeys[core.Qt__Key_Tab] = "Tab"
	e.specialKeys[core.Qt__Key_Backtab] = "Tab"
	e.specialKeys[core.Qt__Key_Escape] = "Esc"

	e.specialKeys[core.Qt__Key_Backslash] = "Bslash"
	e.specialKeys[core.Qt__Key_Space] = "Space"

	goos := runtime.GOOS
	e.shiftModifier = core.Qt__ShiftModifier
	e.altModifier = core.Qt__AltModifier
	e.keyAlt = core.Qt__Key_Alt
	e.keyShift = core.Qt__Key_Shift
	if goos == "darwin" {
		e.controlModifier = core.Qt__MetaModifier
		e.cmdModifier = core.Qt__ControlModifier
		e.metaModifier = core.Qt__AltModifier
		e.keyControl = core.Qt__Key_Meta
		e.keyCmd = core.Qt__Key_Control
	} else {
		e.controlModifier = core.Qt__ControlModifier
		e.metaModifier = core.Qt__MetaModifier
		e.keyControl = core.Qt__Key_Control
		if goos == "linux" {
			e.cmdModifier = core.Qt__MetaModifier
			e.keyCmd = core.Qt__Key_Meta
		}
	}
}

func (e *Editor) close() {
	e.stopOnce.Do(func() {
		close(e.stop)
	})
}

func (e *Editor) cleanup() {
	home, err := homedir.Dir()
	if err != nil {
		return
	}
	sessions := filepath.Join(home, ".gonvim", "sessions")
	os.RemoveAll(sessions)
	os.MkdirAll(sessions, 0755)

	select {
	case <-e.stop:
		return
	default:
	}

	for i, ws := range e.workspaces {
		sessionPath := filepath.Join(sessions, strconv.Itoa(i)+".vim")
		fmt.Println(sessionPath)
		fmt.Println(ws.nvim.Command("mksession " + sessionPath))
		fmt.Println("mksession finished")
	}
}
