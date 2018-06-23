package editor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/svg"
	"github.com/therecipe/qt/widgets"
)

type DeinSide struct {
	widget       *widgets.QWidget
	layout       *widgets.QLayout
	title        *widgets.QLabel
	scrollarea   *widgets.QScrollArea
	searchbox    *widgets.QLineEdit
	searchresult *widgets.QWidget
}

func newDeinSide() *DeinSide {
	w := editor.workspaces[editor.active]

	layout := newHFlowLayout(0, 0, 0, 0, 20)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)
	header := widgets.NewQLabel(nil, 0)
	header.SetContentsMargins(20, 15, 20, 10)
	header.SetText("Dein.vim")
	widget := widgets.NewQWidget(nil, 0)
	widget.SetContentsMargins(0, 0, 0, 100)
	widget.SetLayout(layout)

	searchBoxLayout := widgets.NewQHBoxLayout()
	searchBoxLayout.SetContentsMargins(20, 10, 20, 10)
	searchBoxLayout.SetSpacing(0)
	searchbox := widgets.NewQLineEdit(nil)
	searchBoxLayout.AddWidget(searchbox, 0, 0)
	searchBoxWidget := widgets.NewQWidget(nil, 0)
	searchBoxWidget.SetLayout(searchBoxLayout)

	searchbox.ConnectReturnPressed(doPluginSearch)

	searchresult := widgets.NewQWidget(nil, 0)

	side := &DeinSide{
		widget:       widget,
		layout:       layout,
		title:        header,
		searchbox:    searchbox,
		searchresult: searchresult,
	}

	layout.AddWidget(header)
	layout.AddWidget(searchBoxWidget)

	var userPath, defaultPath string
	userPath, _ = w.nvim.CommandOutput("echo g:dein#cache_directory")
	defaultPath, _ = w.nvim.CommandOutput("echo g:dein#_base_path")
	var deinDirectInstallPath string
	if userPath == "" {
		deinDirectInstallPath = defaultPath
	} else {
		deinDirectInstallPath = userPath
	}

	line := widgets.NewQLabel(nil, 0)
	line.SetText("Path: " + deinDirectInstallPath)
	layout.AddWidget(line)
	//line.Show()

	side.title.Show()
	side.searchbox.Show()

	// Set Color for DeinSide
	fg := editor.fgcolor
	bg := editor.bgcolor
	deinSideStyle := fmt.Sprintf("QWidget {	color: rgba(%d, %d, %d, 1);		border-right: 0px solid;	}", gradColor(fg).R, gradColor(fg).G, gradColor(fg).B)
	side.widget.SetStyleSheet(fmt.Sprintf(".QWidget {padding-top: 5px;	background-color: rgba(%d, %d, %d, 1);	}	", shiftColor(bg, -5).R, shiftColor(bg, -5).G, shiftColor(bg, -5).B) + deinSideStyle)
	side.searchbox.SetStyleSheet(fmt.Sprintf(".QLineEdit { border: 1px solid	%s; border-radius: 1px; background: rgba(%d, %d, %d, 1); selection-background-color: rgba(%d, %d, %d, 1); }	", editor.config.accentColor, gradColor(bg).R, gradColor(bg).G, gradColor(bg).B, gradColor(bg).R, gradColor(bg).G, gradColor(bg).B) + deinSideStyle)

	return side
}

type PluginResults struct {
	TotalResults   int `json:"total_results"`
	ResultsPerPage int `json:"results_per_page"`
	TotalPages     int `json:"total_pages"`
	Plugins        []struct {
		VimorgRating             int      `json:"vimorg_rating"`
		GithubHomepage           string   `json:"github_homepage"`
		UpdatedAt                int      `json:"updated_at"`
		GithubReadmeFilename     string   `json:"github_readme_filename"`
		VimorgShortDesc          string   `json:"vimorg_short_desc"`
		GithubVimScriptsStars    int      `json:"github_vim_scripts_stars"`
		VimorgType               string   `json:"vimorg_type"`
		NormalizedName           string   `json:"normalized_name"`
		Category                 string   `json:"category"`
		Author                   string   `json:"author"`
		PluginManagerUsers       int      `json:"plugin_manager_users"`
		ShortDesc                string   `json:"short_desc"`
		VimorgAuthor             string   `json:"vimorg_author"`
		VimorgNumRaters          int      `json:"vimorg_num_raters"`
		VimorgURL                string   `json:"vimorg_url"`
		GithubVimScriptsBundles  int      `json:"github_vim_scripts_bundles"`
		GithubRepoName           string   `json:"github_repo_name"`
		Tags                     []string `json:"tags"`
		GithubStars              int      `json:"github_stars"`
		GithubVimScriptsRepoName string   `json:"github_vim_scripts_repo_name"`
		VimorgDownloads          int      `json:"vimorg_downloads"`
		GithubRepoID             string   `json:"github_repo_id"`
		Slug                     string   `json:"slug"`
		VimorgID                 string   `json:"vimorg_id"`
		GithubOwner              string   `json:"github_owner"`
		Name                     string   `json:"name"`
		CreatedAt                int      `json:"created_at"`
		GithubShortDesc          string   `json:"github_short_desc"`
		VimorgName               string   `json:"vimorg_name"`
		GithubURL                string   `json:"github_url"`
		GithubBundles            int      `json:"github_bundles"`
		Keywords                 string   `json:"keywords"`
		GithubAuthor             string   `json:"github_author"`
	} `json:"plugins"`
}

