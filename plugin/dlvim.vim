command! Dlvim call s:runDlvim()

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

highlight CurrentInstruction ctermbg=lightblue
sign define DlvimCurrentInstruction linehl=CurrentInstruction
sign define DlvimBreakpoint text=⬤

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
endfunction

function! s:cleanupDlvClientBuffer(bufnr) abort
    call s:ClearBreakpoints()
    call s:ClearCurrentInstruction()
    let l:job = getbufvar(a:bufnr, 'job')
    call job_stop(l:job)
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
