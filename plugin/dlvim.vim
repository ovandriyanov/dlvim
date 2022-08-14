command! -nargs=+           Dlv          call s:start_session([<f-args>])
command!                    DlvBreak     call s:create_or_delete_breakpoint_on_the_current_line()
command!                    DlvContinue  call s:run_command('Continue', 'continue execution', {})
command!                    DlvNext      call s:run_command('Next', 'go to the next instruction', {})
command!                    DlvStep      call s:run_command('Step', 'step one instruction', {})
command!                    DlvStepout   call s:run_command('Stepout', 'step out', {})
command!                    DlvUp        call s:advance_stack_frame('Up', 'move one stack frame up')
command!                    DlvDown      call s:advance_stack_frame('Down', 'move one stack frame down')
command! -nargs=? -range=0  DlvEval      call s:evaluate_expression(<range>, <q-args>)

let s:repository_root = fnamemodify(expand('<sfile>'), ':h:h')
let s:proxy_path = s:repository_root .. '/proxy/proxy'

highlight CurrentInstruction ctermbg=lightblue
highlight CurrentStackFrame ctermbg=lightgreen
highlight CurrentGoroutine ctermbg=lightblue
sign define DlvimCurrentInstruction linehl=CurrentInstruction
sign define DlvimCurrentStackFrame linehl=CurrentStackFrame
sign define DlvimCurrentGoroutine linehl=CurrentGoroutine
sign define DlvimBreakpoint text=●

function! s:create_buffer(subtab_name, session) abort
    let l:buffer_name = s:uniqualize_name(a:session.id, a:subtab_name)
    execute 'badd' l:buffer_name
    let l:buffer = {'number': bufnr(l:buffer_name)}
    call s:setup_subtab_buffer(l:buffer, a:session, a:subtab_name)
    return l:buffer
endfunction

function! s:create_stack_buffer(session) abort
    let l:buffer = s:create_buffer('stack', a:session)
    call s:setup_stack_buffer_mappings(l:buffer.number)
    return l:buffer
endfunction

function! s:setup_stack_buffer_mappings(buffer) abort
    let l:switch_stack_frame_function_name = expand('<SID>') .. 'switch_stack_frame'
    execute printf('nnoremap <buffer> <Cr> :call %s(getcurpos()[1]-1)<Cr>', l:switch_stack_frame_function_name)
endfunction

function! s:create_goroutines_buffer(session) abort
    let l:buffer = s:create_buffer('goroutines', a:session)
    call s:setup_goroutines_buffer_mappings(l:buffer.number)
    return l:buffer
endfunction

function! s:setup_goroutines_buffer_mappings(buffer) abort
    let l:switch_goroutine_function_name = expand('<SID>') .. 'switch_to_goroutine_under_cursor'
    execute printf('nnoremap <buffer> <Cr> :call %s()<Cr>', l:switch_goroutine_function_name)
endfunction

function! s:create_objects_buffer(session) abort
    let l:buffer = s:create_buffer('objects', a:session)
    call setbufvar(l:buffer.number, '&filetype', 'json')
    return l:buffer
endfunction

function! s:create_terminal_buffer(subtab_name, command_factory, session) abort
    execute 'terminal'
    \  '++curwin'
    \  '++kill=TERM'
    \  '++noclose'
    \  a:command_factory(a:session)
    let l:buffer = {'number': bufnr()}
    call s:setup_subtab_buffer(l:buffer, a:session, a:subtab_name)
    return l:buffer
endfunction

let s:subtabs = {
\     'breakpoints': {
\         'index': 0,
\         'create_buffer': function(funcref(expand('<SID>') .. 'create_buffer'), ['breakpoints']),
\     },
\     'stack': {
\         'index': 1,
\         'create_buffer': function(funcref(expand('<SID>') .. 'create_stack_buffer'), []),
\     },
\     'goroutines': {
\         'index': 2,
\         'create_buffer': function(funcref(expand('<SID>') .. 'create_goroutines_buffer'), []),
\     },
\     'objects': {
\         'index': 3,
\         'create_buffer': function(funcref(expand('<SID>') .. 'create_objects_buffer'), []),
\     },
\     'console': {
\         'index': 4,
\         'create_buffer': function(funcref(expand('<SID>') .. 'create_terminal_buffer'), [
\             'console',
\             {session -> 'dlv connect ' .. session.proxy_listen_address},
\         ]),
\     },
\     'log': {
\         'index': 5,
\         'create_buffer': function(funcref(expand('<SID>') .. 'create_terminal_buffer'), [
\             'log',
\             {session -> printf('tail -n +1 -f %s', session.proxy_log_file)},
\         ]),
\     },
\ }

