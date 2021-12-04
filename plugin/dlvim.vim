command! Dlvim call s:run_dlvim()

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

highlight CurrentInstruction ctermbg=lightblue
sign define DlvimCurrentInstruction linehl=CurrentInstruction
sign define DlvimBreakpoint text=●

let s:subtab_names = [
    \ 'breakpoints',
    \ 'stack',
    \ 'log',
\ ]

let s:seed = srand()

function! s:format_subtabs_for_status_line(dlvim_window_id)
    let l:formatted_subtab_names = []
    let l:buffer_map = getwinvar(a:dlvim_window_id, 'dlvim_buffers')
    for l:subtab_name in s:subtab_names
        let l:subtab_bufnr = l:buffer_map[l:subtab_name]
        let l:current_dlvim_bufnr = winbufnr(a:dlvim_window_id)
        if l:current_dlvim_bufnr ==# l:subtab_bufnr
            let l:formatted_subtab_name = '%#ModeMsg#' .. l:subtab_name .. '%#StatusLine#'
        else
            let l:formatted_subtab_name = l:subtab_name
        endif
        let l:formatted_subtab_names = add(l:formatted_subtab_names, l:formatted_subtab_name)
    endfor
    return l:formatted_subtab_names
endfunction

function! s:dlvim_window_status_line(dlvim_window_id)
    let l:status_line = '%#StatusLine#'
    let l:status_line ..= 'Dlvim ['
    let l:status_line ..= ' ' .. join(s:format_subtabs_for_status_line(a:dlvim_window_id), ' | ')
    let l:status_line ..= '] '
    let l:status_line ..= '%#StatusLineNC#'
    let l:status_line ..= '(select with C-l or C-h)'
    let l:status_line ..= '%#StatusLine#'
    return l:status_line
endfunction

function! s:setup_dlvim_window_options(dlvim_window_id)
    call win_execute(a:dlvim_window_id, 'setlocal nonumber')
    call win_execute(a:dlvim_window_id, 'setlocal norelativenumber')
    call win_execute(a:dlvim_window_id, 'resize 10')
    call win_execute(a:dlvim_window_id, 'set winfixheight')

    let l:status_line_expr = expand('<SID>') .. 'dlvim_window_status_line(' .. a:dlvim_window_id .. ')'
    call win_execute(a:dlvim_window_id, 'setlocal statusline=%!' .. l:status_line_expr)
endfunction

function! s:setup_subtab_buffer(bufnr)
    call setbufvar(a:bufnr, '&bufhidden', 'hide')
    call setbufvar(a:bufnr, '&buftype', 'nofile')
endfunction

function! s:create_buffer_for_subtab(buffer_name)
    execute 'badd' a:buffer_name
    let l:bufnr = bufnr(a:buffer_name)
    call s:setup_subtab_buffer(l:bufnr)
    return l:bufnr
endfunction

function! s:uniqualize_name(session_id, name)
    return 'dlvim' .. a:session_id .. '_' .. a:name
endfunction

function! s:create_dlvim_buffers()
    let l:dlvim_session_id = rand(s:seed)
    let l:buffer_map = {}
    let l:buffer_map['breakpoints'] = s:create_buffer_for_subtab(s:uniqualize_name(l:dlvim_session_id, 'breakpoints'))
    let l:buffer_map['stack']       = s:create_buffer_for_subtab(s:uniqualize_name(l:dlvim_session_id, 'stack'))
    let l:buffer_map['log']         = s:create_buffer_for_subtab(s:uniqualize_name(l:dlvim_session_id, 'log'))
    return l:buffer_map
endfunction

function! s:setup_dlvim_window_buffers(dlvim_window_id)
    let l:buffer_map = s:create_dlvim_buffers()
    call setwinvar(a:dlvim_window_id, 'dlvim_buffers', l:buffer_map)
    call win_execute(a:dlvim_window_id, 'buffer ' .. l:buffer_map['breakpoints'])
endfunction

function! s:setup_dlvim_window(dlvim_window_id)
    call s:setup_dlvim_window_buffers(a:dlvim_window_id)
    call s:setup_dlvim_window_options(a:dlvim_window_id)
endfunction

function! s:allocate_dlvim_window()
    let l:previous_window_id = win_getid()
    rightbelow new
    let l:dlvim_window_id = win_getid()
    call win_gotoid(l:previous_window_id)

    return l:dlvim_window_id
endfunction

