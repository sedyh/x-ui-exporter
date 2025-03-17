<p align="center">
  <img src="https://raw.githubusercontent.com/hteppl/3x-ui-exporter/main/.github/images/logo.png" alt="logo">
</p>

# 3X-UI Metrics Exporter

[![GitHub Release](https://img.shields.io/github/v/release/hteppl/3x-ui-exporter?style=flat&color=blue)](https://github.com/hteppl/3x-ui-exporter/releases/latest)
[![DockerHub](https://img.shields.io/badge/DockerHub-hteppl%2Fx--ui--exporter-blue)](https://hub.docker.com/r/hteppl/x-ui-exporter/)
[![GitHub License](https://img.shields.io/github/license/kutovoys/marzban-exporter?color=greeen)](https://github.com/kutovoys/marzban-exporter/blob/main/LICENSE)

3X-UI Metrics Exporter is an application designed to collect and export metrics from
the [3X-UI Web Panel](https://github.com/MHSanaei/3x-ui). This exporter enables monitoring of various aspects
of the VPN service, such as node status, traffic, system metrics, and user information, making the data available for
the Prometheus monitoring system.

## Features

- **Online Monitoring**: Tracks the number of online users per 3X-UI instance.
- **Traffic Metrics**: Tracks total uploaded/downloaded bytes per client or inbound.
- **3X-UI Monitoring**: Provides XRay version and another metrics from 3X-UI.
- **Version and Start Time Information**: Includes core version information and whether the core service has started
  successfully.
- **Configurable via Environment Variables and Command-line Arguments**: Allows customization and configuration through
  both
  environment variables and command-line arguments, making it easy to adjust settings.
- **Support for Multiple Architectures**: Docker images are available for multiple architectures, including AMD64 and
  ARM64,
  ensuring compatibility across various deployment environments.
- **Optional BasicAuth Protection**: Provides the option to secure the metrics endpoint with BasicAuth, adding an
  additional layer of security.
- **Integration with Prometheus**: Designed to integrate seamlessly with Prometheus, facilitating easy setup and
  configuration for monitoring 3X-UI panel.
- **Simplifies VPN Monitoring**: By providing a wide range of metrics, it simplifies the process of monitoring and
  managing
  VPN services, enhancing visibility into system performance and user activity.

## Metrics

Below is a table of the metrics provided by 3X-UI Metrics Exporter.

### Users

Users metrics, such as online:

| Name                      | Description                   |
|---------------------------|-------------------------------|
| `x_ui_total_online_users` | Total number of online users. |

### Clients

Clients metrics (params: `id`, `email`):

| Name                     | Description                        |
|--------------------------|------------------------------------|
| `x_ui_client_up_bytes`   | Total uploaded bytes per client.   |
| `x_ui_client_down_bytes` | Total downloaded bytes per client. |

### Inbounds

Inbounds metrics (params: `id`, `remark`):

| Name                      | Description                         |
|---------------------------|-------------------------------------|
| `x_ui_inbound_up_bytes`   | Total uploaded bytes per inbound.   |
| `x_ui_inbound_down_bytes` | Total downloaded bytes per inbound. |

### System

System metrics (`version` param for `x_ui_xray_version`):

| Name                 | Description                 |
|----------------------|-----------------------------|
| `x_ui_xray_version`  | XRay version used by 3X-UI. |
| `x_ui_panel_threads` | 3X-UI panel threads.        |
| `x_ui_panel_memory`  | 3X-UI panel memory usage.   |
| `x_ui_panel_uptime`  | 3X-UI panel uptime.         |

## Configuration

3X-UI Metrics Exporter can be configured using environment variables, command-line arguments, or a YAML configuration
file. These are alternative methods of configuration, and you should choose one approach for your deployment.

Below is a table of configuration options:

| Variable Name       | Command-Line Argument | Required | Default Value              | Description                                                            |
|---------------------|-----------------------|----------|----------------------------|------------------------------------------------------------------------|
| `CONFIG_FILE`       | `--config-file`       | No       | N/A                        | Path to YAML configuration file. When provided, CLI flags are ignored. |
| `PANEL_BASE_URL`    | `--panel-base-url`    | Yes      | `https://<your-panel-url>` | URL of the 3X-UI management panel.                                     |
| `PANEL_USERNAME`    | `--panel-username`    | Yes      | `<your-panel-username>`    | Username for the 3X-UI panel.                                          |
| `PANEL_PASSWORD`    | `--panel-password`    | Yes      | `<your-panel-password>`    | Password for the 3X-UI panel.                                          |
| `METRICS_PORT`      | `--metrics-port`      | No       | `9090`                     | Port for the metrics server.                                           |
| `METRICS_PROTECTED` | `--metrics-protected` | No       | `false`                    | Enable BasicAuth protection for metrics endpoint.                      |
| `METRICS_USERNAME`  | `--metrics-username`  | No       | `metricsUser`              | Username for BasicAuth, effective if `METRICS_PROTECTED` is `true`.    |
| `METRICS_PASSWORD`  | `--metrics-password`  | No       | `MetricsVeryHardPassword`  | Password for BasicAuth, effective if `METRICS_PROTECTED` is `true`.    |
| `UPDATE_INTERVAL`   | `--update-interval`   | No       | `30`                       | Interval (in seconds) for metrics update.                              |
| `TIMEZONE`          | `--timezone`          | No       | `UTC`                      | Timezone for correct time display.                                     |

### YAML Configuration

You can use a YAML configuration file to configure the exporter by providing the `--config-file` flag. When using a
configuration file, all settings are read from the file and any command-line arguments are ignored. The YAML
configuration file should contain the same parameters as the command-line arguments, but in YAML format.

A sample configuration file `config-example.yaml` is provided with the project, which you can use as a template for your
own configuration. The structure of the YAML file matches the command-line arguments.

Example YAML configuration:

```yaml
# 3X-UI panel connection details
panel-base-url: "https://your-panel-url"
panel-username: "your-panel-username"
panel-password: "your-panel-password"

# General settings
update-interval: 30
timezone: "UTC"

# Metrics server configuration
metrics-port: 9090
metrics-protected: false
metrics-username: "metricsUser"
metrics-password: "MetricsVeryHardPassword"
```

**Note:** When using a configuration file with the `--config-file` flag, all settings are taken from the configuration
file, and any other command-line arguments are ignored.

## Usage

### CLI

```bash
/x-ui-exporter --panel-base-url=<your-panel-url> --panel-username=<your-panel-username> --panel-password=<your-panel-password>
```

Or using a YAML configuration file:

```bash
/x-ui-exporter --config-file=config.yaml
```

The configuration file approach and the command-line arguments approach are alternative methods and cannot be combined.
When using a configuration file, any other command-line arguments are ignored.

### Docker

```bash
docker run -d \
  -e PANEL_BASE_URL=<your-panel-url> \
  -e PANEL_USERNAME=<your-panel-username> \
  -e PANEL_PASSWORD=<your-panel-password> \
  -p 9090:9090 \
  hteppl/x-ui-exporter
```

### Docker Compose

```bash
version: "3"
services:
  x-ui-exporter:
    image: hteppl/x-ui-exporter
    environment:
      - PANEL_BASE_URL=<your-panel-url>
      - PANEL_USERNAME=<your-panel-username>
      - PANEL_PASSWORD=<your-panel-password>
    ports:
      - "9090:9090"
```

### Integration with Prometheus

To collect metrics with Prometheus, add the exporter to your prometheus.yml configuration file:

```yaml
scrape_configs:
  - job_name: "x-ui_exporter"
    static_configs:
      - targets: [ "<exporter-ip>:9090" ]
```

Ensure to replace `<your-panel-url>`, `<your-panel-username>`, `<your-panel-password>`, and `<exporter-ip>`
with your actual information.

## TODO

- ⏳ Implement more useful metrics.
- ⏳ Create public docker image.

## Contribute

Contributions to 3X-UI Metrics Exporter are warmly welcomed. Whether it's bug fixes, new features, or documentation
improvements, your input helps make this project better. Here's a quick guide to contributing:

1. **Fork & Branch**: Fork this repository and create a branch for your work.
2. **Implement Changes**: Work on your feature or fix, keeping code clean and well-documented.
3. **Test**: Ensure your changes maintain or improve current functionality, adding tests for new features.
4. **Commit & PR**: Commit your changes with clear messages, then open a pull request detailing your work.
5. **Feedback**: Be prepared to engage with feedback and further refine your contribution.

Happy contributing! If you're new to this, GitHub's guide
on [Creating a pull request](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request)
is an excellent resource.