if !has_key(g:, 'dlvim')
    let g:dlvim = {
    \   'sessions': {},
    \   'current_session': v:null,
    \ }
endif

function! s:create_or_delete_breakpoint_on_the_current_line() abort
    let l:session = g:dlvim.current_session
    if type(l:session) == type(v:null)
        call s:print_error( 'No debugging session is currently in progress')
        return
    endif

    let l:response = ch_evalexpr(l:session.proxy_job, ['CreateOrDeleteBreakpoint', {
        \ 'file': expand('%:p'),
        \ 'line': line('.'),
    \ }])
    if has_key(l:response, 'Error')
        call s:print_error(l:response.Error)
        return
    endif
    call s:update_breakpoints(l:session)
endfunction

function s:run_command(command_name, command_description, args) abort
    let l:session = g:dlvim.current_session
    if type(l:session) == type(v:null)
        call s:print_error( 'No debugging session is currently in progress')
        return
    endif

    call s:clear_current_instruction_sign(l:session)
    call s:clear_stack_buffer(l:session)
    call s:clear_goroutines_buffer(l:session)
    let l:options = {'callback': function(funcref(expand('<SID>') .. 'on_run_command_response'), [l:session, a:command_description])}
    call ch_sendexpr(l:session.proxy_job, [a:command_name, a:args], l:options)
endfunction

function! s:advance_stack_frame(command_name, command_description) abort
    let l:session = g:dlvim.current_session
    if type(l:session) == type(v:null)
        call s:print_error( 'No debugging session is currently in progress')
        return
    endif

    let l:response = ch_evalexpr(l:session.proxy_job, [a:command_name, {}])
    if has_key(l:response, 'Error')
        call s:print_error(printf('cannot %s: %s', a:command_description, l:response.Error))
        return
    endif

    call s:update_stack_buffer(l:session, l:response.stack_trace, l:response.current_stack_frame)
    call s:follow_location_if_necessary(l:response.stack_trace[l:response.current_stack_frame])
    call s:clear_current_instruction_sign(l:session)
    call s:set_current_instruction_sign(l:session, l:response.stack_trace[l:response.current_stack_frame])
endfunction

function! s:switch_stack_frame(frame_index) abort
    let l:session = g:dlvim.current_session
    if type(l:session) == type(v:null)
        call s:print_error('No debugging session is currently in progress')
        return
    endif

    let l:response = ch_evalexpr(l:session.proxy_job, ['SwitchStackFrame', {'stack_frame': a:frame_index}])
    if has_key(l:response, 'Error')
        call s:print_error(printf('cannot switch to the stack frame %d: %s', a:frame_index, l:response.Error))
        return
    endif

    call s:update_stack_buffer(l:session, l:response.stack_trace, l:response.current_stack_frame)
    call s:follow_location_if_necessary(l:response.stack_trace[l:response.current_stack_frame])
    call s:clear_current_instruction_sign(l:session)
    call s:set_current_instruction_sign(l:session, l:response.stack_trace[l:response.current_stack_frame])
endfunction

function! s:switch_to_goroutine_under_cursor()
    let l:session = g:dlvim.current_session
    if type(l:session) == type(v:null)
        call s:print_error('No debugging session is currently in progress')
        return
    endif

    let l:response = ch_evalexpr(l:session.proxy_job, ['SwitchGoroutine', {'line': getline('.')}])
    call s:run_command('SwitchGoroutine', 'switch goroutine', {'line': getline('.')})
endfunction

function! s:on_run_command_response(session, command_description, channel, response) abort
    if has_key(a:response, 'Error')
        call s:print_error(printf('cannot %s: %s', a:command_description, a:response.Error))
        return
    endif

    call s:update_state(a:session, a:response.state, a:response.stack_trace)
endfunction

function! s:print_error(message) abort
    echohl ErrorMsg
    echomsg a:message
    echohl None
endfunction

function! s:on_breakpoints_updated(session, event_payload) abort
    call s:update_breakpoints(a:session)
endfunction

function! s:on_command_issued(session, event_payload) abort
    call s:clear_current_instruction_sign(a:session)
endfunction

function! s:on_state_updated(session, event_payload) abort
    call s:update_state(a:session, a:event_payload.state, a:event_payload.stack_trace)