type Plugin struct {
	widget *widgets.QWidget
}

func doPluginSearch() {
	w := editor.workspaces[editor.active]
	fg := editor.fgcolor

	response, _ := http.Get("http://vimawesome.com/api/plugins?query=" + editor.deinSide.searchbox.Text())
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)
	jsonBytes := ([]byte)(data)

	var results PluginResults
	if err := json.Unmarshal(jsonBytes, &results); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return
	}

	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQBoxLayout(widgets.QBoxLayout__TopToBottom, widget)
	layout.SetContentsMargins(20, 10, 0, 150)
	layout.SetSpacing(20)

	for _, p := range results.Plugins {
		pluginWidget := widgets.NewQWidget(nil, 0)
		pluginLayout := widgets.NewQBoxLayout(widgets.QBoxLayout__TopToBottom, pluginWidget)
		pluginLayout.SetContentsMargins(2, 2, 2, 2)
		pluginLayout.SetSpacing(1)
		pluginWidget.SetFixedWidth(editor.config.sideWidth - (20 + 20 - 2))

		// * plugin name
		pluginName := widgets.NewQLabel(nil, 0)
		pluginName.SetText(p.Name)
		pluginName.SetStyleSheet(fmt.Sprintf(" .QLabel {font: bold; color: rgba(%d, %d, %d, 1);} ", fg.R, fg.G, fg.B))

		// * plugin description
		var pluginDesc *widgets.QLabel
		if p.ShortDesc != "" {
			pluginDesc = widgets.NewQLabel(nil, 0)
			pluginDesc.SetText(p.ShortDesc)
			pluginDesc.SetWordWrap(true)
		}

		// * plugin info
		pluginInfo := widgets.NewQWidget(nil, 0)
		//pluginInfoLayout := newVFlowLayout(16, 10, 1, 2, 0)
		pluginInfoLayout := widgets.NewQHBoxLayout()
		pluginInfoLayout.SetContentsMargins(0, 0, 0, 0)
		pluginInfoLayout.SetSpacing(8)

		// ** plugin stars
		pluginStars := widgets.NewQWidget(nil, 0)
		pluginStarsLayout := widgets.NewQHBoxLayout()
		pluginStarsLayout.SetContentsMargins(0, 0, 0, 0)
		pluginStarsLayout.SetSpacing(1)

		pluginStarsIcon := svg.NewQSvgWidget(nil)
		pluginStarsIcon.SetFixedSize2(11, 11)
		svgStarContent := w.getSvg("star", fg)
		pluginStarsIcon.Load2(core.NewQByteArray2(svgStarContent, len(svgStarContent)))

		pluginStarsNum := widgets.NewQLabel(nil, 0)
		pluginStarsNum.SetText(strconv.Itoa(p.GithubStars))
		pluginStarsNum.SetContentsMargins(0, 0, 0, 0)
		pluginStarsNum.SetStyleSheet(" .QLabel {font-size: 10px;} ")

		pluginStarsLayout.AddWidget(pluginStarsIcon, 0, 0)
		pluginStarsLayout.AddWidget(pluginStarsNum, 0, 0)

		pluginStars.SetLayout(pluginStarsLayout)

		// ** plugin downloadss
		pluginDownloads := widgets.NewQWidget(nil, 0)
		pluginDownloadsLayout := widgets.NewQHBoxLayout()
		pluginDownloadsLayout.SetContentsMargins(0, 0, 0, 0)
		pluginDownloadsLayout.SetSpacing(1)

		pluginDownloadsIcon := svg.NewQSvgWidget(nil)
		pluginDownloadsIcon.SetFixedSize2(11, 11)
		svgDownloadContent := w.getSvg("download", fg)
		pluginDownloadsIcon.Load2(core.NewQByteArray2(svgDownloadContent, len(svgDownloadContent)))

		pluginDownloadsNum := widgets.NewQLabel(nil, 0)
		pluginDownloadsNum.SetText(strconv.Itoa(p.PluginManagerUsers))
		pluginDownloadsNum.SetContentsMargins(0, 0, 0, 0)
		pluginDownloadsNum.SetStyleSheet(" .QLabel {font-size: 10px;} ")

		pluginDownloadsLayout.AddWidget(pluginDownloadsIcon, 0, 0)
		pluginDownloadsLayout.AddWidget(pluginDownloadsNum, 0, 0)

		pluginDownloads.SetLayout(pluginDownloadsLayout)

		// ** plugin author
		pluginAuthor := widgets.NewQWidget(nil, 0)
		pluginAuthorLayout := widgets.NewQHBoxLayout()
		pluginAuthorLayout.SetContentsMargins(0, 0, 0, 0)
		pluginAuthorLayout.SetSpacing(1)

		pluginAuthorIcon := svg.NewQSvgWidget(nil)
		pluginAuthorIcon.SetFixedSize2(11, 11)
		svgUserContent := w.getSvg("user", fg)
		pluginAuthorIcon.Load2(core.NewQByteArray2(svgUserContent, len(svgUserContent)))

		pluginAuthorNum := widgets.NewQLabel(nil, 0)
		pluginAuthorNum.SetText(p.GithubAuthor)
		pluginAuthorNum.SetContentsMargins(0, 0, 0, 0)
		pluginAuthorNum.SetStyleSheet(" .QLabel {font-size: 10px;} ")

		pluginAuthorLayout.AddWidget(pluginAuthorIcon, 0, 0)
		pluginAuthorLayout.AddWidget(pluginAuthorNum, 0, 0)

		pluginAuthor.SetLayout(pluginAuthorLayout)

		// * plugin info
		pluginStars.AdjustSize()
		pluginDownloads.AdjustSize()
		pluginAuthor.AdjustSize()
		pluginInfoLayout.AddWidget(pluginStars, 0, 0)
		pluginInfoLayout.AddWidget(pluginDownloads, 0, 0)
		pluginInfoLayout.AddWidget(pluginAuthor, 0, 0)
		pluginInfo.SetLayout(pluginInfoLayout)
		pluginInfo.SetFixedWidth(pluginStars.Width() + pluginDownloads.Width() + pluginAuthor.Width() + 3*10)

		// make widget
		pluginLayout.AddWidget(pluginName, 0, 0)
		if p.ShortDesc != "" {
			pluginLayout.AddWidget(pluginDesc, 0, 0)
		}
		pluginLayout.AddWidget(pluginInfo, 0, 0)

		// add to parent widget
		layout.AddWidget(pluginWidget, 0, 0)

		plugin := &Plugin{
			widget: pluginWidget,
		}
		plugin.widget.ConnectEnterEvent(plugin.enterWidget)
		plugin.widget.ConnectLeaveEvent(plugin.leaveWidget)

	}
	widget.AdjustSize()

	editor.deinSide.layout.RemoveWidget(editor.deinSide.searchresult)
	editor.deinSide.searchresult.DestroyQWidget()
	editor.deinSide.layout.AddWidget(widget)
	editor.deinSide.searchresult = widget
	widget.Show()

}

func (p *Plugin) enterWidget(event *core.QEvent) {
	bg := editor.bgcolor
	p.widget.SetStyleSheet(fmt.Sprintf(" .QWidget { background: rgba(%d, %d, %d, 1);} ", shiftColor(bg, -15).R, shiftColor(bg, -15).G, shiftColor(bg, -15).B))
}

func (p *Plugin) leaveWidget(event *core.QEvent) {
	bg := editor.bgcolor
	p.widget.SetStyleSheet(fmt.Sprintf(" .QWidget { background: rgba(%d, %d, %d, 1);} ", shiftColor(bg, -5).R, shiftColor(bg, -5).G, shiftColor(bg, -5).B))
}
