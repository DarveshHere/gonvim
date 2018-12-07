package editor

import (
	"fmt"
	"strings"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/svg"
	"github.com/therecipe/qt/widgets"
)

// Hover is
type Hover struct {
	ws            *Workspace
	cusor         []int
	formattedText string
	text          string
	widget        *widgets.QWidget
	label         *widgets.QLabel
	height        int
	x             int
	y             int
	margin        int
	isVisible     bool
}

func initHover() *Hover {
	widget := widgets.NewQWidget(nil, 0)

	shadow := widgets.NewQGraphicsDropShadowEffect(nil)
	shadow.SetBlurRadius(55)
	shadow.SetColor(gui.NewQColor3(0, 0, 0, 200))
	shadow.SetOffset3(-2, 4)

	widget.SetGraphicsEffect(shadow)
	layout := widgets.NewQHBoxLayout()
	layout.SetContentsMargins(0, 0, 0, 0)

	icon := svg.NewQSvgWidget(nil)
	icon.SetFixedWidth(editor.iconSize)
	icon.SetFixedHeight(editor.iconSize)
	icon.SetContentsMargins(0, 0, 0, 0)
	svgContent := editor.getSvg("func", hexToRGBA(editor.config.SideBar.AccentColor))
	icon.Load2(core.NewQByteArray2(svgContent, len(svgContent)))

	label := widgets.NewQLabel(nil, 0)
	label.SetFont(gui.NewQFont2(editor.config.Editor.FontFamily, editor.config.Editor.FontSize-1, 1, false))

	layout.AddWidget(icon, 0, 0)
	layout.AddWidget(label, 0, 0)

	layout.SetAlignment(icon, core.Qt__AlignTop)
	layout.SetAlignment(label, core.Qt__AlignTop)

	widget.SetLayout(layout)
	m := 10
	widget.SetContentsMargins(m, m, m, m)
	hover := &Hover{
		cusor:  []int{0, 0},
		widget: widget,
		label:  label,
		height: widget.SizeHint().Height(),
		margin: m,
	}
	return hover
}

func (h *Hover) showItem(args []interface{}) {
	text := args[0].(string)
	h.text = text
	cursor := args[1].([]interface{})
	h.cusor[0] = reflectToInt(cursor[0])
	h.cusor[1] = reflectToInt(cursor[1])
	h.update()
	h.move()
	h.hide()
	h.show()
}

func (h *Hover) pos(args []interface{}) {
	h.update()
}

func (h *Hover) update() {
	text := h.text
	i := strings.Index(text, "\n")
	formattedText := ""
	font := gui.NewQFontMetricsF(gui.NewQFont2(editor.config.Editor.FontFamily, editor.config.Editor.FontSize-1, 1, false))
	width := int(font.HorizontalAdvance(text[:i], -1))
	descText := strings.TrimSpace(text[i:])
	wrapDescText := ""
	margin := 4 * h.margin
	if descText != "" {
		switch {
		case (len(descText) <= len(text[:i])):
			h.widget.SetFixedWidth(width + editor.iconSize + margin)
		case (len(descText) > len(text[:i]) && len(descText) <= int(float64(len(text[:i]))*1.4)):
			width = int(font.HorizontalAdvance(descText, -1))
			h.widget.SetFixedWidth(width + editor.iconSize + margin)
		default:
			hoverMaximumWidth := int(float64(width)*1.4) + editor.iconSize + margin
			h.widget.SetFixedWidth(hoverMaximumWidth)
			wrapDescText = makeWrapLabelText(descText, int(float64(width)*1.4), font)
		}
		if wrapDescText != "" {
			formattedText = fmt.Sprintf("<p style=\"color: %s\";>%s</p>\n%s", editor.fgcolor.Hex(), text[:i], wrapDescText)
		} else {
			formattedText = fmt.Sprintf("<p style=\"color: %s\";>%s</p>\n%s", editor.fgcolor.Hex(), text[:i], descText)
		}

	} else {
		h.widget.SetFixedWidth(width + editor.iconSize + margin)
		formattedText = fmt.Sprintf("<p style=\"color: %s\";>%s</p>", editor.fgcolor.Hex(), text)
	}

	h.formattedText = formattedText
	h.label.SetText(formattedText)
}

func (h *Hover) move() {
	text := h.text
	row := h.ws.screen.cursor[0] + h.cusor[0]
	col := h.ws.screen.cursor[1] + h.cusor[1]
	i := strings.Index(text, "(")
	x := float64(col) * h.ws.font.truewidth
	if i > -1 {
		x -= h.ws.font.defaultFontMetrics.HorizontalAdvance(string(text[:i]), -1)
	}
	h.height = h.widget.MinimumSizeHint().Height()
	h.x = int(x) + editor.activity.widget.Width()
	h.y = row*h.ws.font.lineHeight - h.height
	h.widget.Move2(h.x, h.y)
}

func (h *Hover) show() {
	if h.isVisible {
		return
	}
	h.widget.Show()
	h.isVisible = true
}

func (h *Hover) hide() {
	if !h.isVisible {
		return
	}
	h.widget.Hide()
	h.isVisible = false
}
