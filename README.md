# KNoT-MQTT

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/9140aa8c06934071ad6e3cf3b1b148ff)](https://www.codacy.com/manual/joaoaneto/knot-babeltower?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=CESARBR/knot-mqtt&amp;utm_campaign=Badge_Grade)
![Build and test](https://github.com/cesarbr/knot-mqtt/workflows/Build%20and%20test/badge.svg)

## Table of Contents

- [Basic Installation and Usage](#basic-installation-and-usage)
  - [Requirements](#requirements)
  - [Configuration](#configuration)
  - [Compiling and Running](#compiling-and-running)
- [Docker Installation and Usage](#docker-installation-and-usage)
  - [Requirements](#requirements-1)
  - [Building and Running](#building-and-running)
---

## Basic Installation and Usage

### Requirements

- This project requires Go 1.8 or higher.

### Configuration

- The `internal/configuration/mqtt_setup.yaml` file contains the general specifications of the application. Parameters:
  - `mqttBroker`: URL to connect with the Mosquitto broker.
  - `mqttClientID`: Client data responsible for the connection.
  - `topic`: Generic topic (will be subscribed when starting to read sensor data).
  - `mqttQoS`: Quality of Service.
  - `mqttUser`: User of the MQTT service.
  - `mqttPass`: Password of the MQTT service.
    - **Note: If the MQTT service does not require a connection user, we can remove the `mqttUser` and `mqttPass` fields.**

- The `internal/configuration/mqtt_device_config.yaml` file contains JSON structure. This structure provides the capability to customize sensor configuration, allowing for the definition of specific operational and behavioral parameters. Additionally, it enables the assignment of a unique topic to the sensor, indicating the sensor's location or purpose within the network or system at hand. This results in greater flexibility in organizing and managing sensors, ensuring that each sensor can be appropriately configured and identified through its associated topic.

Example `mqtt_device_config.yaml`:
```yaml
  config:
  - 1:
    Topic: topic
    Value: data.0.valor
    Timestamp: data.0.timestamp
  - 2:
    Topic: topic2
    Value: data.0.valor
    Timestamp: data.0.timestamp
```

- To configure the KNoT device, recognized as a KNoT device, spread data as KNoT sensors. Configure each sensor with its ID, name, unit type, and value type. Go to `internal/configuration/device_config.yaml`, where there is a map of devices, and each device has multiple sensors. If no `devices_config.yaml` file exists, create one using the provided template. A new ID and token will be generated upon registration with KNoT cloud.

Refer to the [documentation](https://knot-devel.cesar.org.br/doc/thing/unit-type-value.html) for possible parameter combinations.

Example `device_config.yaml`:

```yaml
0d7cd9d221385e1f:
  id: 0d7cd9d221385e1f
  token: ""
  name: "name"
  config:
  - sensorId: 1
    schema:
      valueType: 2
      unit: 1
      typeId: 65296 
      name: name
    event:
      change: true
      timeSec: 0
      lowerThreshold: null 
      upperThreshold: null
  state: new
  data: []
  error: ""
```

### Compiling and running

Enter the following command:
```shell
go run cmd/main.go
```
Or, with environment variables:

```shell
DEVICE_CONFIG=example/config.yaml KNOT_CONFIG=example/config.yaml MQTT_SETUP=example/config.yaml MQTT_DEVICE_CONFIG=example/config.yaml go run cmd/main.go
```

## Docker installation and usage

### Requirements

- Install docker engine (<https://docs.docker.com/install/>)

### Building and Running

A container is specified at `docker/Dockerfile`. To use it, execute the following steps:

01. Build the image:

    ```bash
    docker build . --file docker/Dockerfile --tag knot-mqtt
    ```

02. Create a file containing the configuration as environment variables(as specified in  Dockerfile).

03. Run the container:

    ```bash
    docker run knot-mqtt
    ```
