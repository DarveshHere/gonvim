package lsp

import (
	"fmt"

	"github.com/neovim/go-client/nvim"
)


// Lsp is
type Lsp struct {
	nvim          *nvim.Nvim
	filetype      string
}


func Register(nvim *nvim.Nvim) {
	nvim.Subscribe("GonvimLsp")
	lsp := &Lsp{
		nvim: nvim,
	}
	nvim.RegisterHandler("GonvimLsp", func(args ...interface{}) {
		go lsp.handle(args...)
	})
}


func (l *Lsp) handle(args ...interface{}) {
	if len(args) < 1 {
		return
	}
	request, ok := args[0].(string)
	if !ok {
		return
	}
	switch request {
	case "textdocument_hover":
		l.hover()
	default:
		fmt.Println("unhandleld lsp request", request)
	}
}

func (l *Lsp) hover() {
	if l.filetype == "" {
		buf, err := l.nvim.CurrentBuffer()
		if err != nil {
			return
		}
		err = l.nvim.BufferOption(buf, "filetype", &(l.filetype))
	}
	var result int
	_ = l.nvim.ExecuteLua(`
-- def callback
local callback_func = function(success, data)
  local util = require('nvim.util')
  local idx = 0
  local long_string = ''

  if util.is_array(data.contents) == true then
    for _, item in ipairs(data.contents) do
      local value
      if type(item) == 'table' then
        value = item.value
      elseif item == nil then
        value = ''
      else
        value = item
      end

      long_string = long_string .. value .. "\n"
    end

  elseif type(data.contents) == 'table' then
    long_string = long_string .. (data.contents.value or '')
  else
    long_string = data.contents
  end

  if long_string == '' then
    long_string = 'LSP: No information available'
  end

  local notify = "call rpcnotify(0, 'Gui', 'hover_show', " .. "'" .. long_string .. "'"  .. ", [(winline() <= 2) ? 3 : -1, 1] ," .. idx .. ")"
  vim.api.nvim_command(notify)
end

-- request to lsp
vim.lsp.client.request_async('textDocument/hover', {}, callback_func, '` + l.filetype + `')`, &result)
}



