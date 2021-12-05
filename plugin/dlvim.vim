command! -nargs=+ Dlv call s:start_session([<f-args>])

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

highlight CurrentInstruction ctermbg=lightblue
sign define DlvimCurrentInstruction linehl=CurrentInstruction
sign define DlvimBreakpoint text=●

function! s:create_buffer(subtab_name, session) abort
    let l:buffer_name = s:uniqualize_name(a:session.id, a:subtab_name)
    execute 'badd' l:buffer_name
    let l:bufnr = bufnr(l:buffer_name)
    call s:setup_subtab_buffer(l:bufnr, a:session, a:subtab_name)
    return l:bufnr
endfunction

function! s:create_terminal(subtab_name, session) abort
    let l:buffer_name = s:uniqualize_name(a:session.id, a:subtab_name)
    terminal ++curwin ++kill=TERM ++noclose
    call s:setup_subtab_buffer(bufnr(), a:session, a:subtab_name)
    return bufnr()
endfunction

let s:subtabs = {
    \ 'breakpoints': {
        \ 'index': 0,
        \ 'create_buffer': function(funcref(expand('<SID>') .. 'create_buffer'), ['breakpoints']),
    \ },
    \ 'stack': {
        \ 'index': 1,
        \ 'create_buffer': function(funcref(expand('<SID>') .. 'create_buffer'), ['stack']),
    \ },
    \ 'log': {
        \ 'index': 2,
        \ 'create_buffer': function(funcref(expand('<SID>') .. 'create_terminal'), ['log']),
    \ },
\ }

let s:subtab_names = [
    \ 'breakpoints',
    \ 'stack',
    \ 'log',
\ ]

let s:seed = srand()

function! s:format_subtabs_for_status_line(window_id) abort
    let l:formatted_subtab_names = []
    for l:subtab_name in keys(s:subtabs)
        let l:subtab_bufnr = b:dlvim.session.buffers[l:subtab_name]
        if winbufnr(a:window_id) ==# l:subtab_bufnr
            let l:formatted_subtab_name = '%#ModeMsg#' .. l:subtab_name .. '%#StatusLine#'
        else
            let l:formatted_subtab_name = l:subtab_name
        endif
        let l:formatted_subtab_names = add(l:formatted_subtab_names, l:formatted_subtab_name)
    endfor
    return l:formatted_subtab_names
endfunction

function! s:dlvim_window_status_line() abort
    let l:status_line = '%#StatusLine#'
    let l:status_line ..= 'Dlvim ['
    let l:status_line ..= ' ' .. join(s:format_subtabs_for_status_line(win_getid()), ' | ')
    let l:status_line ..= '] '
    let l:status_line ..= '%#StatusLineNC#'
    let l:status_line ..= '(select with C-l or C-h)'
    let l:status_line ..= '%#StatusLine#'
    return l:status_line
endfunction

function! s:get_next_subtab_name(current_subtab_name, direction) abort
    let l:subtab_index = s:subtabs[a:current_subtab_name].index
    let l:offset = a:direction ==# 'right' ? 1 : -1
    let l:next_subtab_index = (l:subtab_index + l:offset) % len(s:subtabs)
    return s:subtab_names[l:next_subtab_index]
endfunction

function! s:rotate_subtab(direction) abort
    let l:current_subtab_name = b:dlvim.subtab_name
    let l:next_subtab_name = s:get_next_subtab_name(l:current_subtab_name, a:direction)
    let l:next_bufnr = b:dlvim.session.buffers[l:next_subtab_name]


    let l:old_eventignore=&eventignore
    set eventignore=BufWinLeave
    execute 'buffer' l:next_bufnr
    let &eventignore=l:old_eventignore
endfunction

function! s:collect_garbage(bufnr_being_left) abort
    let l:session = getbufvar(a:bufnr_being_left, 'dlvim').session
    for l:bufnr in values(l:session.buffers)
        if l:bufnr == a:bufnr_being_left
            continue
        endif
        if bufwinid(l:bufnr) != -1
            return
        endif
    endfor

    call setbufvar(a:bufnr_being_left, '&bufhidden', 'wipe')
    if getbufvar(a:bufnr_being_left, '&buftype') ==# 'terminal'
        let l:job = term_getjob(a:bufnr_being_left)
        call job_stop(l:job)
    endif

    let l:values = values(l:session.buffers)
    for l:bufnr in l:values
        if l:bufnr == a:bufnr_being_left
            continue
        endif
        execute l:bufnr . 'bwipeout!'
    endfor
    echo 'Dlvim exited'
