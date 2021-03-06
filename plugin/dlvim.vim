command! Dlvim call s:runDlvim()

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

highlight CurrentInstruction ctermbg=lightblue
sign define DlvimCurrentInstruction linehl=CurrentInstruction
sign define DlvimBreakpoint text=●

if !exists('g:DlvimSessions')
    let g:DlvimSessions = {} " window number -> session info
endif

function! s:runDlvim() abort
    let l:codewinid = win_getid()
    rightbelow new
    let w:dlvim = 1

    let l:sessionID = win_getid()

    let l:log_bufname = 'dlvim' . l:sessionID . '_log'
    execute 'edit ' . l:log_bufname
    set bufhidden=hide

    let l:job = job_start(s:proxy_py_path, {'mode': 'json', 'err_io': 'buffer', 'err_name': l:log_bufname})
    call ch_evalexpr(l:job, ['init', l:sessionID])
    let l:chan = ch_open('127.0.0.1:7778', {'mode': 'json'})
    let w:job = l:job
    let w:chan = l:chan
    let w:codewinid = l:codewinid
    let g:DlvimSessions[l:sessionID] = v:null
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
