import os
import time
import uuid
from bili_ticket_gt_python import ClickPy
from multiprocessing import Manager, Process, Queue


def geetest_worker(input_queue: Queue, output_dict):
    while True:
        task_id, gt, challenge = input_queue.get()
        click = ClickPy()
        try:
            validate = click.simple_match_retry(gt, challenge)
            output_dict[task_id] = (validate, validate + "|jordan", None)
        except Exception as e:
            output_dict[task_id] = (None, None, str(e))


WORKER_COUNT = int(os.environ.get("GEETEST_WORKER_COUNT", 10))


class GeetestPool:
    def __init__(self):
        self.manager = Manager()
        self.output_dict = self.manager.dict()

        self.input_queue: Queue = Queue()
        self.processes = []

        for _ in range(WORKER_COUNT):
            p = Process(
                target=geetest_worker, args=(self.input_queue, self.output_dict)
            )
            p.daemon = True
            p.start()
            self.processes.append(p)

    def submit(self, gt: str, challenge: str):
        task_id = str(uuid.uuid4())
        self.input_queue.put((task_id, gt, challenge))
        return task_id

    def get_result(self, task_id: str, timeout: float = 10.0):
        start = time.time()
        while time.time() - start < timeout:
            if task_id in self.output_dict:
                return self.output_dict.pop(task_id)
            time.sleep(0.05)
        raise TimeoutError("Geetest 验证超时")
