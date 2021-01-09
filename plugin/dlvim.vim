command! Dlvim call s:runDlvim()

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

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
    echom 'DELETE BUFFER ' . a:bufnr
    let l:job = getbufvar(a:bufnr, 'job')
    call job_stop(l:job)
endfunction

function! ProxyRequest(req) abort
    let l:result = ch_evalexpr(b:chan, a:req)
    echom 'RESULT: ' . l:result
endfunction

function! s:BreakpointName(breakpoint_id, dlv_bufnr) abort
    return 'Dlv' . a:dlv_bufnr . 'Breakpoint' . a:breakpoint_id
endfunction

function! OnBreakpointsUpdated(bufnr) abort
    "echom 'BREAKPOINTS UPDATED'
    let l:chan = getbufvar(a:bufnr, 'chan')
    let l:bufnr = winbufnr(getbufvar(a:bufnr, 'codewinid'))

    let l:breakpoints = ch_evalexpr(l:chan, ['get_breakpoints'])

    execute 'sign unplace * group=Dlvim buffer=' . l:bufnr
    for l:sign in sign_getdefined()
        if l:sign['name'] =~# 'Dlv[0-9]\+Breakpoint[0-9]\+'
            call sign_undefine(l:sign['name'])
        endif
    endfor
    for l:b in l:breakpoints['result']['Breakpoints']
        if l:b['id'] <= 0
            continue
        endif

        let l:bpname = s:BreakpointName(l:b['id'], a:bufnr)
	    execute 'sign define ' . l:bpname . ' text=⬤'
	    execute 'sign place ' . l:b['id'] . ' group=Dlvim line=' . l:b['line'] . ' name=' . l:bpname . ' buffer=' . l:bufnr
    endfor
    redraw
endfunction
