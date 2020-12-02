

# Install
Make sure that you have a working [Golang installation](https://golang.org/doc/install) and that your [GOPATH](https://golang.org/doc/code.html#GOPATH) is set.
Then compile drng-exporter via:
```
git clone https://github.com/Dr-Electron/drng-exporter
cd drng-exporter
go build
```
Or download the precompiled binary from [Releases](https://github.com/Dr-Electron/drng-exporter/releases).  
You can run it with `./drng-exporter`. Get a list of possible arguments with `./drng-exporter -h`
# Run as service
Create the service file:
```
nano /etc/systemd/system/drng-exporter.service
```
Add the content:
```
[Unit]
Description=drng-exporter service
After=network.target

[Service]
Restart=on-failure
ExecStart=/path-to-drng-exporter/drng-exporter

[Install]
WantedBy=multi-user.target
```
Add arguments to `ExecStart` if you need a non default setup.  
Now tell systemd you added a new service, enable and start it.
```
sudo systemctl daemon-reload
sudo systemctl enable drng-exporter.service
sudo systemctl start drng-exporter.service
```
# Prometheus and Grafana
This section will describe how to set up a Grafana dashboard for dRNG.
## Prometheus
Add `drng_metrics` as job in the `prometheus.yml` file.
You can also add other committee members to the job.
The following could be an example config:
```
global:
  scrape_interval: 5s
scrape_configs:
  - job_name: goshimmer_local
    static_configs:
    - targets: ['localhost:9311']
  - job_name: drng_metrics # Metric job name
    static_configs:
    - targets: ['member1-url:1236', 'member1-url:2112'] # dRNG and drng-exporter metric ip:port
      labels:
        instance: 'member1-name' # the name you want to show in the dashboard for this member
    - targets: ['member2-url:1236', 'member2-url:2112']
      labels:
        instance: 'member2-name'
```
## Grafana
### Install Worldmap Panel Plugin
Binary:
```
grafana-cli plugins install grafana-worldmap-panel
```
Docker-Compose:
Add the following under `environment`:
```
- GF_INSTALL_PLUGINS=grafana-worldmap-panel
```
Restart Grafana.
### Add Dashboard
You can use the Grafana dashboard under `grafana/dRNG-dashboard.json` in this repo.
Download it with:
```
wget https://raw.githubusercontent.com/Dr-Electron/drng-exporter/master/grafana/dRNG-dashboard.json
```
Copy the file into `grafana/dashboards`
