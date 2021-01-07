#!/usr/bin/env python3

import jsonstreamer


class JsonParser:
    def __init__(self):
        self.ready_obj = None
        self.staging_obj = None
        self.init_streamer()


    def parse(self, data):
        self.streamer.consume(data)
        obj, self.ready_obj = self.ready_obj, None
        if obj is not None:
            self.streamer.close()
            self.init_streamer()
        return obj


    def on_start(self):
        self.staging_obj = {}


    def on_end(self):
        self.ready_obj = self.staging_obj


    def on_pair(self, pair):
        self.staging_obj[pair[0]] = pair[1]


    def init_streamer(self):
        self.streamer = jsonstreamer.ObjectStreamer()
        self.streamer.add_listener('object_stream_start', self.on_start)
        self.streamer.add_listener('object_stream_end', self.on_end)
        self.streamer.add_listener('pair', self.on_pair)


if __name__ == '__main__':
    print('Testing...')

    p = JsonParser()
    data = '{"a": 1, "b": [{"x": 5, "y": 10}, {"x": 15, "y": 20}]}'
    canonical = {"a": 1, "b": [{"x": 5, "y": 10}, {"x": 15, "y": 20}]}
    obj = p.parse(data)
    assert obj == canonical

    obj = p.parse(data[:10])
    assert not obj
    obj = p.parse(data[10:])
    assert obj == canonical

    p = JsonParser()
    data = '{}'
    canonical = {}
    obj = p.parse(data)
    assert obj == canonical

    data = '{"a": 1}'
    canonical = {"a": 1}
    obj = p.parse(data)
    assert obj == canonical

    print('OK')
