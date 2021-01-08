#!/usr/bin/env python3

import jsonstreamer


class JsonParser:
    def __init__(self):
        self.ready_objects = []
        self.staging_obj = None
        self.init_streamer()


    def parse(self, data):
        self.streamer.consume(data)
        objects = self.ready_objects
        self.ready_objects = []
        return objects


    def close(self):
        self.streamer.close()


    def on_element(self, element):
        self.ready_objects.append(element)


    def init_streamer(self):
        self.streamer = jsonstreamer.ObjectStreamer()
        self.streamer.consume('{}')
        self.streamer.add_listener('element', self.on_element)


if __name__ == '__main__':
    print('Testing...')

    p = JsonParser()
    data = '{"a": 1, "b": [{"x": 5, "y": 10}, {"x": 15, "y": 20}]}'
    canonical = [{"a": 1, "b": [{"x": 5, "y": 10}, {"x": 15, "y": 20}]}]
    objs = p.parse(data)
    assert objs == canonical

    objs = p.parse(data[:10])
    assert not objs
    objs = p.parse(data[10:])
    assert objs == canonical

    p = JsonParser()
    data = '{}'
    canonical = [{}]
    objs = p.parse(data)
    assert objs == canonical

    data = '{"a": 1}'
    canonical = [{"a": 1}]
    objs = p.parse(data)
    assert objs == canonical

    p = JsonParser()
    data = '{"a": 1} {"b": 2}'
    canonical = [{"a": 1}, {"b": 2}]
    objs = p.parse(data)
    assert len(objs) == 2
    assert objs == canonical
    p.close()

    p = JsonParser()
    objs = p.parse(data[:3])
    assert len(objs) == 0
    objs = p.parse(data[3:])
    assert objs == canonical

    p = JsonParser()
    data = '[{"a": 1}, {"b": 2}] [1,2,"kek"] {"x": "y"}'
    objs = p.parse(data)
    assert len(objs) == 3
    assert objs[0] == [{"a": 1}, {"b": 2}]
    assert objs[1] == [1, 2, "kek"]
    assert objs[2] == {"x": "y"}

    print('OK')
