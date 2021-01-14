command! Dlvim call s:runDlvim()

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

highlight CurrentInstruction ctermbg=lightblue
sign define DlvimCurrentInstruction linehl=CurrentInstruction
sign define DlvimBreakpoint text=●

if !exists('g:DlvimBuffers')
    let g:DlvimBuffers = {}
endif

function! s:runDlvim() abort
    let l:codewinid = win_getid()
    new
    let l:job = job_start(s:proxy_py_path, {'mode': 'json', 'err_io': 'buffer', 'err_name': 'thelog'})
    call ch_evalexpr(l:job, ['init', bufnr()])
    let l:chan = ch_open('127.0.0.1:7778', {'mode': 'json'})
    terminal ++curwin ++close dlv connect 127.0.0.1:7777
    resize 8
    execute 'autocmd BufDelete <buffer> call s:cleanupDlvClientBuffer(' . bufnr() . ')'
    let b:job = l:job
    let b:chan = l:chan
    let b:codewinid = l:codewinid
    let g:DlvimBuffers[bufnr()] = v:null
endfunction

function! s:cleanupDlvClientBuffer(bufnr) abort
    call s:ClearBreakpoints()
    call s:ClearCurrentInstruction()
    let l:job = getbufvar(a:bufnr, 'job', v:null)
    if l:job !=# v:null
        call job_stop(l:job)
    endif
    unlet g:DlvimBuffers[a:bufnr]
endfunction

function! DlvCleanup() abort
    for l:bufnr in keys(g:DlvimBuffers)
        call s:cleanupDlvClientBuffer(l:bufnr)
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

function! OnBreakpointsUpdated(bufnr) abort
    let l:chan = getbufvar(a:bufnr, 'chan')
    let l:bufnr = winbufnr(getbufvar(a:bufnr, 'codewinid'))

    call s:ClearBreakpoints()
    let l:breakpoints = ch_evalexpr(l:chan, ['get_breakpoints'])
    for l:b in l:breakpoints['result']['Breakpoints']
        if l:b['id'] <= 0
            continue
        endif

        let l:bpname = s:BreakpointName(l:b['id'], a:bufnr)
	    execute 'sign place ' . l:b['id'] . ' group=DlvimBreakpoints line=' . l:b['line'] . ' name=DlvimBreakpoint buffer=' . l:bufnr
    endfor
    redraw
endfunction

function! s:ClearCurrentInstruction() abort
    sign unplace * group=DlvimCurrentInstruction
endfunction

function! s:SetCurrentInstruction(bufnr, file, line) abort
    let l:prevwinid = win_getid()
    let l:codewinid = getbufvar(a:bufnr, 'codewinid')

    call s:ClearCurrentInstruction()

    call win_gotoid(l:codewinid)
    execute 'edit ' . a:file
    call setbufvar(a:bufnr, 'dlvim_source_file', a:file)
    execute 'sign place 1 name=DlvimCurrentInstruction group=DlvimCurrentInstruction line=' . a:line . ' buffer=' . bufnr()
    execute a:line
    normal zz
    call win_gotoid(l:prevwinid)
endfunction

function! OnStateUpdated(bufnr) abort
    let l:chan = getbufvar(a:bufnr, 'chan')
    let l:state = ch_evalexpr(l:chan, ['get_state'])
    if !has_key(l:state['result'], 'State') || l:state['result']['State']['Running']
        call s:ClearCurrentInstruction()
    else
        let l:curthread = l:state['result']['State']['currentThread']
        call s:SetCurrentInstruction(a:bufnr, l:curthread['file'], l:curthread['line'])
    endif
    redraw
endfunction

function! GetDlvimBuffer(bufnr) abort
    if a:bufnr > 0
        return a:bufnr
    endif
    if len(g:DlvimBuffers) == 1
        return +keys(g:DlvimBuffers)[0]
    endif
    if len(g:DlvimBuffers) == 0
        throw 'No debug is currently in process'
    else
        throw 'More than one debug is currently in process, please disambiguate the debug session by specifying a buffer number'
    endif
endfunction

function! DlvimToggleBreakpointUnderCursor(bufnr = -1) abort
    let l:bufnr = GetDlvimBuffer(a:bufnr)
    let l:chan = getbufvar(l:bufnr, 'chan')
    call ch_evalexpr(l:chan, ['toggle_breakpoint', fnamemodify(getbufvar(l:bufnr, 'dlvim_source_file', bufname()), ':p'), line('.')])
endfunction

function! DlvimNext(bufnr = -1) abort
    let l:bufnr = GetDlvimBuffer(a:bufnr)
    let l:chan = getbufvar(l:bufnr, 'chan')
    call ch_evalexpr(l:chan, ['next'])
endfunction

function! DlvimContinue(bufnr = -1) abort
    let l:bufnr = GetDlvimBuffer(a:bufnr)
    let l:chan = getbufvar(l:bufnr, 'chan')
    call ch_evalexpr(l:chan, ['continue'])
endfunction

function! DlvimStep(bufnr = -1) abort
    let l:bufnr = GetDlvimBuffer(a:bufnr)
    let l:chan = getbufvar(l:bufnr, 'chan')
    call ch_evalexpr(l:chan, ['step'])
endfunction

function! DlvimStepOut(bufnr = -1) abort
    let l:bufnr = GetDlvimBuffer(a:bufnr)
    let l:chan = getbufvar(l:bufnr, 'chan')
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

function! DlvimPrint(object, bufnr = -1) abort
    let l:bufnr = GetDlvimBuffer(a:bufnr)
    let l:chan = getbufvar(l:bufnr, 'chan')
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
