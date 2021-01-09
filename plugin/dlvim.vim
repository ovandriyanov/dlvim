command! Dlvim call s:runDlvim()

let s:proxy_py_path = '/home/ovandriyanov/github/ovandriyanov/dlvim/proxy/proxy.py'
"let s:proxy_py_path = ['bash', '-c', 'while true; do sleep 5; echo kek; done']

function! s:runDlvim() abort
    new
    let b:job = job_start(s:proxy_py_path, {'mode': 'json', 'err_io': 'buffer', 'err_name': 'thelog'})
    call ch_evalexpr(b:job, 'Are you ready?')
    execute 'autocmd BufDelete <buffer> call s:cleanupDlvClientBuffer(' . bufnr() . ')'
    let b:chan = ch_open('127.0.0.1:7778', {'mode': 'json'})
endfunction

function! s:cleanupDlvClientBuffer(bufnr) abort
    let l:job = getbufvar(a:bufnr, 'job')
    call job_stop(l:job)
endfunction

function! ProxyRequest(req) abort
    let l:result = ch_evalexpr(b:chan, a:req)
    echom 'RESULT: ' . l:result
endfunction
