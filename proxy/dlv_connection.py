from json_parser import JsonParser
import asyncio
import copy
import ctypes
import json
import socket

class DlvConnection:
    def __init__(self, loop: asyncio.AbstractEventLoop, socket: socket.socket):
        self.loop = loop
        self.socket = socket
        self.reqid = ctypes.c_uint32(0)
        self.futures = {}
        self.send_queue = []
        self.send_queue_ready = asyncio.Future()
        self.json_parser = JsonParser()

        self.sender_task = loop.create_task(self.send_requests())
        self.receiver_task = loop.create_task(self.receive_responses())


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
            print('{}: {}'.format(taskname, e))


    async def close(self):
        self.socket.close()
        self.send_queue_ready.set_exception(Exception('Canceled'))
        await self.join_task(self.sender_task, 'Sender task')
        await self.join_task(self.receiver_task, 'Receiver task')


    def request(self, req):
        req_id = self.next_id()
        future = asyncio.Future()
        self.futures[req_id] = future

        req_copy = copy.deepcopy(req)
        req_copy['id'] = req_id

        self.send_queue.append(req_copy)
        if len(self.send_queue) == 1:
            self.send_queue_ready.set_result(True)

        return future


    def send_notification(self, notification):
        self.send_queue.append(notification)
        if len(self.send_queue) == 1:
            self.send_queue_ready.set_result(True)


    def next_id(self):
        if self.reqid.value in self.futures:
            raise Exception('Request ID collision')
        result = self.reqid.value
        self.reqid.value += 1
        return result


    async def send_requests(self):
        while True:
            if len(self.send_queue) == 0:
                await self.send_queue_ready
            self.send_queue_ready = asyncio.Future()

            sending = self.send_queue
            self.send_queue = []
            for item in sending:
                await self.loop.sock_sendall(self.socket, self.marshal(item))
                print('PRX --> DLV {}'.format(item))


    async def receive_responses(self):
        while True:
            data = await self.loop.sock_recv(self.socket, 4096)
            if not data:
                return
            for j in self.json_parser.parse(bytes.decode(data, 'ascii')):
                print('PRX <-- DLV {}'.format(j))
                if 'id' not in j:
                    continue
                future: asyncio.Future = self.futures[j['id']]
                del j['id']
                future.set_result(j)


    def marshal(self, data):
        return bytes(json.dumps(data) + '\n', 'ascii')


async def test():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM, socket.IPPROTO_IP)
    s.connect(('127.0.0.1', 8888))
    s.setblocking(False)

    loop = asyncio.get_event_loop()
    c = DlvConnection(loop, s)
    response = await c.request({
        "method": "RPCServer.SetApiVersion",
        "params": [ { "APIVersion": 2 } ]
    })
    print(response)


if __name__ == '__main__':
    asyncio.get_event_loop().create_task(test())
    asyncio.get_event_loop().run_forever()