endfunction

function! s:update_breakpoints(session) abort
    let l:response = ch_evalexpr(a:session.proxy_job, ['ListBreakpoints', {}])
    if has_key(l:response, 'Error')
        throw printf('cannot list breakpoints: %s', l:response.Error)
    endif
    let a:session.breakpoints = l:response.Breakpoints

    call s:update_breakpoint_signs(a:session)
    call s:set_buffer_contents(a:session, 'breakpoints', s:jsonify_list(a:session.breakpoints))
endfunction

function! s:jsonify_list(breakpoints_list) abort
    let l:result = []
    for l:breakpoint in a:breakpoints_list
        let l:result += [json_encode(l:breakpoint)]
    endfor
    return result
endfunction

function! s:update_state(session, state, stack_trace) abort
    if type(a:stack_trace) != type(v:null)
        call s:update_stack_buffer(a:session, a:stack_trace, 0)
    else
        call s:clear_stack_buffer(a:session)
    endif

    let l:current_goroutine = get(a:state, 'currentGoroutine', v:null)
    if type(l:current_goroutine) == type(v:null)
        let l:current_goroutine_id = -1
    else
        let l:current_goroutine_id = l:current_goroutine.id
    endif

    let l:response = ch_evalexpr(a:session.proxy_job, ['ListGoroutines', {'current_goroutine_id': l:current_goroutine_id}])
    if has_key(l:response, 'Error')
        call s:print_error('Cannot list goroutines: ' .. l:response.Error)
        let l:goroutines = []
        let l:current_goroutine_index = 0
    else
        let l:goroutines = l:response.goroutines
        let l:current_goroutine_index = l:response.current_goroutine_index
    endif

    call s:update_goroutines_buffer(a:session, l:goroutines, l:current_goroutine_index)

    call s:clear_current_instruction_sign(a:session)
    if a:state.exited
        echom 'Program exited with status ' .. a:state.exitStatus
        return
    endif
    if type(l:current_goroutine) == type(v:null)
        return
    endif
    let l:user_current_loc = get(l:current_goroutine, 'userCurrentLoc', v:null)
    if type(l:user_current_loc) == type(v:null)
        return
    endif

    call s:follow_location_if_necessary(l:user_current_loc)
    call s:set_current_instruction_sign(a:session, l:user_current_loc)
endfunction

function! s:clear_current_instruction_sign(session) abort
    call sign_unplace(a:session.current_instruction_sign_group)
endfunction

function! s:clear_current_goroutine_sign(session) abort
    call sign_unplace(a:session.current_goroutine_sign_group)
endfunction

function! s:clear_current_stack_frame_sign(session) abort
    call sign_unplace(a:session.current_stack_frame_sign_group)
endfunction

function! s:clear_stack_buffer(session) abort
    let l:stack_buffer = a:session.buffers.stack.number
    call deletebufline(l:stack_buffer, 1, '$')
    call s:clear_current_stack_frame_sign(a:session)
endfunction

function! s:update_stack_buffer(session, stack_trace, current_stack_frame) abort
    call s:set_buffer_contents(a:session, 'stack', s:jsonify_list(a:stack_trace))
    call s:set_current_stack_frame_sign(a:session, a:current_stack_frame)
endfunction

function! s:clear_goroutines_buffer(session) abort
    let l:goroutines_buffer = a:session.buffers.goroutines.number
    call deletebufline(l:goroutines_buffer, 1, '$')
    call s:clear_current_goroutine_sign(a:session)
endfunction

function! s:update_goroutines_buffer(session, goroutines, current_goroutine_index) abort
    call s:set_buffer_contents(a:session, 'goroutines', a:goroutines)
    if a:current_goroutine_index != -1
        call s:set_current_goroutine_sign(a:session, a:current_goroutine_index)
    endif
endfunction

function! s:follow_location_if_necessary(location) abort
    let l:code_window_id = -1
    let l:previous_window_id = -1
    let l:current_window_number = 1
    if &buftype !=# ''
        let l:previous_window_id = win_getid()
        for l:buffer in tabpagebuflist()
            if getbufvar(l:buffer, '&buftype') ==# ''
                let l:code_window_id = win_getid(l:current_window_number)
                break
            endif
            let l:current_window_number += 1
        endfor

        if l:code_window_id == -1
            split
            let l:code_window_id = win_getid()
        else
            call win_gotoid(l:code_window_id)
        endif
    endif

    if resolve(expand('%:p')) ==# resolve(fnamemodify(a:location.file, ':p'))
        if a:location.line < line('w0') || a:location.line > line('w$')
            " We are jumping outside of the current screen. Mark the current
            " position so that we can jump back to it with Ctrl+O
            normal! m'
            execute a:location.line
        endif
    else
        if &modified && !&hidden
            split
        endif
        execute 'edit' '+' .. a:location.line a:location.file
    endif

    if l:previous_window_id != -1
        call win_gotoid(l:previous_window_id)
    endif
