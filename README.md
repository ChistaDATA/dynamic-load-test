## Project Overview
This project uses Locust, a Python-based load testing tool, to perform load testing on web services. The test tasks (like GET and POST requests) are dynamically defined in a YAML file (tasks.yaml) and loaded when the test starts. The project uses environment variables for credentials (such as USERNAME and PASSWORD) and the dotenv library to manage them.

## Dependencies
The project depends on the following Python packages:

1. locust: For load testing.
2. pyyaml: To load and parse the YAML task configuration.
3. python-dotenv: To load environment variables from a .env file.
4. logging: For logging the actions taken during load testing (built-in Python module).

## Setting Up the Environment
#### Step 1: Create a Python Virtual Environment
1. Install Python. Make sure `Python 3.7+` is installed.
2. Create a Virtual Environment:
```sh
python3 -m venv venv
```
3. Activate the Virtual Environment:
- On Linux / MacOS:
```sh
source venv/bin/activate
```
- On Windows:
```sh
.\venv\Scripts\activate
```

4. Install Dependencies:
```sh
pip install -r requirements.txt
```

#### Step 2: Create the .env File
Create a .env file in the root directory of the project to store environment variables for authentication (username and password):

```sh
USERNAME=your_username
PASSWORD=your_password
```

## Create the tasks.yaml File
In the root of the project directory, create the tasks.yaml file to define the tasks.

#### Example tasks.yaml Configuration
```YAML
- name: "get_user_events1"
  method: "GET"
  endpoint: "/"
  params:
    query: "SELECT * FROM analytics.user_events LIMIT 100"

- name: "get_user_events2"
  method: "GET"
  endpoint: "/"
  params:
    query: "SELECT * FROM analytics.user_events LIMIT 200"

- name: "get_user_events3"
  method: "GET"
  endpoint: "/"
  params:
    query: "SELECT * FROM analytics.user_events LIMIT 300"
```
- name: A descriptive name for the task.
- method: HTTP method (GET).
- endpoint: The URL path of the request.
- params: Parameters for the request (used with GET).

You can define as many tasks as needed in the YAML file, with each task specifying the method (GET), endpoint, parameters (for GET).

## Running the Test
### Step 1: Start Locust
Open a terminal and run the following command to start Locust:
```sh
locust -f locustfile.py
```
### Step 2: Open the Web Interface
After Locust starts, open your browser and navigate to the Locust Web UI at:
```sh
http://localhost:8089
```

### Step 3: Configure the Test Settings
 On the main page of the Locust Web UI, you'll need to configure the following settings:

- **Host**: In the Host field, provide the base URL for the application you are testing (e.g., http://example.com). This is the URL that Locust will target for all the requests.

- **Number of users to simulate**: Enter the total number of virtual users you want to simulate in the test. For example, if you want to simulate 1000 users, enter 1000.

- **Ramp up(users started / second)**: Enter the rate at which new users will be spawned per second. For example, a hatch rate of 10 means that 10 new users will be created every second.


