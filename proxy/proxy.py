#!/usr/bin/env python3

from dlv_connection import DlvConnection
from json_parser import JsonParser
import asyncio
import json
import jsonstreamer
import socket
import sys


proxy_listen_addr = ('127.0.0.1', 7777)
dlv_server_addr = ('127.0.0.1', 8888)


def log(msg):
    print(msg, file=sys.stderr)


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


async def run_proxy_server(loop):
    proxy_server = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
    proxy_server.setblocking(False)
    proxy_server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    proxy_server.bind(proxy_listen_addr)
    proxy_server.listen(100)
    log('Listening at {}'.format(proxy_listen_addr))

    while True:
        log('Waiting for client...')
        client_socket, addr = await loop.sock_accept(proxy_server)
        log('Accepted client {}'.format(addr))

        try:
            dlv_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
            dlv_socket.setblocking(False)
            await loop.sock_connect(dlv_socket, dlv_server_addr)
            log('Connected to DLV for client {}'.format(addr))
            dlv_conn = DlvConnection(loop, dlv_socket)
        except:
            client_socket.close()
            raise
        loop.create_task(read_requests(loop, client_socket, dlv_conn))


async def read_requests(loop, client_socket, dlv_conn):
    async for j in BufferedSocket(client_socket).jsons():
        log('CLT --> PRX {}'.format(json.dumps(j)))
        if 'id' in j:
            # Request
            response = await dlv_conn.request(j)
            if j['method'] == 'RPCServer.CreateBreakpoint':
                log('BREAKPOINTS: {}'.format(await get_breakpoints(loop, dlv_conn)))
            response['id'] = j['id']
            await loop.sock_sendall(client_socket, bytes(json.dumps(response) + '\n', 'ascii'))
            log('CLT <-- PRX {}'.format(response))
        else:
            # Notification
            log('Notification')
            dlv_conn.send_notification(j)


async def get_breakpoints(loop, dlv_conn):
    return await dlv_conn.request({'method': 'RPCServer.ListBreakpoints', 'params': [{}]})


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.create_task(run_proxy_server(loop))
    loop.run_forever()
