from signal import SIGINT, SIGTERM
import asyncio
from asyncio.subprocess import PIPE
from asyncio import create_subprocess_exec


async def _read_stream(stream, callback):
    while True:
        line = await stream.readline()
        if line:
            callback(line)
        else:
            break


class AsyncSubprocess(object):
    @staticmethod
    async def create(command):
        loop = asyncio.get_event_loop()
        process = await create_subprocess_exec(*command, stdout=PIPE, stderr=PIPE)
        io_task = asyncio.gather(*(
            loop.create_task(
                _read_stream(getattr(process, attr),
                             lambda x: print(f'background process: {x.decode("UTF8")}', end='')))
            for attr in ['stdout', 'stderr']
        ))
        return AsyncSubprocess(process, io_task)

    def __init__(self, process, io_task):
        self.process = process
        self.io_task = io_task

    async def wait(self):
        await self.io_task
        await self.process.wait()

    def terminate(self):
        self.process.terminate()
