" =========================
" ~/.vimrc
"
" =========================

set nocompatible
" --- Basic ---
" --- ALE ---
let g:go_highlight_fields = 1
let g:go_highlight_function_calls = 1
let g:ale_lint_delay = 0
let g:go_highlight_functions = 1
let g:go_highlight_operators = 1
let g:go_highlight_types = 1
syntax on

" --- Improved Command-line Completion ---
set wildmenu
set wildmode=longest:full,full

let g:ale_lint_delay = 0
set ignorecase
set smartcase


set signcolumn=yes
set number
set autoindent
set smartindent
set expandtab
set tabstop=4
set shiftwidth=4
set softtabstop=4
set noshowmode
" --- Syntax highlighting ---
set background=dark
set termguicolors
" --- Cursor ---
set guicursor=a:block
" --- jj = Escape ---
inoremap jj <Esc>

" --- ; = : in Normal mode ---
nnoremap ; :
"random shit rebinds
let mapleader = " "

nnoremap <leader>w :w<CR>
nnoremap <leader>q :q<CR>
nnoremap <leader>x :x<CR>
nnoremap <leader>v <C-V>
nnoremap <leader>a ma
nnoremap <leader><leader> 'a
" --- Persistent undo ---
set undofile
set undodir=~/.vim/undo
set history=1000

" --- Persistent command/search history ---
set viminfo='1000,<1000,s100,h

" --- Make sure undo directory exists ---
if !isdirectory(expand('~/.vim/undo'))
    call mkdir(expand('~/.vim/undo'), 'p')
endif

" --- Search highlighting ---
set hlsearch
set incsearch

" --- Better indentation ---
filetype plugin indent on

" --- Show matching brackets ---
set showmatch

" --- Don't wrap code ---
set nowrap
" Enable line highlighting and relative numbering

" Customize the current line number color
" guifg changes text color, guibg changes the number's background color
highlight CursorLineNr ctermfg=Yellow gui=bold


" Error line highlighting
" nnoremap  <C-L> : silent! cclose<Bar> :silent! lclose<Bar> :silent! nohlsearch<CR>
nnoremap <silent> <C-L> :silent! cclose<Bar>silent! lclose<Bar>silent! nohlsearch<CR>
" For the currently selected/active search term
hi CurSearch ctermbg=Green ctermfg=Black guibg=Green guifg=Black
set timeout
set timeoutlen=150
set makeprg=go\ build\ .
set shellcmdflag=-c
autocmd BufRead,BufNewFile *.vala set filetype=c

" --- Plugin management ---
let g:lsp_semantic_enabled = 1
let g:lsp_semantic_delay = 0

call plug#begin('~/.vim2/plugged')

Plug 'rose-pine/vim'
Plug 'dense-analysis/ale'
Plug 'luochen1990/rainbow'
Plug 'jiangmiao/auto-pairs'
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
Plug 'prabirshrestha/asyncomplete.vim'
Plug 'tpope/vim-commentary'
Plug 'voldikss/vim-floaterm'
call plug#end()


" Keep vim-go's syntax, but let ALE own Go diagnostics and formatting.
let g:go_gopls_enabled = 0
let g:go_code_completion_enabled = 0
let g:go_def_mode = 'gopls'
let g:go_info_mode = 'gopls'
let g:go_diagnostics_level = 0
let g:go_highlight_diagnostic_errors = 0
let g:go_fmt_autosave = 0
let g:go_fmt_command = 'goimports'
" --- Rose Pine color scheme ---
colorscheme rosepine

" Match Go syntax colors to the Rose Pine palette.
highlight goStatement guifg=#c4a7e7 ctermfg=Magenta gui=bold cterm=bold
highlight goConditional guifg=#c4a7e7 ctermfg=Magenta gui=bold cterm=bold
highlight goRepeat guifg=#c4a7e7 ctermfg=Magenta gui=bold cterm=bold
highlight goLabel guifg=#c4a7e7 ctermfg=Magenta
highlight goDeclaration guifg=#c4a7e7 ctermfg=Magenta gui=bold cterm=bold
highlight goTypeDecl guifg=#c4a7e7 ctermfg=Magenta gui=bold cterm=bold
highlight goFunction guifg=#9ccfd8 ctermfg=Cyan gui=bold cterm=bold
highlight goFunctionCall guifg=#9ccfd8 ctermfg=Cyan gui=bold cterm=bold
highlight goBuiltins guifg=#9ccfd8 ctermfg=Cyan gui=bold cterm=bold
highlight goField guifg=#ebbcba ctermfg=LightRed
highlight goParamName guifg=#ebbcba ctermfg=LightRed
highlight goType guifg=#31748f ctermfg=Blue gui=bold cterm=bold
highlight goTypeName guifg=#31748f ctermfg=Blue gui=bold cterm=bold
highlight goTypeConstructor guifg=#31748f ctermfg=Blue gui=bold cterm=bold
highlight goString guifg=#f6c177 ctermfg=Yellow
highlight goRawString guifg=#f6c177 ctermfg=Yellow
highlight goComment guifg=#6e6a86 ctermfg=DarkGray
highlight Delimiter guifg=#908caa ctermfg=LightGray
highlight goOperator guifg=#c4a7e7 ctermfg=Magenta

