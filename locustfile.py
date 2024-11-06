import yaml
import logging
from locust import HttpUser, task, between
import os
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(level=logging.INFO)

class WebsiteUser(HttpUser):
    wait_time = between(1, 3)

    def on_start(self):
        """Called when a user starts. Load tasks from the YAML file."""
        self.dynamic_tasks = []
        self.username = os.getenv("USERNAME")
        self.password = os.getenv("PASSWORD")
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
            method = task_definition.get("method")
            endpoint = task_definition.get("endpoint")
            params = task_definition.get("params", {})
            body = task_definition.get("body", {})

            if method == "GET":
                def create_get_task(endpoint, params):
                    def get_task(self):
                        response = self.client.get(endpoint, params=params, auth=(self.username, self.password))
                        logging.info(f"Performed GET request to {endpoint} with params {params}")
                        if response.status_code == 403:
                            print("403 Forbidden: Access Denied")
                        else:
                            print(f"Query response: {response.status_code}")
                    return get_task

                self.dynamic_tasks.append(create_get_task(endpoint, params))
            else:
                logging.error(f"Unknown method: {method}")

    @task
    def perform_dynamic_tasks(self):
        """This will trigger all dynamically defined tasks."""
        if not self.dynamic_tasks:
            logging.error("No tasks defined!")
            return

        for task_method in self.dynamic_tasks:
            task_method(self)