function! s:setup_tab_variables(code_window_id, dlvim_window_id)
    let t:dlvim_code_window_id = a:code_window_id
    let t:dlvim_window_id = a:dlvim_window_id
endfunction

function! s:run_dlvim() abort
    let l:code_window_id = win_getid()
    let l:dlvim_window_id = s:allocate_dlvim_window()
    call s:setup_tab_variables(l:code_window_id, l:dlvim_window_id)
    call s:setup_dlvim_window(l:dlvim_window_id)

    " let l:log_bufname = 'dlvim' . l:sessionID . '_log'
    " execute 'edit ' . l:log_bufname
    " set bufhidden=hide

    " let l:job = job_start(s:proxy_py_path, {'mode': 'json', 'err_io': 'buffer', 'err_name': l:log_bufname})
    " call ch_evalexpr(l:job, ['init', l:sessionID])
    " let l:chan = ch_open('localhost:7778', {'mode': 'json'})
    " let w:job = l:job
    " let w:chan = l:chan
    " let w:codewinid = l:codewinid
    " let g:DlvimSessions[l:sessionID] = v:null
endfunction

function! s:getSessionVariable(sessionID, varname, default = 0) abort
    let l:tabwin = win_id2tabwin(a:sessionID)
    if l:tabwin == [0, 0]
        throw 'Session ' . a:sessionID . 'not found'
    endif
    return gettabwinvar(l:tabwin[0], l:tabwin[1], a:varname, a:default)
endfunction

function! s:setSessionVariable(sessionID, varname, value) abort
    let l:tabwin = win_id2tabwin(a:sessionID)
    if l:tabwin == [0, 0]
        throw 'Session ' . a:sessionID . 'not found'
    endif
    return settabwinvar(l:tabwin[0], l:tabwin[1], a:varname, a:value)
endfunction

function! s:cleanup(sessionID) abort
    let l:tabwin = win_id2tabwin(a:sessionID)
    if l:tabwin == [0, 0] || gettabwinvar(l:tabwin[0], l:tabwin[1], 'dlvim') != 1
        return
    endif

    call s:ClearBreakpoints()
    call s:ClearCurrentInstruction()
    let l:job = s:getSessionVariable(a:sessionID, 'job', v:null)
    if l:job !=# v:null
        call job_stop(l:job)
    endif
    if has_key(g:DlvimSessions, a:sessionID)
        unlet g:DlvimSessions[a:sessionID]
    endif
endfunction

function! DlvCleanup() abort
    for l:sessionID in keys(g:DlvimSessions)
        call s:cleanup(l:sessionID)
    endfor
endfunction

function! ProxyRequest(req) abort
    let l:result = ch_evalexpr(b:chan, a:req)
endfunction

function! s:BreakpointName(breakpoint_id, dlv_bufnr) abort
    return 'Dlv' . a:dlv_bufnr . 'Breakpoint' . a:breakpoint_id
endfunction

function! s:ClearBreakpoints() abort
    sign unplace * group=DlvimBreakpoints
endfunction

function! OnBreakpointsUpdated(sessionID) abort
    let l:chan = s:getSessionVariable(a:sessionID, 'chan')
    let l:code_bufnr = winbufnr(s:getSessionVariable(a:sessionID, 'codewinid'))

    call s:ClearBreakpoints()
    let l:breakpoints = ch_evalexpr(l:chan, ['get_breakpoints'])
    for l:b in l:breakpoints['result']['Breakpoints']
        if l:b['id'] <= 0
            continue
        endif

        let l:bpname = s:BreakpointName(l:b['id'], a:sessionID)
	    execute 'sign place ' . l:b['id'] . ' group=DlvimBreakpoints line=' . l:b['line'] . ' name=DlvimBreakpoint buffer=' . l:code_bufnr
    endfor
    redraw
endfunction

function! s:ClearCurrentInstruction() abort
    sign unplace * group=DlvimCurrentInstruction
endfunction

function! s:SetCurrentInstruction(sessionID, file, line) abort
    let l:prevwinid = win_getid()
    let l:codewinid = getbufvar(a:sessionID, 'codewinid')

    call s:ClearCurrentInstruction()

    call win_gotoid(l:codewinid)
    execute 'edit ' . a:file
    call s:setSessionVariable(a:sessionID, 'dlvim_source_file', a:file) " TODO: this has something to do with symbolic links, but not sure what. It needs to be fixed
    execute 'sign place 1 name=DlvimCurrentInstruction group=DlvimCurrentInstruction line=' . a:line . ' buffer=' . bufnr()
    execute a:line
    normal zz
    call win_gotoid(l:prevwinid)
