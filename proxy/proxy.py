#!/usr/bin/env python3

import asyncio
import jsonstreamer
import socket
from json_parser import JsonParser

proxy_listen_addr = ('127.0.0.1', 7777)
dlv_server_addr = ('127.0.0.1', 8888)


class BufferedSocket:
    def __init__(self, sock: socket.socket):
        self.sock = sock
        self.buf = bytes()
        self.json_parser = JsonParser()


    async def jsons(self):
        while True:
            data_chunk = await asyncio.get_event_loop().sock_recv(self.sock, 4096)
            if not data_chunk:
                return
            obj = self.json_parser.parse(bytes.decode(data_chunk))
            if obj is not None:
                yield obj


async def run_proxy_server(loop):
    proxy_server = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
    proxy_server.setblocking(False)
    proxy_server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    proxy_server.bind(proxy_listen_addr)
    proxy_server.listen(100)
    print('Listening at {}'.format(proxy_listen_addr))

    while True:
        print('Waiting for client...')
        client_socket, addr = await loop.sock_accept(proxy_server)
        print('Accepted client {}'.format(addr))

        #dlv_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
        dlv_socket = None
        #await loop.sock_connect(dlv_socket, dlv_server_addr)
        loop.create_task(read_requests(loop, client_socket, dlv_socket))


async def read_requests(loop, client_socket, dlv_socket):
    async for json in BufferedSocket(client_socket).jsons():
        print('JSON: {}'.format(json))


loop = asyncio.get_event_loop()
loop.create_task(run_proxy_server(loop))
loop.run_forever()