" Keep inactive line numbers muted and mark the active row clearly.
set cursorline
highlight LineNr guifg=#6e6a86 ctermfg=DarkGray
highlight CursorLine guibg=#26233a ctermbg=Black
highlight CursorLineNr guifg=#ebbcba ctermfg=LightRed

" --- Keep terminal background color unchanged ---
" Preserve the terminal background instead of the theme's default background.
function! NormaliseBackground()
    highlight Normal ctermbg=NONE guibg=NONE
    highlight NonText ctermbg=NONE guibg=NONE
    highlight LineNr ctermbg=NONE guibg=NONE
    highlight SignColumn ctermbg=NONE guibg=NONE
    highlight EndOfBuffer ctermbg=NONE guibg=NONE
endfunction

autocmd ColorScheme rosepine call NormaliseBackground()
" Run it immediately for the current session
call NormaliseBackground()
" ALE — Go
let g:ale_linters = {'go': ['gopls']}

let g:ale_fixers = {'go': ['gofmt', 'goimports']}

let g:ale_fix_on_save = 1

let g:ale_lint_on_text_changed = 'always'
let g:ale_lint_on_insert_leave = 1
let g:ale_lint_on_enter = 1
let g:ale_lint_delay = 0
let g:ale_echo_cursor = 1

" Highlight ONLY the offending source text
let g:ale_set_highlights = 1
let g:ale_set_signs = 1
let g:ale_virtualtext_cursor = 2

" highlight ALEError guifg=#191724 guibg=#eb6f92
" highlight ALEWarning guifg=#191724 guibg=#f6c177
" 
" highlight ALEError ctermfg=Black ctermbg=Red
" highlight ALEWarning ctermfg=Black ctermbg=Yellow
" 1. Enable the rainbow plugin globally

"highlight ALEError guifg=#eb6f92 guibg=#2d202a gui=bold
"highlight ALEWarning guifg=#f6c177 guibg=#2e2824 gui=bold
"
"highlight ALEError ctermfg=Red ctermbg=NONE cterm=underline
"highlight ALEWarning ctermfg=Yellow ctermbg=NONE cterm=underline

"highlight ALEError guifg=NONE guibg=NONE gui=undercurl guisp=#eb6f92
"highlight ALEWarning guifg=NONE guibg=NONE gui=undercurl guisp=#f6c177
"
"highlight ALEError ctermfg=NONE ctermbg=NONE cterm=underline
"highlight ALEWarning ctermfg=NONE ctermbg=NONE cterm=underline
"

highlight ALEError guifg=#eb6f92 guibg=NONE gui=italic
highlight ALEError guifg=#eb6f92 guibg=NONE gui=italic
highlight ALEWarning guifg=#f6c177 guibg=NONE gui=italic

highlight ALEError ctermfg=Red ctermbg=NONE cterm=italic
highlight ALEWarning ctermfg=Yellow ctermbg=NONE cterm=italic

function! s:ExpandALEWordHighlights() abort
    if !exists('b:ale_highlight_items')
        return
    endif

    let l:items = deepcopy(b:ale_highlight_items)

    for l:item in l:items
        if l:item.lnum != get(l:item, 'end_lnum', l:item.lnum)
            continue
        endif

        let l:text = getline(l:item.lnum)
        let l:start = l:item.col - 1
        let l:word_start = match(l:text, '\k\+')

        while l:word_start >= 0
            let l:word_end = matchend(l:text, '\k\+', l:word_start)

            if l:word_start <= l:start && l:start < l:word_end
                let l:item.col = l:word_start + 1
                let l:item.end_col = l:word_end
                break
            endif

            let l:word_start = match(l:text, '\k\+', l:word_end)
        endwhile
    endfor

    let b:ale_highlight_items = l:items
    call ale#highlight#UpdateHighlights()
endfunction

augroup ALEWordHighlights
    autocmd!
    autocmd User ALELintPost call <SID>ExpandALEWordHighlights()
augroup END

let g:lsp_semantic_enabled = 0
let g:lsp_semantic_delay = 0
autocmd BufWinEnter * RainbowToggleOn

let g:ale_lint_on_text_changed = 'always'
let g:ale_lint_delay = 0


let g:rainbow_active = 1

let g:rainbow_conf = {'guifgs': ['#c4a7e7', '#9ccfd8', '#ebbcba', '#f6c177', '#eb6f92', '#31748f'], 'ctermfgs': ['Magenta', 'Cyan', 'LightRed', 'Yellow', 'Red', 'Blue'], 'parentheses': ['start=/(/ end=/)/ fold', 'start=/\[/ end=/\]/ fold', 'start=/{/ end=/}/ fold']}


" 5. Clean up auto-pairs conflict by forcing a syntax refresh on file load
autocmd BufWinEnter,Syntax * RainbowToggleOn







augroup remember_cursor_position
    autocmd!
    autocmd BufReadPost * if line("'\"") > 0 && line("'\"") <= line("$") | execute "normal! g`\"" | endif
