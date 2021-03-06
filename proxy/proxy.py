#!/usr/bin/env python3

from dlv_connection import DlvConnection
from vim_connection import VimConnection
from json_parser import JsonParser
import asyncio
import json
import jsonstreamer
import socket
import sys


proxy_listen_addr = ('127.0.0.1', 7777)
vim_listen_addr = ('127.0.0.1', 7778)
dlv_server_addr = ('127.0.0.1', 8888)
dlv_argv = ['/home/ovandriyanov/bin/dlv', 'exec', '/home/ovandriyanov/go/src/kek/main', '--listen', '127.0.0.1:8888', '--headless', '--accept-multiclient']
bufnr = -1


def log(msg):
    print(msg, file=sys.stderr, flush=True)


class BufferedSocket:
    def __init__(self, sock: socket.socket):
        self.sock = sock
        self.json_parser = JsonParser()


    async def jsons(self):
        while True:
            data_chunk = await asyncio.get_event_loop().sock_recv(self.sock, 4096)
            if not data_chunk:
                return
            for obj in self.json_parser.parse(bytes.decode(data_chunk)):
                yield obj


def make_listen_socket(addr):
    listen_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
    listen_socket.setblocking(False)
    listen_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    listen_socket.bind(addr)
    listen_socket.listen(100)
    log('Listening at {}'.format(addr))
    return listen_socket


async def run_proxy_server(loop, listen_socket, dlv_conn, vim_conn):
    while True:
        log('Waiting for dlv client...')
        client_socket, addr = await loop.sock_accept(listen_socket)
        log('Accepted dlv client {}'.format(addr))
        loop.create_task(read_dlv_client_requests(loop, client_socket, dlv_conn, vim_conn))


async def accept_vim(listen_socket):
    log('Waiting for vim to connect...')
    client_socket, addr = await loop.sock_accept(listen_socket)
    log('Accepted vim {}'.format(addr))
    return VimConnection(asyncio.get_event_loop(), client_socket)



async def handle_vim_req(vim_conn, dlv_conn, req, future):
    if req[0] == 'get_breakpoints':
        breakpoints = await get_breakpoints(dlv_conn)
        future.set_result(breakpoints)
    elif req[0] == 'get_state':
        state = await get_state(dlv_conn)
        future.set_result(state)
    elif req[0] == 'toggle_breakpoint':
        try:
            await toggle_breakpoint(dlv_conn, req[1], req[2])
            future.set_result(None)
            vim_conn.ex('call OnBreakpointsUpdated({})'.format(bufnr))
        except Exception as e:
            future.set_result(str(e))
    elif req[0] == 'next':
        try:
            asyncio.get_event_loop().create_task(command(vim_conn, dlv_conn, 'next'))
            future.set_result(None)
            vim_conn.ex('call OnStateUpdated({})'.format(bufnr))
        except Exception as e:
            future.set_result(str(e))
    elif req[0] == 'continue':
        try:
            asyncio.get_event_loop().create_task(command(vim_conn, dlv_conn, 'continue'))
            future.set_result(None)
            vim_conn.ex('call OnStateUpdated({})'.format(bufnr))
        except Exception as e:
            future.set_result(str(e))
    elif req[0] == 'step':
        try:
            asyncio.get_event_loop().create_task(command(vim_conn, dlv_conn, 'step'))
            future.set_result(None)
            vim_conn.ex('call OnStateUpdated({})'.format(bufnr))
        except Exception as e:
            future.set_result(str(e))
    elif req[0] == 'stepout':
        try:
            asyncio.get_event_loop().create_task(command(vim_conn, dlv_conn, 'stepOut'))
            future.set_result(None)
            vim_conn.ex('call OnStateUpdated({})'.format(bufnr))
        except Exception as e:
            future.set_result(str(e))
    elif req[0] == 'eval':
        try:
            value = await evaluate(dlv_conn, req[1])
            future.set_result([value, None])
        except Exception as e:
            log('EXCEPTION: {}'.format(e))
            future.set_result([None, str(e)])


async def handle_vim_requests(dlv_conn, vim_conn):
    log('Receiving vim requests...')
    async for (req, future) in vim_conn.receive_requests():
        asyncio.get_event_loop().create_task(handle_vim_req(vim_conn, dlv_conn, req, future))


def is_pc_change_command(j):
    return j['method'] == 'RPCServer.Command' and j['params'][0]['name'] in {'continue', 'next', 'step', 'stepOut'}


async def handle_dlv_request(client_socket, vim_conn, dlv_conn, j):
    log('CLT --> PRX {}'.format(json.dumps(j)))
    if is_pc_change_command(j):
        vim_conn.ex('call OnStateUpdated({})'.format(bufnr))
    response = await dlv_conn.request(j)
    if j['method'] in {'RPCServer.CreateBreakpoint', 'RPCServer.ClearBreakpoint'}:
        vim_conn.ex('call OnBreakpointsUpdated({})'.format(bufnr))
    elif is_pc_change_command(j):
        vim_conn.ex('call OnStateUpdated({})'.format(bufnr))
    response['id'] = j['id']
    await asyncio.get_event_loop().sock_sendall(client_socket, bytes(json.dumps(response) + '\n', 'ascii'))
    log('CLT <-- PRX {}'.format(json.dumps(response)))