endfunction

function! s:set_current_instruction_sign(session, location) abort
    let l:arbitrary_id = 1 " we only have one current instruction at a time, so we pick an arbitrary id
    let l:place_result = sign_place(l:arbitrary_id, a:session.current_instruction_sign_group, 'DlvimCurrentInstruction', a:location.file, {'lnum': a:location.line})
    if l:place_result == -1
        call s:print_error('cannot set current instruction sign at ' json_encode(a:location))
    endif
endfunction

function! s:set_current_stack_frame_sign(session, current_stack_frame) abort
    let l:arbitrary_id = 1 " we only have one current stack frame at a time, so we pick an arbitrary id
    let l:line_number = a:current_stack_frame + 1 " add 1 since lines are numbered from 1 and stack frames are numbered from 0
    let l:place_result = sign_place(l:arbitrary_id, a:session.current_stack_frame_sign_group, 'DlvimCurrentStackFrame', a:session.buffers.stack.number, {'lnum': l:line_number})
    if l:place_result == -1
        call s:print_error('cannot set current instruction sign at ' json_encode(a:location))
    endif
endfunction

function! s:set_current_goroutine_sign(session, current_goroutine_index) abort
    let l:arbitrary_id = 1 " we only have one current stack frame at a time, so we pick an arbitrary id
    let l:line_number = a:current_goroutine_index + 1 " add 1 since lines are numbered from 1 and goroutines are numbered from 0
    let l:place_result = sign_place(l:arbitrary_id, a:session.current_goroutine_sign_group, 'DlvimCurrentGoroutine', a:session.buffers.goroutines.number, {'lnum': l:line_number})
    if l:place_result == -1
        call s:print_error('cannot set current goroutine sign at ' json_encode(a:location))
    endif
endfunction

function! s:save_cursor_positions(subtab_name) abort
    let l:winnr = 0
    let l:window_cursor_positions = {}
    for l:bufnr in tabpagebuflist()
        let l:winnr += 1
        let l:dlvim = getbufvar(l:bufnr, 'dlvim', v:null)
        if type(l:dlvim) == type(v:null) || l:dlvim.subtab_name !=# a:subtab_name
            continue
        endif
        let l:window_cursor_positions[l:winnr] = getcurpos(l:winnr)[1:]
    endfor
    return l:window_cursor_positions
endfunction

function! s:restore_cursor_positions(window_cursor_positions) abort
    let l:old_window_id = win_getid()
    for [l:winnr, l:cursor_position] in items(a:window_cursor_positions)
        call win_gotoid(win_getid(l:winnr))
        call cursor(l:cursor_position)
    endfor
    call win_gotoid(l:old_window_id)
endfunction

" buffer_content is a list of strings representing lines of the buffer to be set
function! s:set_buffer_contents(session, buffer_name, buffer_content) abort
    let l:cursor_positions = s:save_cursor_positions(a:buffer_name)

    let l:buffer = a:session.buffers[a:buffer_name].number
    call deletebufline(l:buffer, 1, '$') " Delete everything
    let l:line_number = 0
    for l:line in a:buffer_content
        call appendbufline(l:buffer, l:line_number, l:line)
        let l:line_number += 1
    endfor
    call deletebufline(l:buffer, '$') " Delete the last line

    call s:restore_cursor_positions(l:cursor_positions)
endfunction

let s:event_handlers = {
\     'BREAKPOINTS_UPDATED': funcref(expand('<SID>') .. 'on_breakpoints_updated'),
\     'COMMAND_ISSUED':      funcref(expand('<SID>') .. 'on_command_issued'),
\     'STATE_UPDATED':       funcref(expand('<SID>') .. 'on_state_updated'),
\ }

let s:subtab_names = [
\     'breakpoints',
\     'stack',
\     'goroutines',
\     'objects',
\     'console',
\     'log',
\ ]

let s:seed = srand()