augroup END
highlight LspSemanticVariable guifg=#ebbcba ctermfg=LightRed
highlight LspSemanticProperty guifg=#ebbcba ctermfg=LightRed
highlight LspSemanticParameter guifg=#ebbcba ctermfg=LightRed
highlight LspSemanticFunction guifg=#9ccfd8 ctermfg=Cyan gui=bold cterm=bold
highlight LspSemanticMethod guifg=#9ccfd8 ctermfg=Cyan gui=bold cterm=bold
highlight LspSemanticType guifg=#31748f ctermfg=Blue gui=bold cterm=bold
highlight LspSemanticString guifg=#f6c177 ctermfg=Yellow
highlight LspSemanticNumber guifg=#c4a7e7 ctermfg=Magenta

set completeopt=menuone,noselect,noinsert

let g:lsp_document_highlight_enabled = 1
let g:lsp_document_highlight_delay = 0
" Completion menu
" Completion documentation popup
set completepopup=align:item,border:off,highlight:NormalFloat

" Popup appearance
highlight NormalFloat guibg=#1f1d2e guifg=#e0def4
highlight Pmenu      guibg=#1f1d2e guifg=#e0def4
highlight PmenuSel   guibg=#31748f guifg=#e0def4 gui=bold

" Slight transparency
set pumheight=3
highlight PmenuSbar guibg=#26233a
highlight PmenuThumb guibg=#9ccfd8
let g:lsp_completion_documentation_enabled = 0

" --- Ctrl+j / Ctrl+k Completion Controls ---
" Ctrl+j to go down the list, Ctrl+k to go up
inoremap <expr> <C-j> pumvisible() ? "\<C-n>" : "\<C-j>"
inoremap <expr> <C-k> pumvisible() ? "\<C-p>" : "\<C-k>"

" Enter selects the item and closes the menu cleanly
inoremap <expr> <CR>  pumvisible() ? "\<C-y>" : "\<CR>"
let g:ale_completion_enabled = 1
" Automatically show pmenu, insert the match inline, and don't auto-select

" Automatically trigger Go omni-completion as you type
autocmd FileType go autocmd TextChangedI <buffer> if !pumvisible() | call feedkeys("\<C-x>\<C-o>", 'n') | endif

" Use Tab to cycle down, Shift-Tab to cycle up the pmenu
" inoremap <expr> <Tab> pumvisible() ? "\<C-n>" : "\<Tab>"
inoremap <expr> <S-Tab> pumvisible() ? "\<C-p>" : "\<S-Tab>"

" run go install golang.org/x/tools/cmd/goimports@latest


" run curl -fLo ~/.vim/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
" go install golang.org/x/tools/gopls@latest
" go install golang.org/x/tools/cmd/goimports@latest
" " 1. Force Vim to look for filetype plugins
filetype plugin on

" 2. Manually assign Go's built-in completion if Vim missed it
autocmd FileType go setlocal omnifunc=go#complete#Complete

" 3. Smart Tab: If pmenu is open, cycle down.
"    If omnifunc is set, trigger it.
"    Otherwise, fall back to standard keyword completion (Ctrl+N).



"inoremap <expr> <Tab> pumvisible() ? "\<C-n>" :
"      \ (&omnifunc != "") ? "\<C-x>\<C-o>" : "\<Tab>"
"
" --- Smart Tab: Cycle pmenu if visible, otherwise insert a normal Tab ---
inoremap <expr> <Tab> pumvisible() ? "\<C-n>" : "\<Tab>"
" 4. Shift-Tab to cycle backward
inoremap <expr> <S-Tab> pumvisible() ? "\<C-p>" : "\<S-Tab>"




" --- Jump to errors instantly ---
nnoremap <leader>e :ALENextWrap<CR>
nnoremap <leader>E :ALEPreviousWrap<CR>

" --- Highlight modified text lines natively ---
" This enables highlight for changed text within the current buffer session
highlight ModifiedLineNr ctermfg=Yellow guifg=#f6c177 gui=bold

" You can use Vim's built-in diff/change tracking or jump to changes with:
" g; - go to older position in change list
" g, - go to newer position in change list
" 


" open a terminal in vim
" --- Toggle floating terminal with <space>t ---
nnoremap <silent> <leader>t :FloatermToggle<CR>
tnoremap <silent> <leader>t <C-\><C-n>:FloatermToggle<CR>
j

" --- Floaterm Customization Examples ---
let g:floaterm_width = 0.7    
let g:floaterm_height =0.7     
let g:floaterm_position = 'topright' " Can also be 'top', 'bottom', 'left', 'right', 'topleft', etc.
let g:floaterm_borderchars = '─│─│╭╮╯╰' " Custom border style if you want to tweak the look
" --- Style Floaterm borders to match transparency ---
highlight Floaterm guibg=NONE ctermbg=NONE
highlight FloatermBorder guifg=#9ccfd8 guibg=NONE ctermfg=Green ctermbg=NONE
highlight NormalFloat guibg=NONE ctermbg=NONE

 