endfunction

function! OnStateUpdated(sessionID) abort
    let l:chan = s:getSessionVariable(a:sessionID, 'chan')
    let l:state = ch_evalexpr(l:chan, ['get_state'])
    if !has_key(l:state['result'], 'State') || l:state['result']['State']['Running']
        call s:ClearCurrentInstruction()
    else
        let l:curthread = l:state['result']['State']['currentThread']
        call s:SetCurrentInstruction(a:sessionID, l:curthread['file'], l:curthread['line'])
    endif
    redraw
endfunction

function! GetDlvimSession(sessionID) abort
    if a:sessionID > 0
        return a:sessionID
    endif
    if len(g:DlvimSessions) == 1
        return +keys(g:DlvimSessions)[0]
    endif
    if len(g:DlvimSessions) == 0
        throw 'No debug is currently in process'
    else
        throw 'More than one debug is currently in process, please disambiguate the debug session by specifying a buffer number'
    endif
endfunction

function! DlvimToggleBreakpointUnderCursor(sessionID = -1) abort
    let l:sessionID = GetDlvimSession(a:sessionID)
    let l:chan = s:getSessionVariable(l:sessionID, 'chan')
    call ch_evalexpr(l:chan, ['toggle_breakpoint', fnamemodify(s:getSessionVariable(l:sessionID, 'dlvim_source_file', bufname()), ':p'), line('.')])
endfunction

function! DlvimNext(sessionID = -1) abort
    let l:sessionID = GetDlvimSession(a:sessionID)
    let l:chan = s:getSessionVariable(l:sessionID, 'chan')
    call ch_evalexpr(l:chan, ['next'])
endfunction

function! DlvimContinue(sessionID = -1) abort
    let l:sessionID = GetDlvimSession(a:sessionID)
    let l:chan = s:getSessionVariable(l:sessionID, 'chan')
    call ch_evalexpr(l:chan, ['continue'])
endfunction

function! DlvimStep(sessionID = -1) abort
    let l:sessionID = GetDlvimSession(a:sessionID)
    let l:chan = s:getSessionVariable(l:sessionID, 'chan')
    call ch_evalexpr(l:chan, ['step'])
endfunction

function! DlvimStepOut(sessionID = -1) abort
    let l:sessionID = GetDlvimSession(a:sessionID)
    let l:chan = s:getSessionVariable(l:sessionID, 'chan')
    call ch_evalexpr(l:chan, ['stepout'])
endfunction

function! GetCurrentWord() abort
    let l:isk = &isk
    let &isk = 'a-z,A-Z,48-57,_,.,-,>'
    let l:word = expand('<cword>')
    let &isk = l:isk
    return l:word
endfunction

function! GetLastSelection() abort
    let l:zreg = getreg('z')
    silent normal gv"zy
    let l:selection = getreg('z')
    call setreg('z', l:zreg)
    return l:selection
endfunction

function! DlvimPrint(object, sessionID = -1) abort
    let l:sessionID = GetDlvimSession(a:sessionID)
    let l:chan = getSessionVariable(l:sessionID, 'chan')
    let l:response = ch_evalexpr(l:chan, ['eval', a:object])
    if l:response[1] !=# v:null
        echoerr l:response[1]
        return
    endif
    echo l:response[0]
endfunction

nnoremap <C-^>ac<C-^>b :call DlvimToggleBreakpointUnderCursor()<Cr>
nnoremap <C-^>ac<C-^>n :call DlvimNext()<Cr>
nnoremap <C-^>ac<C-^>c :call DlvimContinue()<Cr>
nnoremap <C-^>ac<C-^>s :call DlvimStep()<Cr>
nnoremap <C-^>ac<C-^>o :call DlvimStepOut()<Cr>
nnoremap <C-^>ac<C-^>k :call DlvimUp()<Cr>
nnoremap <C-^>ac<C-^>j :call DlvimDown()<Cr>
nnoremap <C-^>ac<C-^>i :call DlvimInterrupt()<Cr>
nnoremap <C-^>ac<C-^>p :call DlvimPrint(GetCurrentWord())<Cr>
vnoremap <C-^>ac<C-^>p :<C-U>call DlvimPrint(GetLastSelection())<Cr>

augroup dlvim
    autocmd!
    autocmd QuitPre * call s:cleanup(win_getid())
augroup END