function! s:format_subtabs_for_status_line(window_id) abort
    let l:formatted_subtab_names = []
    for l:subtab_name in s:subtab_names
        let l:subtab_bufnr = b:dlvim.session.buffers[l:subtab_name].number
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
    let l:next_bufnr = b:dlvim.session.buffers[l:next_subtab_name].number


    let l:old_eventignore=&eventignore
    set eventignore=BufWinLeave
    execute 'buffer' l:next_bufnr
    let &eventignore=l:old_eventignore
endfunction

function! s:collect_garbage(bufnr_being_left) abort
    let l:session = getbufvar(a:bufnr_being_left, 'dlvim').session
    for l:buffer in values(l:session.buffers)
        let l:bufnr = l:buffer.number
        if l:bufnr == a:bufnr_being_left
            continue
        endif
        if bufwinid(l:bufnr) != -1
            return
        endif
    endfor

    call job_stop(l:session.proxy_job)
    while job_status(l:session.proxy_job) ==# 'run'
        sleep 20m
    endwhile

    call setbufvar(a:bufnr_being_left, '&bufhidden', 'wipe')
    if getbufvar(a:bufnr_being_left, '&buftype') ==# 'terminal'
        let l:job = term_getjob(a:bufnr_being_left)
        call job_stop(l:job)
    endif

    let l:values = values(l:session.buffers)
    for l:buffer in l:values
        let l:bufnr = l:buffer.number
        if l:bufnr == a:bufnr_being_left
            continue
        endif
        execute l:bufnr . 'bwipeout!'
    endfor
    echo 'Dlvim exited'

    call s:clear_breakpoint_signs(l:session)
    call s:clear_current_instruction_sign(l:session)

    call remove(g:dlvim.sessions, l:session.id)
    let g:dlvim.current_session = v:null
    for l:remaining_session in values(g:dlvim.sessions)
        let g:dlvim.current_session = l:remaining_session
        break
    endfor
endfunction

function! s:setup_subtab_buffer(buffer, session, subtab_name) abort
    let l:bufnr = a:buffer.number
    call setbufvar(l:bufnr, '&bufhidden', 'hide')
    call setbufvar(l:bufnr, '&buflisted', '0')
    if getbufvar(l:bufnr, '&buftype') !=# 'terminal'
        call setbufvar(l:bufnr, '&buftype', 'nofile')
    endif
    call setbufvar(l:bufnr, 'dlvim', {
    \     'session':     a:session,
    \     'subtab_name': a:subtab_name,
    \ })

    execute 'buffer' l:bufnr

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

let g:dlvim_debug_rpc = 1

function! s:create_proxy_job(session, dlv_argv, proxy_log_file) abort
    let l:job_options = {
    \      'mode':      'json',
    \      'err_io':    'file',
    \      'err_name':  a:proxy_log_file,
    \ }

    let l:job_args = []
    if g:dlvim_debug_rpc
        let l:job_args += ['--debug-rpc']
    endif

    let l:job = job_start([s:proxy_path] + l:job_args, l:job_options)

    let l:init_response = ch_evalexpr(l:job, ['Initialize', {'dlv_argv': a:dlv_argv}])
    if has_key(l:init_response, 'Error')
        throw printf('proxy initialization failed: %s', l:init_response.Error)
    endif
    let l:proxy_listen_address = l:init_response.proxy_listen_address

    call ch_sendexpr(l:job, ['GetNextEvent', {}], {'callback': function(funcref(expand('<SID>') .. 'on_next_event'), [a:session])})
    return [l:job, l:proxy_listen_address, l:init_response.state]
endfunction

function! s:on_next_event(session, channel, response) abort
    if has_key(a:response, 'Error')
        call s:print_error('cannot get next event: ' .. a:response.Error)
        return
    endif

    let l:event = a:response
    if !has_key(s:event_handlers, l:event.kind)
        echom 'Unhandled event: ' .. l:event.kind
    endif

    try
        call s:event_handlers[l:event.kind](a:session, l:event.payload)
    finally
        call ch_sendexpr(a:channel, ['GetNextEvent', {}], {'callback': function(funcref(expand('<SID>') .. 'on_next_event'), [a:session])})
    endtry
endfunction