endfunction

function! s:setup_subtab_buffer(bufnr, session, subtab_name) abort
    call setbufvar(a:bufnr, '&bufhidden', 'hide')
    call setbufvar(a:bufnr, '&buflisted', '0')
    if getbufvar(a:bufnr, '&buftype') !=# 'terminal'
        call setbufvar(a:bufnr, '&buftype', 'nofile')
    endif
    call setbufvar(a:bufnr, 'dlvim', {
        \ 'session':     a:session,
        \ 'subtab_name': a:subtab_name,
    \ })

    execute 'buffer' a:bufnr

    let l:rotate_subtab_function_name = expand('<SID>') .. 'rotate_subtab'
    execute printf('nnoremap <buffer> <C-l> :call %s("right")<Cr>', l:rotate_subtab_function_name)
    execute printf('nnoremap <buffer> <C-h> :call %s("left" )<Cr>', l:rotate_subtab_function_name)
    execute printf('tnoremap <buffer> <C-l> <C-^>:call %s("right")<Cr>', l:rotate_subtab_function_name)
    execute printf('tnoremap <buffer> <C-h> <C-^>:call %s("left" )<Cr>', l:rotate_subtab_function_name)

    let l:status_line_expr = '%{%' .. expand('<SID>') .. 'dlvim_window_status_line()' .. '%}'
    execute printf('setlocal statusline=%s', l:status_line_expr)

    autocmd BufWinLeave <buffer> call s:collect_garbage(str2nr(expand('<abuf>')))
endfunction

function! s:uniqualize_name(session_id, name) abort
    return printf('dlvim%s_%s', a:session_id, a:name)
endfunction

function! s:create_buffers(session)
    let l:previous_window_id = win_getid()
    new

    let buffers = {}
    let l:old_eventignore=&eventignore
    set eventignore=BufWinLeave
    for [l:subtab_name, l:subtab] in items(s:subtabs)
        let l:buffers[l:subtab_name] = l:subtab.create_buffer(a:session)
    endfor
    close
    let &eventignore = l:old_eventignore

    call win_gotoid(l:previous_window_id)
    return l:buffers
endfunction

function! s:create_session(dlv_argv) abort
    let l:proxy_job = {}

    let l:session = {
        \ 'id':      rand(s:seed),
        \ 'proxy_job': l:proxy_job,
        \ 'buffers': {},
    \ }
    let l:session.buffers = s:create_buffers(l:session)

    return l:session
endfunction

function! s:allocate_dlvim_window() abort
    let l:previous_window_id = win_getid()
    rightbelow new
    let l:window_id = win_getid()
    call win_gotoid(l:previous_window_id)
    return l:window_id
endfunction

function! s:setup_dlvim_window(window_id, session) abort
    call win_execute(a:window_id, 'resize 10')
    call win_execute(a:window_id, 'set winfixheight')

    let l:old_eventignore=&eventignore
    set eventignore=BufWinLeave
    call win_execute(a:window_id, printf('buffer %d', a:session['buffers'][s:subtab_names[0]]))
    let &eventignore = l:old_eventignore
endfunction

function! s:start_session(dlv_argv) abort
    let l:session = s:create_session(a:dlv_argv)
    let l:window_id = s:allocate_dlvim_window()
    call s:setup_dlvim_window(l:window_id, l:session)

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

"function! s:cleanup(sessionID) abort
"    let l:tabwin = win_id2tabwin(a:sessionID)
"    if l:tabwin == [0, 0] || gettabwinvar(l:tabwin[0], l:tabwin[1], 'dlvim') != 1
"        return
"    endif
"
"    call s:ClearBreakpoints()
"    call s:ClearCurrentInstruction()
"    let l:job = s:getSessionVariable(a:sessionID, 'job', v:null)
"    if l:job !=# v:null
"        call job_stop(l:job)
"    endif
"    if has_key(g:DlvimSessions, a:sessionID)
"        unlet g:DlvimSessions[a:sessionID]
"    endif
"endfunction

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

"augroup dlvim
"    autocmd!
"    autocmd BufWinLeave * call s:cleanup(win_getid())
"augroup END
