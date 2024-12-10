import logging
import os
import sys
import time

import clickhouse_driver.errors
import yaml
from clickhouse_driver import connect
from dotenv import load_dotenv
from locust import User
from locust import task, between

load_dotenv()

logging.basicConfig(level=logging.INFO)


class ClickHouseClient:
    def __init__(self, environment):
        self.environment = environment
        self.connection = None
        self.cursor = None
        self.host = None
        self.port = None
        self.user = None
        self.password = None

    def connect(self, host="localhost", port=8123, user="anonymous", password="anonymous@"):
        self.host = host
        self.port = port
        self.user = user
        self.password = password

        self.connection = connect(host=self.host, port=self.port, user=self.user, password=self.password, secure=True)
        self.cursor = self.connection.cursor()

    def execute(self, name, query):

        start_perf_counter = time.perf_counter()
        exception = None
        response = []
        try:
            response = self.cursor.execute(query)
            response_time = (time.perf_counter() - start_perf_counter) * 1000
        except clickhouse_driver.errors.Error as e:
            exception = e
            response_time = (time.perf_counter() - start_perf_counter) * 1000

        if exception:
            self.environment.events.request.fire(
                request_type="Query execution",
                name=name,
                exception=exception,
                response_time=response_time,
                response_length=len(str(exception)),
            )
        else:
            self.environment.events.request.fire(
                request_type="Query execution",
                name=name,
                response_time=response_time,
                response_length=sys.getsizeof(response),
            )

    def disconnect(self):
        self.connection.close()


class ClickHouseUser(User):
    abstract = True

    def __init__(self, environment):
        super().__init__(environment)
        self.client = ClickHouseClient(environment=environment)


class WmtUser(ClickHouseUser):
    wait_time = between(1, 3)

    def __init__(self, environment):
        super().__init__(environment)
        self.dynamic_tasks = {}

    def on_start(self):
        """Called when a user starts. Load tasks from the YAML file."""
        host = os.getenv("HOST")
        port = int(os.getenv("PORT"))
        username = os.getenv("USERNAME")
        password = os.getenv("PASSWORD")
        self.client.connect(host, port, username, password)

        self.load_tasks("tasks.yaml")

    def load_tasks(self, filename):
        """Load tasks from the YAML file."""
        try:
            with open(filename, 'r') as file:
                tasks = yaml.safe_load(file)
                self.add_dynamic_tasks(tasks)
        except Exception as e:
            logging.error(f"Error loading tasks: {e}")

    def add_dynamic_tasks(self, tasks):
        """Convert YAML task definitions to task methods."""
        for task_definition in tasks:
            task_name = task_definition.get("name")
            query = task_definition.get("query")
            self.dynamic_tasks[task_name] = query

    @task
    def perform_dynamic_tasks(self):
        """This will trigger all dynamically defined tasks."""
        if not self.dynamic_tasks:
            logging.error("No tasks defined!")
            return

        for name, query in self.dynamic_tasks.items():
            self.client.execute(name, query)

    def on_stop(self):
        self.client.disconnect()