function! s:create_session(dlv_argv) abort
    let l:proxy_log_file = tempname()
    let l:session = {
    \   'id':                               string(rand(s:seed)),
    \   'proxy_job':                        v:null,
    \   'proxy_listen_address':             v:null,
    \   'proxy_log_file':                   l:proxy_log_file,
    \   'buffers':                          v:null,
    \   'code_window_id':                   win_getid(),
    \   'breakpoints':                      [],
    \   'breakpoint_sign_group':            '',
    \   'current_instruction_sign_group':   '',
    \   'current_stack_frame_sign_group':   '',
    \   'current_goroutine_sign_group':     '',
    \ }

    let [l:proxy_job, l:proxy_listen_address, l:debugger_state] = s:create_proxy_job(l:session, a:dlv_argv, l:proxy_log_file)

    let l:session.breakpoint_sign_group = 'DlvimBreakpoints' .. l:session.id
    let l:session.current_instruction_sign_group = 'DlvimCurrentInstruction' .. l:session.id
    let l:session.current_stack_frame_sign_group = 'DlvimCurrentStackFrame' .. l:session.id
    let l:session.current_goroutine_sign_group = 'DlvimCurrentGoroutine' .. l:session.id
    let l:session.proxy_job = l:proxy_job
    let l:session.proxy_listen_address = l:proxy_listen_address
    let l:session.buffers = s:create_buffers(l:session)

    if type(l:debugger_state) != type(v:null)
        call s:update_state(l:session, l:debugger_state, v:null)
    endif
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
    call win_execute(a:window_id, printf('buffer %d', a:session.buffers[s:subtab_names[0]].number))
    let &eventignore = l:old_eventignore
endfunction

function! s:update_breakpoint_signs(session) abort
    call s:clear_breakpoint_signs(a:session)
    for l:breakpoint in a:session.breakpoints
        let l:place_result = sign_place(l:breakpoint.id, a:session.breakpoint_sign_group, 'DlvimBreakpoint', l:breakpoint.file, {'lnum': l:breakpoint.line})
        if l:place_result == -1
            call s:print_error('cannot set breakpoint sign for breakpoint ' .. l:breakpoint.id)
        endif
    endfor
endfunction

function! s:clear_breakpoint_signs(session) abort
    call sign_unplace(a:session.breakpoint_sign_group)
endfunction

function! s:start_session(dlv_argv) abort
    try
        let l:session = s:create_session(a:dlv_argv)
    catch
        echoerr 'cannot create session: ' .. v:exception
        return
    endtry
    let l:window_id = s:allocate_dlvim_window()
    call s:setup_dlvim_window(l:window_id, l:session)

    let g:dlvim.sessions[l:session.id] = l:session
    let g:dlvim.current_session = l:session
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

function! s:evaluate_expression(range, expr = '') abort
    if a:range != 0 " visual mode
        let l:args = {'expression': s:get_selection()}
    elseif a:expr == '' " normal mode, no expression given -> parse expression under the cursor
        let l:args = {'line': getline('.'), 'cursor_position': getcurpos()[2]-1}
    else
        let l:args = {'expression': a:expr} " parse the given expression
    endif

    let l:session = g:dlvim.current_session
    if type(l:session) == type(v:null)
        call s:print_error('No debugging session is currently in progress')
        return
    endif

    let l:response = ch_evalexpr(l:session.proxy_job, ['Evaluate', l:args])
    if has_key(l:response, 'Error')
        call s:print_error(l:response.Error)
        return
    endif

    call s:set_buffer_contents(l:session, 'objects', l:response.pretty)
    if get(l:response, 'one_line', '') ==# ''
        echo 'See objects subtab for the evaluation result'
    else
        echo l:response.one_line
    endif
endfunction

function! s:update_objects_buffer(session) abort
endfunction

function! s:get_selection() abort
    let l:saved_zreg = getreg('z')
    normal! gv"zy
    let l:expr = getreg('z')
    call setreg('z', l:saved_zreg)
    redraw
    return l:expr
endfunction

nnoremap <C-^>ac<C-^>b :DlvBreak<Cr>
nnoremap <C-^>ac<C-^>c :DlvContinue<Cr>
nnoremap <C-^>ac<C-^>n :DlvNext<Cr>
nnoremap <C-^>ac<C-^>s :DlvStep<Cr>
nnoremap <C-^>ac<C-^>o :DlvStepout<Cr>
nnoremap <C-^>ac<C-^>j :DlvUp<Cr>
nnoremap <C-^>ac<C-^>k :DlvDown<Cr>
nnoremap <C-^>ac<C-^>p :DlvEval<Cr>
vnoremap <C-^>ac<C-^>p :DlvEval<Cr>
