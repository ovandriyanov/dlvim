from json_parser import JsonParser
import asyncio
import json
import socket
import sys

def log(msg):
    print(msg, file=sys.stderr, flush=True)

class VimConnection:
    def __init__(self, loop: asyncio.AbstractEventLoop, socket: socket.socket):
        self.loop = loop
        self.socket = socket
        self.send_queue = []
        self.send_queue_ready = asyncio.Future()
        self.json_parser = JsonParser()

        self.sender_task = loop.create_task(self.send())


    def __del__(self):
        self.close()


    def __enter__(self):
        pass


    def __exit__(self):
        return self.close()


    async def join_task(self, task, taskname):
        try:
            await task
        except Exception as e:
            log('{}: {}'.format(taskname, e))


    async def close(self):
        self.socket.close()
        self.send_queue_ready.set_exception(Exception('Canceled'))
        await self.join_task(self.sender_task, 'Sender task')


    def ex(self, ex_command):
        req = ["ex", ex_command]
        self.send_queue.append(req)
        if len(self.send_queue) == 1:
            self.send_queue_ready.set_result(True)


    async def send(self):
        while True:
            if len(self.send_queue) == 0:
                await self.send_queue_ready
            self.send_queue_ready = asyncio.Future()

            sending = self.send_queue
            self.send_queue = []
            for item in sending:
                await self.loop.sock_sendall(self.socket, self.marshal(item))
                log('PRX --> VIM {}'.format(json.dumps(item)))


    async def receive_requests(self):
        while True:
            data = await self.loop.sock_recv(self.socket, 4096)
            if not data:
                return
            for j in self.json_parser.parse(bytes.decode(data, 'ascii')):
                log('PRX <-- VIM {}'.format(json.dumps(j)))
                reqid, req = j
                future = asyncio.Future()
                yield req, future
                result = await future
                self.send_queue.append([reqid, result])
                if len(self.send_queue) == 1:
                    self.send_queue_ready.set_result(True)


    def marshal(self, data):
        return bytes(json.dumps(data) + '\n', 'ascii')