async def read_dlv_client_requests(loop, client_socket, dlv_conn, vim_conn):
    async for j in BufferedSocket(client_socket).jsons():
        if 'id' in j:
            loop.create_task(handle_dlv_request(client_socket, vim_conn, dlv_conn, j))
        else:
            dlv_conn.send_notification(j)


async def connect_to_dlv():
    dlv_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
    dlv_socket.setblocking(False)
    await asyncio.get_event_loop().sock_connect(dlv_socket, dlv_server_addr)
    log('Connected to DLV')
    return DlvConnection(loop, dlv_socket)


async def get_breakpoints(dlv_conn):
    return await dlv_conn.request({'method': 'RPCServer.ListBreakpoints', 'params': [{}]})


async def get_state(dlv_conn):
    return await dlv_conn.request({'method': 'RPCServer.State', "params": [{'NonBlocking': True}]})


async def toggle_breakpoint(dlv_conn, file_name, line_number):
    response = await dlv_conn.request({
        "method": "RPCServer.FindLocation",
        "params": [{
            "Scope": {
                "GoroutineID": -1,
                "Frame": 0,
                "DeferredCall": 0
            },
            "Loc": "{}:{}".format(file_name, line_number),
            "IncludeNonExecutableLines": False,
        }],
    })

    if response['error'] is not None:
        raise Exception(response['error'])
    locations = response['result']['Locations']
    if len(locations) < 1:
        raise Exception('No locations found for line {} at file {}'.format(line_number, file_name))
    loc = locations[0]

    current_breakpoints = await get_breakpoints(dlv_conn)
    existing_bp_id = -1
    for bp in current_breakpoints['result']['Breakpoints']:
        if bp['id'] < 0:
            continue
        if bp['file'] == file_name and bp['line'] == line_number:
            existing_bp_id = bp['id']
            break
    if existing_bp_id > 0:
        response = await dlv_conn.request({
            "method": "RPCServer.ClearBreakpoint",
            "params": [{
                "Id": existing_bp_id,
                "Name": ""
            }],
        })
    else:
        response = await dlv_conn.request({
            "method": "RPCServer.CreateBreakpoint",
            "params": [{
                "Breakpoint": {
                    "id": 0,
                    "name": "",
                    "addr": loc['pc'],
                    "addrs": [ loc['pc'] ],
                    "file": "",
                    "line": 0,
                    "Cond": "",
                    "continue": False,
                    "traceReturn": False,
                    "goroutine": False,
                    "stacktrace": 0,
                    "LoadArgs": None,
                    "LoadLocals": None,
                    "hitCount": None,
                    "totalHitCount": 0
                }
            }],
        })
    if response['error'] is not None:
        raise Exception(result['error'])


async def command(vim_conn, dlv_conn, cmd):
    await dlv_conn.request({
        "method": "RPCServer.Command",
        "params": [{
            "name": cmd,
            "ReturnInfoLoadConfig": {
                "FollowPointers": True,
                "MaxVariableRecurse": 1,
                "MaxStringLen": 64,
                "MaxArrayValues": 64,
                "MaxStructFields": -1
            }
        }],
    })
    vim_conn.ex('call OnStateUpdated({})'.format(bufnr))


async def evaluate(dlv_conn, expr):
    response = await dlv_conn.request({
        "method": "RPCServer.Eval",
        "params": [{
            "Scope": {
                "GoroutineID": -1,
                "Frame": 0,
                "DeferredCall": 0
            },
            "Expr": expr,
            "Cfg": {
                "FollowPointers": True,
                "MaxVariableRecurse": 1,
                "MaxStringLen": 64,
                "MaxArrayValues": 64,
                "MaxStructFields": -1
            }
        }],
    })
    if response['error'] is not None:
        raise Exception(response['error'])
    return response['result']['Variable']['value']


async def continue_execution(dlv_conn):
    pass


async def step(dlv_conn):
    pass


async def stepout(dlv_conn):
    pass


async def run_dlv_server():
    process = await asyncio.subprocess.create_subprocess_exec(
        *dlv_argv,
        stdin = asyncio.subprocess.DEVNULL,
        stdout = asyncio.subprocess.PIPE,
        stderr = asyncio.subprocess.DEVNULL,
    )
    line = await process.stdout.readline()
    if line != b'API server listening at: 127.0.0.1:8888\n':
        raise Exception('Wtf: {}'.format(line))
    return process


if __name__ == '__main__':
    loop = asyncio.get_event_loop()

    dlv_process = loop.run_until_complete(run_dlv_server())
    dlv_conn = loop.run_until_complete(connect_to_dlv())
    vim_listen_socket = make_listen_socket(vim_listen_addr)
    proxy_listen_socket = make_listen_socket(proxy_listen_addr)
    req = json.loads(sys.stdin.readline())
    assert len(req[1]) == 2
    assert req[1][0] == 'init'
    bufnr = req[1][1]
    print('[{}, "Ready to accept vim"]'.format(req[0]), flush=True)
    vim_conn = loop.run_until_complete(accept_vim(vim_listen_socket))

    loop.create_task(dlv_process.communicate())
    loop.create_task(run_proxy_server(loop, proxy_listen_socket, dlv_conn, vim_conn))
    loop.create_task(handle_vim_requests(dlv_conn, vim_conn))
    loop.run_forever()
