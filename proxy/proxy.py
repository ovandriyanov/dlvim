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
dlv_argv = ['/home/ovandriyanov/bin/dlv', 'exec', '/home/ovandriyanov/go/src/kek/main', '--listen', '127.0.0.1:8888', '--headless']
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


async def handle_vim_requests(vim_conn):
    log('Receiving vim requests...')
    async for (req, future) in vim_conn.receive_requests():
        future.set_result(True)


async def read_dlv_client_requests(loop, client_socket, dlv_conn, vim_conn):
    async for j in BufferedSocket(client_socket).jsons():
        log('CLT --> PRX {}'.format(json.dumps(j)))
        if 'id' in j:
            # Request
            response = await dlv_conn.request(j)
            if j['method'] == 'RPCServer.CreateBreakpoint':
                vim_conn.ex('call OnBreakpointsUpdated({})'.format(bufnr))
            response['id'] = j['id']
            await loop.sock_sendall(client_socket, bytes(json.dumps(response) + '\n', 'ascii'))
            log('CLT <-- PRX {}'.format(response))
        else:
            # Notification
            log('Notification')
            dlv_conn.send_notification(j)


async def connect_to_dlv():
    dlv_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
    dlv_socket.setblocking(False)
    await asyncio.get_event_loop().sock_connect(dlv_socket, dlv_server_addr)
    log('Connected to DLV')
    return DlvConnection(loop, dlv_socket)


async def get_breakpoints(loop, dlv_conn):
    return await dlv_conn.request({'method': 'RPCServer.ListBreakpoints', 'params': [{}]})


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
    loop.create_task(handle_vim_requests(vim_conn))
    loop.run_forever()
